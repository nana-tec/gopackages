package linkvaluer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"path"
	"strings"
	"time"
)

// Client defines the interface for LinkValuer operations

type Client interface {
	Login() error
	Refresh() error
	CreateValuation(req *CreateRequest) (*CreateValuationPayload, error)
	ViewAssessments() (*AssessmentsPayload, error)
	DownloadReport(bookingNo string) ([]byte, string, error)
	GetToken() string
	IsTokenValid() bool
	ViewAPIRequests() (*ViewAPIRequestsResponse, error)
}

type client struct {
	config     *Config
	httpClient *http.Client
	endpoint   string
	tokens     *TTLCache[string, string]
}

const defaultRequestTimeout = 60 * time.Second

func defaultTransport() *http.Transport {
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   20 * time.Second,
			KeepAlive: 40 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
	}
}

func NewClient(cfg *Config) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, &ClientError{Type: InternalError, Code: ErrInvalidConfig, Message: err.Error(), Operation: "NewClient"}
	}

	hc := cfg.NewHTTPClient()
	if hc == nil {
		hc = &http.Client{}
	}
	if hc.Transport == nil {
		hc.Transport = defaultTransport()
	}
	if hc.Timeout == 0 {
		hc.Timeout = defaultRequestTimeout
	}

	return &client{
		config:     cfg,
		httpClient: hc,
		endpoint:   strings.TrimRight(cfg.GetEndpoint(), "/"),
		tokens:     NewTTL[string, string](cfg.TokenTTL),
	}, nil
}

func (c *client) debugLog(format string, args ...any) {
	if c.config.Debug {
		log.Printf("[LinkValuer] "+format, args...)
	}
}

// token helpers
func (c *client) setAccessToken(tok string, ttl time.Duration)  { c.tokens.Set("lv_access", tok, ttl) }
func (c *client) setRefreshToken(tok string, ttl time.Duration) { c.tokens.Set("lv_refresh", tok, ttl) }
func (c *client) accessToken() (string, bool)                   { return c.tokens.Get("lv_access") }
func (c *client) refreshToken() (string, bool)                  { return c.tokens.Get("lv_refresh") }

func (c *client) IsTokenValid() bool { _, ok := c.accessToken(); return ok }
func (c *client) GetToken() string   { t, _ := c.accessToken(); return t }

// extractTokenPair tries multiple shapes
func extractTokenPair(body []byte) (access, refresh string) {
	// Try direct struct
	var s struct {
		Success      bool   `json:"success"`
		Message      string `json:"message"`
		Token        string `json:"token"`
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		Data         struct {
			Token        string `json:"token"`
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(body, &s)
	access = firstNonEmpty(s.AccessToken, s.Token, s.Data.AccessToken, s.Data.Token)
	refresh = firstNonEmpty(s.RefreshToken, s.Data.RefreshToken)
	if access != "" || refresh != "" {
		return
	}
	// Fallback to generic map
	var m map[string]any
	if err := json.Unmarshal(body, &m); err == nil {
		access = getString(m, "access_token", "token")
		if d, ok := m["data"].(map[string]any); ok {
			if access == "" {
				access = getString(d, "access_token", "token")
			}
			refresh = getString(d, "refresh_token")
		}
		if refresh == "" {
			refresh = getString(m, "refresh_token")
		}
	}
	return
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if s != "" {
			return s
		}
	}
	return ""
}

func getString(m map[string]any, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func (c *client) ensureAccessToken() error {
	if _, ok := c.accessToken(); ok {
		return nil
	}
	c.debugLog("no access token cached; logging in")
	return c.Login()
}

// isTimeoutErr reports whether err is a network or context timeout error
func isTimeoutErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return true
	}
	// Some transports may wrap timeout messages; fallback to substring check
	if strings.Contains(err.Error(), "timeout") {
		return true
	}
	return false
}

func (c *client) requestTimeout() time.Duration {
	if c.httpClient != nil && c.httpClient.Timeout > 0 {
		return c.httpClient.Timeout
	}
	return defaultRequestTimeout
}

func (c *client) Login() error {
	payload, err := json.Marshal(c.config.Credentials)
	if err != nil {
		return newInternalError("Login", ErrMarshalRequest, err)
	}
	url := c.endpoint + "/get-token"

	retries := 0
	if c.config != nil {
		retries = c.config.Retries
	}

	var resp *http.Response
	var body []byte
	for attempt := 0; attempt <= retries; attempt++ {
		ctx, cancel := context.WithTimeout(c.config.Context, c.requestTimeout())
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
		if err != nil {
			cancel()
			return newInternalError("Login", ErrCreateRequest, err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			if isTimeoutErr(err) && attempt < retries {
				c.debugLog("Login attempt %d timed out; retrying", attempt+1)
				continue
			}
			return newExternalError("Login", ErrHTTPRequest, err.Error())
		}
		// success - read body and break
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return newInternalError("Login", ErrReadResponse, err)
		}
		break
	}
	c.debugLog("login status=%d body=%s", resp.StatusCode, string(body))
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return &ClientError{Type: ExternalError, Code: ErrLoginFailed, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), Operation: "Login", HTTPStatus: resp.StatusCode}
	}
	access, refresh := extractTokenPair(body)
	if access == "" {
		return newExternalError("Login", ErrInvalidCredentials, "missing access token in response")
	}
	c.setAccessToken(access, c.config.TokenTTL)
	if refresh != "" {
		c.setRefreshToken(refresh, 30*24*time.Hour)
	}
	return nil
}

func (c *client) Refresh() error {
	refresh, ok := c.refreshToken()
	if !ok || refresh == "" {
		return newExternalError("Refresh", ErrTokenRefresh, "no refresh token cached")
	}
	url := c.endpoint + "/refresh-token"

	retries := 0
	if c.config != nil {
		retries = c.config.Retries
	}

	var resp *http.Response
	var body []byte
	for attempt := 0; attempt <= retries; attempt++ {
		ctx, cancel := context.WithTimeout(c.config.Context, c.requestTimeout())
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			cancel()
			return newInternalError("Refresh", ErrCreateRequest, err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", refresh))

		resp, err = c.httpClient.Do(req)
		if err != nil {
			if isTimeoutErr(err) && attempt < retries {
				c.debugLog("Refresh attempt %d timed out; retrying", attempt+1)
				continue
			}
			return newExternalError("Refresh", ErrHTTPRequest, err.Error())
		}
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return newInternalError("Refresh", ErrReadResponse, err)
		}
		break
	}
	c.debugLog("refresh status=%d body=%s", resp.StatusCode, string(body))
	if resp.StatusCode != http.StatusOK {
		return &ClientError{Type: ExternalError, Code: ErrTokenRefresh, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), Operation: "Refresh", HTTPStatus: resp.StatusCode}
	}
	access, newRefresh := extractTokenPair(body)
	if access == "" {
		return newExternalError("Refresh", ErrTokenRefresh, "missing access token in response")
	}
	c.setAccessToken(access, c.config.TokenTTL)
	if newRefresh != "" {
		c.setRefreshToken(newRefresh, 30*24*time.Hour)
	}
	return nil
}

func (c *client) authJSON(method, endpoint string, payload []byte) (*http.Response, []byte, error) {
	if err := c.ensureAccessToken(); err != nil {
		return nil, nil, err
	}
	url := c.endpoint + ensureLeadingSlash(endpoint)

	retries := 0
	if c.config != nil {
		retries = c.config.Retries
	}

	var resp *http.Response
	var body []byte
	for attempt := 0; attempt <= retries; attempt++ {
		ctx, cancel := context.WithTimeout(c.config.Context, c.requestTimeout())
		req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
		if err != nil {
			cancel()
			return nil, nil, newInternalError("authJSON:createRequest", ErrCreateRequest, err)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.GetToken()))

		resp, err = c.httpClient.Do(req)
		if err != nil {
			if isTimeoutErr(err) && attempt < retries {
				c.debugLog("authJSON attempt %d timed out; retrying", attempt+1)
				continue
			}
			return nil, nil, newExternalError("authJSON:do", ErrHTTPRequest, err.Error())
		}
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return nil, nil, newInternalError("authJSON:read", ErrReadResponse, err)
		}
		if resp.StatusCode == http.StatusUnauthorized {
			_ = resp.Body.Close()
			if err := c.Refresh(); err != nil {
				return nil, nil, err
			}
			// retry once after refreshing token
			ctx2, cancel2 := context.WithTimeout(c.config.Context, c.requestTimeout())
			req2, err := http.NewRequestWithContext(ctx2, method, url, bytes.NewReader(payload))
			if err != nil {
				cancel2()
				return nil, nil, newInternalError("authJSON:createRequest-retry", ErrCreateRequest, err)
			}
			req2.Header.Set("Accept", "application/json")
			req2.Header.Set("Content-Type", "application/json")
			req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.GetToken()))
			resp, err = c.httpClient.Do(req2)
			if err != nil {
				return nil, nil, newExternalError("authJSON:retry", ErrHTTPRequest, err.Error())
			}
			body, err = io.ReadAll(resp.Body)
			if err != nil {
				_ = resp.Body.Close()
				return nil, nil, newInternalError("authJSON:read-retry", ErrReadResponse, err)
			}
		}
		return resp, body, nil
	}
	// if we reach here it means attempts exhausted
	return nil, nil, newExternalError("authJSON:do", ErrHTTPRequest, fmt.Sprintf("request failed after %d attempts", retries+1))
}

func (c *client) DownloadReport(bookingNo string) ([]byte, string, error) {
	if err := c.ensureAccessToken(); err != nil {
		return nil, "", err
	}
	p := path.Join("/download-pdf", bookingNo)
	url := c.endpoint + ensureLeadingSlash(p)

	retries := 0
	if c.config != nil {
		retries = c.config.Retries
	}

	var resp *http.Response
	var body []byte
	for attempt := 0; attempt <= retries; attempt++ {
		ctx, cancel := context.WithTimeout(c.config.Context, c.requestTimeout())
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			cancel()
			return nil, "", newInternalError("DownloadReport", ErrCreateRequest, err)
		}
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.GetToken()))

		resp, err = c.httpClient.Do(req)
		if err != nil {
			if isTimeoutErr(err) && attempt < retries {
				c.debugLog("DownloadReport attempt %d timed out; retrying", attempt+1)
				continue
			}
			return nil, "", newExternalError("DownloadReport", ErrHTTPRequest, err.Error())
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode == http.StatusUnauthorized {
			if err := c.Refresh(); err != nil {
				return nil, "", err
			}
			// retry once after refresh
			ctx2, cancel2 := context.WithTimeout(c.config.Context, c.requestTimeout())
			req2, err := http.NewRequestWithContext(ctx2, http.MethodGet, url, nil)
			if err != nil {
				cancel2()
				return nil, "", newInternalError("DownloadReport", ErrCreateRequest, err)
			}
			req2.Header.Set("Accept", "*/*")
			req2.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.GetToken()))
			resp, err = c.httpClient.Do(req2)
			if err != nil {
				return nil, "", newExternalError("DownloadReport", ErrHTTPRequest, err.Error())
			}
			defer func() { _ = resp.Body.Close() }()
		}
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", newInternalError("DownloadReport", ErrReadResponse, err)
		}
		break
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.Header.Get("Content-Type"), &ClientError{Type: ExternalError, Code: ErrDownloadReport, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), Operation: "DownloadReport", HTTPStatus: resp.StatusCode}
	}
	return body, resp.Header.Get("Content-Type"), nil
}

func ensureLeadingSlash(p string) string {
	if p == "" {
		return ""
	}
	if p[0] == '/' {
		return p
	}
	return "/" + p
}

func (c *client) CreateValuation(reqBody *CreateRequest) (*CreateValuationPayload, error) {
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, newInternalError("CreateValuation", ErrMarshalRequest, err)
	}
	resp, body, err := c.authJSON(http.MethodPost, "/create-api-request", payload)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, &ClientError{Type: ExternalError, Code: ErrCreateValuation, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), Operation: "CreateValuation", HTTPStatus: resp.StatusCode}
	}
	var out CreateValuationPayload
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, newInternalError("CreateValuation", ErrUnmarshalResponse, err)
	}
	return &out, nil
}

func (c *client) ViewAssessments() (*AssessmentsPayload, error) {
	resp, body, err := c.authJSON(http.MethodGet, "/view-assessment", nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, &ClientError{Type: ExternalError, Code: ErrViewAssessments, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), Operation: "ViewAssessments", HTTPStatus: resp.StatusCode}
	}
	var out AssessmentsPayload
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, newInternalError("ViewAssessments", ErrUnmarshalResponse, err)
	}
	return &out, nil
}

func (c *client) ViewAPIRequests() (*ViewAPIRequestsResponse, error) {
	resp, body, err := c.authJSON(http.MethodGet, "/view-api-requests", nil)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, &ClientError{Type: ExternalError, Code: ErrViewAPIRequests, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), Operation: "ViewAPIRequests", HTTPStatus: resp.StatusCode}
	}

	var out ViewAPIRequestsResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, newInternalError("ViewAPIRequests", ErrUnmarshalResponse, err)
	}
	return &out, nil
}

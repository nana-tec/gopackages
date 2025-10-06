package linkvaluer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
}

type client struct {
	config     *Config
	httpClient *http.Client
	endpoint   string
	tokens     *TTLCache[string, string]
}

func NewClient(cfg *Config) (Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, &ClientError{Type: InternalError, Code: ErrInvalidConfig, Message: err.Error(), Operation: "NewClient"}
	}
	return &client{
		config:     cfg,
		httpClient: cfg.NewHTTPClient(),
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

func (c *client) Login() error {
	payload, err := json.Marshal(c.config.Credentials)
	if err != nil {
		return newInternalError("Login", ErrMarshalRequest, err)
	}
	url := c.endpoint + "/get-token"
	req, err := http.NewRequestWithContext(c.config.Context, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return newInternalError("Login", ErrCreateRequest, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return newExternalError("Login", ErrHTTPRequest, err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return newInternalError("Login", ErrReadResponse, err)
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
	} // assume long-lived
	return nil
}

func (c *client) Refresh() error {
	refresh, ok := c.refreshToken()
	if !ok || refresh == "" {
		return newExternalError("Refresh", ErrTokenRefresh, "no refresh token cached")
	}
	url := c.endpoint + "/refresh-token"
	req, err := http.NewRequestWithContext(c.config.Context, http.MethodGet, url, nil)
	if err != nil {
		return newInternalError("Refresh", ErrCreateRequest, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", refresh))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return newExternalError("Refresh", ErrHTTPRequest, err.Error())
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return newInternalError("Refresh", ErrReadResponse, err)
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
	req, err := http.NewRequestWithContext(c.config.Context, method, url, bytes.NewReader(payload))
	if err != nil {
		return nil, nil, newInternalError("authJSON:createRequest", ErrCreateRequest, err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.GetToken()))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, nil, newExternalError("authJSON:do", ErrHTTPRequest, err.Error())
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return nil, nil, newInternalError("authJSON:read", ErrReadResponse, err)
	}
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		if err := c.Refresh(); err != nil {
			return nil, nil, err
		}
		// retry once
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.GetToken()))
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, nil, newExternalError("authJSON:retry", ErrHTTPRequest, err.Error())
		}
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, nil, newInternalError("authJSON:read-retry", ErrReadResponse, err)
		}
	}
	return resp, body, nil
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
	defer resp.Body.Close()
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
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, &ClientError{Type: ExternalError, Code: ErrViewAssessments, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), Operation: "ViewAssessments", HTTPStatus: resp.StatusCode}
	}
	var out AssessmentsPayload
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, newInternalError("ViewAssessments", ErrUnmarshalResponse, err)
	}
	return &out, nil
}

func (c *client) DownloadReport(bookingNo string) ([]byte, string, error) {
	if err := c.ensureAccessToken(); err != nil {
		return nil, "", err
	}
	p := path.Join("/download-pdf", bookingNo)
	url := c.endpoint + ensureLeadingSlash(p)
	req, err := http.NewRequestWithContext(c.config.Context, http.MethodGet, url, nil)
	if err != nil {
		return nil, "", newInternalError("DownloadReport", ErrCreateRequest, err)
	}
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.GetToken()))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", newExternalError("DownloadReport", ErrHTTPRequest, err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		if err := c.Refresh(); err != nil {
			return nil, "", err
		}
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.GetToken()))
		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, "", newExternalError("DownloadReport", ErrHTTPRequest, err.Error())
		}
		defer resp.Body.Close()
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", newInternalError("DownloadReport", ErrReadResponse, err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, resp.Header.Get("Content-Type"), &ClientError{Type: ExternalError, Code: ErrDownloadReport, Message: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)), Operation: "DownloadReport", HTTPStatus: resp.StatusCode}
	}
	return body, resp.Header.Get("Content-Type"), nil
}

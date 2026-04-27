package directline

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client provides methods to interact with Directline API
type Client struct {
	cfg       *Config
	http      *http.Client
	tokens    *TTLCache[string, string]
	bodyTypes *TTLCache[int, []BodyType]
	mu        sync.Mutex
	endpoint  string
	debug     bool
}

// NewClient constructs a new Directline client
func NewClient(cfg *Config) (*Client, error) {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.directline.co.ke/hayabusa"
	}
	if err := cfg.Validate(); err != nil {
		return nil, newInternalError("NewClient", ErrInvalidConfig, err, cfg.Debug)
	}
	hc := cfg.NewHTTPClient()
	if hc == nil {
		hc = &http.Client{Timeout: cfg.Timeout}
	}
	if hc.Timeout == 0 && cfg.Timeout > 0 {
		hc.Timeout = cfg.Timeout
	}
	c := &Client{
		cfg:       cfg,
		http:      hc,
		tokens:    NewTTL[string, string](cfg.TokenTTL),
		bodyTypes: NewTTL[int, []BodyType](1 * time.Hour),
		endpoint:  strings.TrimRight(cfg.BaseURL, "/"),
		debug:     cfg.Debug,
	}
	return c, nil
}

func (c *Client) IsDebug() bool {
	return c.debug
}

func (c *Client) SetDebug(v bool) {
	c.debug = v
}

func (c *Client) debugLog(format string, args ...any) {
	if c.debug {
		fmt.Printf("[Directline DEBUG] "+format+"\n", args...)
	}
}

func (c *Client) GetToken(ctx context.Context) (string, error) {
	// check cache
	if tok, ok := c.tokens.Get("directline_token"); ok {
		return tok, nil
	}
	// single-flight via mutex
	c.mu.Lock()
	defer c.mu.Unlock()
	if tok, ok := c.tokens.Get("directline_token"); ok {
		return tok, nil
	}
	// request new token
	p, err := json.Marshal(c.cfg.Credentials)
	if err != nil {
		return "", newInternalError("GetToken:marshal", ErrMarshalRequest, err, c.IsDebug())
	}
	url := c.endpoint + "/api/token/generate"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(p))
	if err != nil {
		return "", newInternalError("GetToken:create", ErrCreateRequest, err, c.IsDebug())
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", newExternalError("GetToken:do", ErrHTTPRequest, err.Error(), 0, c.IsDebug())
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", newInternalError("GetToken:read", ErrReadResponse, err, c.IsDebug())
	}
	c.debugLog("token status=%d body=%s", resp.StatusCode, string(body))
	if resp.StatusCode != http.StatusOK {
		return "", newAuthError("GetToken", ErrLoginFailed, string(body), c.IsDebug())
	}
	var tr TokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", newInternalError("GetToken:unmarshal", ErrUnmarshalResponse, err, c.IsDebug())
	}
	if tr.AccessToken == "" {
		return "", newAuthError("GetToken", ErrLoginFailed, "missing access token", c.IsDebug())
	}
	// compute ttl
	ttl := c.cfg.TokenTTL
	if tr.ExpiresIn > 0 {
		ttl = time.Duration(tr.ExpiresIn) * time.Second
	}
	c.tokens.Set("directline_token", tr.AccessToken, ttl)
	return tr.AccessToken, nil
}

func (c *Client) ensureToken(ctx context.Context) error {
	_, ok := c.tokens.Get("directline_token")
	if ok {
		return nil
	}
	_, err := c.GetToken(ctx)
	return err
}

func (c *Client) doAuthJSON(ctx context.Context, method, path string, payload []byte, out interface{}) (int, error) {
	if err := c.ensureToken(ctx); err != nil {
		return 0, err
	}
	url := c.endpoint + ensureLeadingSlash(path)
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(payload))
	if err != nil {
		return 0, newInternalError("doAuthJSON:create", ErrCreateRequest, err, c.IsDebug())
	}
	req.Header.Set("Content-Type", "application/json")
	if tok, ok := c.tokens.Get("directline_token"); ok {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tok))
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, newExternalError("doAuthJSON:do", ErrHTTPRequest, err.Error(), 0, c.IsDebug())
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, newInternalError("doAuthJSON:read", ErrReadResponse, err, c.IsDebug())
	}
	c.debugLog("status=%d body=%s", resp.StatusCode, string(b))
	if resp.StatusCode == http.StatusUnauthorized {
		// try refresh
		if _, err := c.GetToken(ctx); err != nil {
			return resp.StatusCode, err
		}
		// retry once
		if tok, ok := c.tokens.Get("directline_token"); ok {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tok))
		}
		resp2, err := c.http.Do(req)
		if err != nil {
			return 0, newExternalError("doAuthJSON:retry", ErrHTTPRequest, err.Error(), 0, c.IsDebug())
		}
		defer func() { _ = resp2.Body.Close() }()
		b, err = io.ReadAll(resp2.Body)
		if err != nil {
			return resp2.StatusCode, newInternalError("doAuthJSON:read-retry", ErrReadResponse, err, c.IsDebug())
		}
		c.debugLog("token status=%d body=%s", resp2.StatusCode, string(b))
		if resp2.StatusCode != http.StatusOK && resp2.StatusCode != http.StatusCreated {
			return resp2.StatusCode, newExternalError("doAuthJSON", ErrHTTPRequest, string(b), resp2.StatusCode, c.IsDebug())
		}
		if out != nil {
			if err := json.Unmarshal(b, out); err != nil {
				return resp2.StatusCode, newInternalError("doAuthJSON:unmarshal-retry", ErrUnmarshalResponse, err, c.IsDebug())
			}
		}
		return resp2.StatusCode, nil
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return resp.StatusCode, newExternalError("doAuthJSON", ErrHTTPRequest, string(b), resp.StatusCode, c.IsDebug())
	}
	if out != nil {
		if err := json.Unmarshal(b, out); err != nil {
			return resp.StatusCode, newInternalError("doAuthJSON:unmarshal", ErrUnmarshalResponse, err, c.IsDebug())
		}
	}
	return resp.StatusCode, nil
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

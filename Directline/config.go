package directline

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// Credentials used to authenticate with Directline
type Credentials struct {
	ClientID       string `json:"clientId"`
	ConsumerSecret string `json:"consumerSecret"`
}

// Config contains settings for the Directline client
type Config struct {
	BaseURL     string
	Credentials Credentials
	HTTPClient  *http.Client
	// TokenTTL is used as a fallback if the token response does not provide expires_in
	TokenTTL time.Duration
	Timeout  time.Duration
	Debug    bool
	Retries  int
	Context  context.Context
}

// Validate basic configuration
func (c *Config) Validate() error {
	if c == nil {
		return errors.New("directline: nil config")
	}
	if c.BaseURL == "" {
		return errors.New("directline: base url required")
	}
	if c.Credentials.ClientID == "" || c.Credentials.ConsumerSecret == "" {
		return errors.New("directline: credentials are required")
	}
	return nil
}

// NewHTTPClient returns either the configured http client or a default one
func (c *Config) NewHTTPClient() *http.Client {
	if c == nil {
		return nil
	}
	return c.HTTPClient
}

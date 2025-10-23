package linkvaluer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

// Environment for LinkValuer
// Kept for parity; service exposes a single public endpoint but allow overrides
type Environment string

const (
	Production Environment = "production"
)

// Credentials holds authentication info for LinkValuer
// The API expects email and password for token generation
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Config contains client configuration
type Config struct {
	Credentials        Credentials
	Environment        Environment
	CustomEndpoint     string
	Timeout            time.Duration
	InsecureSkipVerify bool
	Debug              bool
	Context            context.Context
	TokenTTL           time.Duration // TTL for access token fallback if API doesn't provide expiry
	Retries            int           // Number of retries on timeout (default 2)
}

// Validate verifies minimal config
func (c *Config) Validate() error {
	if c.Credentials.Email == "" || c.Credentials.Password == "" {
		return fmt.Errorf("missing credentials")
	}
	if c.Environment == "" && c.CustomEndpoint == "" {
		// Allow default production
		c.Environment = Production
	}
	if c.Context == nil {
		c.Context = context.Background()
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	if c.TokenTTL == 0 {
		c.TokenTTL = 12 * time.Hour
	}
	if c.Retries == 0 {
		c.Retries = 2
	}
	return nil
}

// GetEndpoint resolves base URL
func (c *Config) GetEndpoint() string {
	if c.CustomEndpoint != "" {
		return c.CustomEndpoint
	}
	switch c.Environment {
	case Production:
		return "https://portal.linksvaluers.com/api"
	default:
		return "https://portal.linksvaluers.com/api"
	}
}

// NewHTTPClient returns an http.Client honoring TLS options
func (c *Config) NewHTTPClient() *http.Client {
	transport := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: c.InsecureSkipVerify}}
	return &http.Client{Timeout: c.Timeout, Transport: transport}
}

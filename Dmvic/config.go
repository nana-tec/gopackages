package dmvic

import (
	"context"
	"fmt"
	"time"
)

// Environment represents the DMVIC environment type (production or UAT).
// It defines which DMVIC API endpoint to use for operations.
type Environment string

const (
	// Production represents the production DMVIC environment
	Production Environment = "production"
	// UAT represents the User Acceptance Testing DMVIC environment
	UAT Environment = "uat"
)

// Credentials holds authentication information for DMVIC API access.
// It contains the username and password required for login operations.
type Credentials struct {
	Username string `json:"username"` // Username for DMVIC authentication
	Password string `json:"password"` // Password for DMVIC authentication
}

// Config contains all configuration needed to create a DMVIC client.
// It includes authentication details, environment settings, timeout configurations,
// and certificate paths for mutual TLS authentication.
type Config struct {
	Credentials        Credentials     // Authentication credentials
	ClientID           string          // Client identifier for API requests
	Environment        Environment     // Target environment (production or UAT)
	CustomEndpoint     string          // Custom endpoint URL (overrides Environment)
	Timeout            time.Duration   // HTTP request timeout
	TokenTTL           time.Duration   // Time to live for authentication tokens
	InsecureSkipVerify bool            // Skip TLS certificate verification
	Debug              bool            // Enable debug logging
	Context            context.Context // Context for HTTP requests
	AuthCertPath       string          // Path to client certificate file
	AuthKeyPath        string          // Path to client private key file
	AuthCaCertPath     string          // Path to CA certificate file
}

// Validate checks if the configuration is complete and valid.
// It ensures all required fields are set and applies default values where appropriate.
// Returns an error if any required configuration is missing or invalid.
func (c *Config) Validate() error {
	if c.Credentials.Username == "" || c.Credentials.Password == "" {
		return fmt.Errorf("missing credentials")
	}
	if c.ClientID == "" {
		return fmt.Errorf("missing ClientID")
	}
	if c.Environment == "" && c.CustomEndpoint == "" {
		return fmt.Errorf("either Environment or CustomEndpoint must be specified")
	}
	if c.Environment != Production && c.Environment != UAT {
		return fmt.Errorf("invalid Environment: %s, must be 'production' or 'uat'", c.Environment)
	}
	if c.AuthCertPath == "" || c.AuthKeyPath == "" {
		return fmt.Errorf("missing authentication certificate or key path")
	}
	if c.AuthCaCertPath == "" {
		return fmt.Errorf("missing authentication CA certificate path")
	}
	if c.Context == nil {
		ctx := context.Background()
		c.Context = ctx
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	return nil
}

// GetEndpoint returns the appropriate API endpoint URL based on configuration.
// If CustomEndpoint is set, it takes precedence over the Environment setting.
// Otherwise, it returns the standard endpoint for the specified environment.
func (c *Config) GetEndpoint() string {
	if c.CustomEndpoint != "" {
		return c.CustomEndpoint
	}
	switch c.Environment {
	case Production:
		return "https://api.dmvic.com/api"
	case UAT:
		return "https://uat-api.dmvic.com/api"
	default:
		return "https://uat-api.dmvic.com/api"
	}
}

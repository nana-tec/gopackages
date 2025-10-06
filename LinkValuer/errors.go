package linkvaluer

import "fmt"

type ErrorType string

const (
	InternalError ErrorType = "InternalError"
	ExternalError ErrorType = "ExternalError"
)

const (
	ErrInvalidConfig      = 1001
	ErrMarshalRequest     = 1002
	ErrCreateRequest      = 1003
	ErrHTTPRequest        = 1004
	ErrReadResponse       = 1005
	ErrUnmarshalResponse  = 1007
	ErrUnauthorized       = 2003
	ErrInvalidCredentials = 2004
	ErrTokenRefresh       = 2005
	ErrLoginFailed        = 2006

	ErrCreateValuation = 3000
	ErrViewAssessments = 3100
	ErrDownloadReport  = 3200
)

type ClientError struct {
	Type       ErrorType `json:"type"`
	Code       int       `json:"code"`
	Message    string    `json:"message"`
	Operation  string    `json:"operation,omitempty"`
	HTTPStatus int       `json:"http_status,omitempty"`
}

func (e *ClientError) Error() string {
	if e.Operation != "" {
		return fmt.Sprintf("linkvaluer %s error %d: %s", e.Operation, e.Code, e.Message)
	}
	return fmt.Sprintf("linkvaluer error %d: %s", e.Code, e.Message)
}

func newInternalError(op string, code int, err error) *ClientError {
	return &ClientError{Type: InternalError, Code: code, Message: err.Error(), Operation: op}
}

func newExternalError(op string, code int, message string) *ClientError {
	return &ClientError{Type: ExternalError, Code: code, Message: message, Operation: op}
}

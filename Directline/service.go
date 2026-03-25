package directline

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func (c *Client) CalculatePremium(ctx context.Context, req *CalculatePremiumRequest) (*CalculatePremiumResponse, error) {
	var out CalculatePremiumResponse
	_, err := c.doAuthJSON(ctx, "POST", "/calculate-premium", mustMarshal(req), &out)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GetBodyTypes(ctx context.Context) ([]BodyType, error) {
	var out []BodyType
	_, err := c.doAuthJSON(ctx, "GET", "/body-types", nil, &out)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) CreateTransaction(ctx context.Context, endpoint string, req *TransactionRequest) (*TransactionResponse, *ValidationErrorResponse, error) {
	var out TransactionResponse
	path := fmt.Sprintf("/api/bima/%s", endpoint)
	status, err := c.doAuthJSON(ctx, http.MethodPost, path, mustMarshal(req), &out)
	if err != nil {
		// If it's an APIError representing a validation (HTTP 400), try to unmarshal validation details
		if apiErr, ok := err.(*APIError); ok && apiErr.HTTPCode == http.StatusBadRequest {
			var vr ValidationErrorResponse
			// try to unmarshal the last body from doAuthJSON? We don't return it — instead caller should inspect apiErr.Message
			_ = json.Unmarshal([]byte(apiErr.Message), &vr)
			return nil, &vr, apiErr
		}
		return nil, nil, err
	}
	if status == http.StatusOK || status == http.StatusCreated {
		return &out, nil, nil
	}
	return nil, nil, newExternalError("CreateTransaction", ErrHTTPRequest, "unexpected status", status)
}

// CalculatePremiumDetailed posts to /api/premium/calculate and returns PremiumResponse on 200 or PremiumErrorResponse on 400
func (c *Client) CalculatePremiumDetailed(ctx context.Context, req *PremiumRequest) (*PremiumResponse, *PremiumErrorResponse, error) {
	var out PremiumResponse
	status, err := c.doAuthJSON(ctx, http.MethodPost, "/api/premium/calculate", mustMarshal(req), &out)
	if err != nil {
		if apiErr, ok := err.(*APIError); ok && apiErr.HTTPCode == http.StatusBadRequest {
			var perr PremiumErrorResponse
			_ = json.Unmarshal([]byte(apiErr.Message), &perr)
			return nil, &perr, apiErr
		}
		return nil, nil, err
	}
	if status == http.StatusOK || status == http.StatusCreated {
		return &out, nil, nil
	}
	return nil, nil, newExternalError("CalculatePremiumDetailed", ErrHTTPRequest, "unexpected status", status)
}

// GetBodyTypesFor retrieves available body types for a vehicleTypeId and caches results for 1 hour
func (c *Client) GetBodyTypesFor(ctx context.Context, vehicleTypeId int) ([]BodyType, error) {
	// check cache
	if bts, ok := c.bodyTypes.Get(vehicleTypeId); ok {
		return bts, nil
	}
	// not cached: request
	path := fmt.Sprintf("/api/premium/bodytypes/%d", vehicleTypeId)
	var resp BodyTypesResponse
	status, err := c.doAuthJSON(ctx, http.MethodGet, path, nil, &resp)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return nil, newExternalError("GetBodyTypesFor", ErrHTTPRequest, fmt.Sprintf("no body types for vehicleTypeId %d", vehicleTypeId), status)
	}
	if status != http.StatusOK {
		return nil, newExternalError("GetBodyTypesFor", ErrHTTPRequest, "unexpected status", status)
	}
	// map response to []BodyType
	out := make([]BodyType, 0, len(resp.BodyTypes))
	for _, b := range resp.BodyTypes {
		out = append(out, BodyType{ID: b.BodyTypeID, Name: b.BodyTypeName})
	}
	// cache for 1 hour
	c.bodyTypes.Set(vehicleTypeId, out, 1*time.Hour)
	return out, nil
}

// Convenience wrappers for common endpoints
func (c *Client) CreateBusComprehensive(ctx context.Context, req *TransactionRequest) (*TransactionResponse, *ValidationErrorResponse, error) {
	return c.CreateTransaction(ctx, EndpointBusComp, req)
}

func (c *Client) CreateBusTPO(ctx context.Context, req *TransactionRequest) (*TransactionResponse, *ValidationErrorResponse, error) {
	return c.CreateTransaction(ctx, EndpointBusTPO, req)
}

func (c *Client) CreatePrivateComprehensive(ctx context.Context, req *TransactionRequest) (*TransactionResponse, *ValidationErrorResponse, error) {
	return c.CreateTransaction(ctx, EndpointPrivateComp, req)
}

func (c *Client) CreatePrivateTPO(ctx context.Context, req *TransactionRequest) (*TransactionResponse, *ValidationErrorResponse, error) {
	return c.CreateTransaction(ctx, EndpointPrivateTPO, req)
}

func mustMarshal(v interface{}) []byte {
	b, _ := jsonMarshal(v)
	return b
}

func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

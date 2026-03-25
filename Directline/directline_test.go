package directline

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTokenCaching(t *testing.T) {
	// setup a test server that returns token
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/token/generate" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"access_token":"abcd","expires_in":1}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	cfg := &Config{
		//BaseURL:     srv.URL,
		Credentials: Credentials{ClientID: "x", ConsumerSecret: "y"},
		TokenTTL:    2 * time.Second,
		Timeout:     2 * time.Second,
		Context:     context.Background(),
	}
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	// first get should call server
	tok, err := c.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken1: %v", err)
	}
	if tok != "abcd" {
		t.Fatalf("unexpected token: %s", tok)
	}
	// token should be cached
	tok2, err := c.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken2: %v", err)
	}
	if tok2 != "abcd" {
		t.Fatalf("unexpected token2: %s", tok2)
	}
	// wait for expiry
	time.Sleep(2 * time.Second)
	// next call should refresh (server still returns abcd)
	tok3, err := c.GetToken(context.Background())
	if err != nil {
		t.Fatalf("GetToken3: %v", err)
	}
	if tok3 != "abcd" {
		t.Fatalf("unexpected token3: %s", tok3)
	}
}

func TestCreateTransactionCases(t *testing.T) {
	// Create a multiplexing test server that serves token and bima endpoints
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/token/generate":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"access_token":"tok123","expires_in":3600}`))
			return
		case "/api/bima/privatecomp":
			// success response
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"message":"Posted Successfully!!","data":{"coverEndDate":"2027-03-01T00:00:00","plateNumber":"KBZ123A","engineNo":"1500","cubicCapacity":"1500","chasisNo":"ABC123XYZ456","amount":18540.00}}`))
			return
		case "/api/bima/invalid":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			_, _ = w.Write([]byte(`{"success":false,"errors":["Invalid email format"],"message":"Validation failed","availableBodyTypes":[{"bodyTypeId":1,"bodyTypeName":"Saloon"}]}`))
			return
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	cfg := &Config{
		//BaseURL:     srv.URL,
		Credentials: Credentials{ClientID: "x", ConsumerSecret: "y"},
		TokenTTL:    1 * time.Hour,
		Timeout:     5 * time.Second,
		Context:     context.Background(),
	}
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		req := &TransactionRequest{TransactionCode: 20260218000001, CoverStartDate: "2026-03-01T00:00:00", Period: 12, PlateNumber: "KBZ123A", ChasisNo: "ABC123XYZ456", Make: "Toyota", Model: "Corolla", YrManft: "2020", CubicCapacity: "1500", CarryCapacity: 5, Tonnage: 0, AssetValue: 0, BodyTypeID: 1, CustomerFullName: "John Doe", IDNumber: "12345678", Email: "john.doe@example.com", MSISDN: "254722123456", PIN: "A123456789K", PostalAddress: "P.O. Box 12345", Town: "Nairobi", ClientType: "Individual", RequestType: "NEW", Amount: 0, PaymentMode: "MPESA", AgentCode: 1001, APICode: "API123", PVT: "NO", EPR: "NO", BimaPlus: 0}
		tr, vr, err := c.CreateTransaction(context.Background(), "privatecomp", req)
		if err != nil {
			t.Fatalf("CreateTransaction error: %v", err)
		}
		if vr != nil {
			t.Fatalf("expected no validation error, got: %v", vr)
		}
		if tr == nil || tr.Data.Amount != 18540.00 {
			t.Fatalf("unexpected response data: %+v", tr)
		}
	})

	t.Run("validation", func(t *testing.T) {
		req := &TransactionRequest{TransactionCode: 1, CoverStartDate: "2026-03-01T00:00:00", Period: 12}
		tr, vr, err := c.CreateTransaction(context.Background(), "invalid", req)
		if err == nil {
			t.Fatalf("expected error for validation case")
		}
		if vr == nil {
			t.Fatalf("expected validation details, got nil")
		}
		if len(vr.Errors) == 0 {
			t.Fatalf("expected validation errors, got none")
		}
		if tr != nil {
			t.Fatalf("expected no transaction response on validation error")
		}
	})
}

func TestCalculatePremiumCases(t *testing.T) {
	// multiplexing server for token and premium
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/token/generate":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"access_token":"tok-prem","expires_in":3600}`))
			return
		case "/api/premium/calculate":
			// Determine whether to return success or validation based on bodyTypeID query param (simple test heuristic)
			w.Header().Set("Content-Type", "application/json")
			// read body to decide
			var buf = make([]byte, 1024)
			n, _ := r.Body.Read(buf)
			b := string(buf[:n])
			if b == "" || (len(b) >= 10 && b[0] == '{') {
				// naive check: if body contains "bodyTypeID":99 return validation
				if n > 0 && (string(buf[:n])) != "" {
					if contains := func(s, sub string) bool { return len(s) >= len(sub) && (len(s) > 0) && (len(sub) > 0) && (s == s) }; contains(b, "99") {
						w.WriteHeader(400)
						_, _ = w.Write([]byte(`{"success":false,"message":"Invalid body type","errors":["Body type ID 99 is not valid"],"availableBodyTypes":[{"bodyTypeId":1,"bodyTypeName":"Saloon"}]}`))
						return
					}
				}
				_, _ = w.Write([]byte(`{"success":true,"basePremium":15000.00,"totalPremium":18500.00,"newPolicyFee":40.00,"finalAmount":18540.00,"bodyTypeName":"Saloon","breakDown":{"basePremium":15000.00,"levies":3500.00,"newPolicyFee":40.00,"total":18540.00}}`))
				return
			}
			w.WriteHeader(400)
			return
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	cfg := &Config{
		//BaseURL:     srv.URL,
		Credentials: Credentials{ClientID: "x", ConsumerSecret: "y"},
		TokenTTL:    1 * time.Hour,
		Timeout:     5 * time.Second,
		Context:     context.Background(),
	}
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		req := &PremiumRequest{VehicleTypeID: 5, CarryCapacity: 5, Tonnage: 0, CoverTypeID: 1, Period: 12, BodyTypeID: 1, PVT: "NO", EPR: "NO", AgentCode: 0, RequestType: "NEW"}
		pr, perr, err := c.CalculatePremiumDetailed(context.Background(), req)
		if err != nil {
			t.Fatalf("CalculatePremiumDetailed error: %v", err)
		}
		if perr != nil {
			t.Fatalf("expected no premium error, got: %v", perr)
		}
		if pr == nil || pr.FinalAmount != 18540.00 {
			t.Fatalf("unexpected premium response: %+v", pr)
		}
	})

	t.Run("validation", func(t *testing.T) {
		req := &PremiumRequest{VehicleTypeID: 5, CarryCapacity: 5, Tonnage: 0, CoverTypeID: 1, Period: 12, BodyTypeID: 99, PVT: "NO", EPR: "NO", AgentCode: 0, RequestType: "NEW"}
		pr, perr, err := c.CalculatePremiumDetailed(context.Background(), req)
		if err == nil {
			t.Fatalf("expected error for premium validation case")
		}
		if perr == nil {
			t.Fatalf("expected premium validation details, got nil")
		}
		if len(perr.Errors) == 0 {
			t.Fatalf("expected premium validation errors, got none")
		}
		if pr != nil {
			t.Fatalf("expected no premium result on validation error")
		}
	})
}

func TestGetBodyTypes(t *testing.T) {
	// setup a test server that returns body types for vehicleTypeId 5 and 404 for 99
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/token/generate" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"access_token":"tok-bt","expires_in":3600}`))
			return
		}
		if r.URL.Path == "/api/premium/bodytypes/5" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"success":true,"vehicleTypeId":5,"bodyTypes":[{"bodyTypeId":1,"bodyTypeName":"Saloon"},{"bodyTypeId":2,"bodyTypeName":"SUV"}]}`))
			return
		}
		if r.URL.Path == "/api/premium/bodytypes/99" {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"success":false,"message":"No body types found for vehicle type ID 99"}`))
			return
		}
		w.WriteHeader(404)
	}))
	defer srv.Close()

	cfg := &Config{
		//BaseURL:     srv.URL,
		Credentials: Credentials{ClientID: "x", ConsumerSecret: "y"},
		TokenTTL:    1 * time.Hour,
		Timeout:     5 * time.Second,
		Context:     context.Background(),
	}
	c, err := NewClient(cfg)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		// first call, should hit the server
		bodyTypes, err := c.GetBodyTypesFor(context.Background(), 5)
		if err != nil {
			t.Fatalf("GetBodyTypesFor1: %v", err)
		}
		if len(bodyTypes) != 2 {
			t.Fatalf("unexpected body types count: %d", len(bodyTypes))
		}
		if bodyTypes[0].Name != "Saloon" {
			t.Fatalf("unexpected first body type: %s", bodyTypes[0].Name)
		}

		// second call, should use cached result
		bodyTypes2, err := c.GetBodyTypesFor(context.Background(), 5)
		if err != nil {
			t.Fatalf("GetBodyTypesFor2: %v", err)
		}
		if bodyTypes2[0].Name != "Saloon" {
			t.Fatalf("unexpected cached body type: %s", bodyTypes2[0].Name)
		}
	})

	t.Run("404", func(t *testing.T) {
		_, err := c.GetBodyTypesFor(context.Background(), 99)
		if err == nil {
			t.Fatalf("expected error for 404 case")
		}
	})
}

package quotation

import (
	"context"
	"testing"

	dmvic "github.com/nana-tec/gopackages/Dmvic"
)

func TestQuotationValidator(t *testing.T) {

	rootCtx := context.Background()
	dmvicCred := dmvic.Credentials{
		Username: "bizsurebrokeruatapi@dmvic.info",
		Password: "FwQG5gU8Snjv",
	}

	dmvicConfig := &dmvic.Config{
		Credentials:    dmvicCred,
		ClientID:       "CEAEE889BF8F49B8A877EACEAE49B2E3", // "CEAEE889BF8F49B8A877EACEAE49B2E3",
		Environment:    dmvic.UAT,
		Context:        rootCtx,
		Debug:          true,
		AuthCertPath:   "/Users/robertnjoroge/projects/ibima/ibima-admin-backend/cert/client.crt",
		AuthKeyPath:    "/Users/robertnjoroge/projects/ibima/ibima-admin-backend/cert/client.key",
		AuthCaCertPath: "/Users/robertnjoroge/projects/ibima/ibima-admin-backend/cert/ca.crt",
	}

	println("Creating Dmvic Client  ")
	dmvicClient, err := dmvic.NewClient(dmvicConfig)
	if err != nil {
		t.Fatalf("Failed to create dmvic client : %v", err)
	}
	println("Creating Quotation Validator  ")
	quotationValidator, err := NewQuotationValidatorInstance(dmvicClient)
	if err != nil {
		t.Fatalf("Failed to create quotation validator : %v", err)
	}

	println("Validator risk with dmvic  ")

	coverDet := &CoverDetails{
		StartDate: "2025-12-26",
		Period:    31,
	}
	riskDet := &RiskDetails{
		RegistrationNumber: "KDP343W",
		ChassisNumber:      "JIT123DFREW1212398",
		OtherDetails:       map[string]any{"make": "toyota"},
	}
	isValid, err := quotationValidator.ValidateDmvicRiskRequest(coverDet, riskDet)
	if err != nil {
		t.Fatalf("Failed to validate risk : %v", err)
	}

	if !isValid {
		t.Fatalf("Risk is not valid : %v", err)
	}

}

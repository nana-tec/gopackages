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

	println("Creating Dmvic Service  ")
	dmvicSrv, err := dmvic.NewDmvicServiceInstance(dmvicClient)
	if err != nil {
		t.Fatalf("Failed to create dmvic Srv : %v", err)
	}

	println("Creating Quotation Validator  ")
	quotationValidator, err := NewQuotationValidatorInstance(dmvicSrv)
	if err != nil {
		t.Fatalf("Failed to create quotation validator : %v", err)
	}

	println("Validating risk with dmvic  ")

	coverDet := &CoverDetails{
		StartDate: "2025-12-26",
		Period:    31,
	}
	riskDet := &dmvic.RiskDetails{
		RegistrationNumber: "KDM330X",
		ChassisNumber:      "",
	}
	validation, err := quotationValidator.ValidateDmvicRiskRequest(rootCtx, coverDet, riskDet)
	if err != nil {
		t.Fatalf("Failed to validate risk : %v", err)
	}

	// for production kdm has active cover else it should be true
	if dmvicConfig.Environment == dmvic.Production {
		if !validation.HasActiveCover {
			t.Fatalf("Risk should be having an active cover in production : %v", err)
		}

	}

	if dmvicConfig.Environment == dmvic.UAT {

		if validation.HasActiveCover {
			t.Fatalf("Risk should not be having an active cover in UAT : %v", err)
		}

	}

}

//go test -v --run TestQuotationValidator ./insurance/quotation

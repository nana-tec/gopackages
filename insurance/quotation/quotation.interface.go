package quotation

import (
	"context"

	dmvic "github.com/nana-tec/gopackages/Dmvic"
)

type CoverDetails struct {
	StartDate string
	Period    int
}

type RiskDetails struct {
	RegistrationNumber string
	ChassisNumber      string
	OtherDetails       map[string]any
}

type ClientDetails struct {
	Name      string
	IDnumber  string
	PinNumber string
}

type QuotationValidator interface {
	ValidateQuotationRequest(ctx context.Context, cover *CoverDetails, risk *RiskDetails, client *ClientDetails) (bool, error)
	ValidateDmvicRiskRequest(ctx context.Context, cover *CoverDetails, risk *dmvic.RiskDetails) (dmvic.MotorCoverValidationResponse, error)
}

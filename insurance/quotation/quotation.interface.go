package quotation

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
	ValidateQuotationRequest(cover *CoverDetails, risk *RiskDetails, client *ClientDetails) (bool, error)
	ValidateDmvicRiskRequest(cover *CoverDetails, risk *RiskDetails) (bool, error)
}

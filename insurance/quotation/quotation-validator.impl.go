package quotation

import (
	"fmt"
	"time"

	dmvic "github.com/nana-tec/gopackages/Dmvic"
)

type QuotationValidatorInstance struct {
	DmvicClient dmvic.Client
}

func NewQuotationValidatorInstance(DmvicClient dmvic.Client) (*QuotationValidatorInstance, error) {

	return &QuotationValidatorInstance{
		DmvicClient: DmvicClient,
	}, nil
}

func (qval *QuotationValidatorInstance) ValidateQuotationRequest(cover *CoverDetails, risk *RiskDetails, client *ClientDetails) (bool, error) {
	return true, nil
}

func (qval *QuotationValidatorInstance) ValidateDmvicRiskRequest(cover *CoverDetails, risk *RiskDetails) (bool, error) {
	t, err := time.Parse(time.DateOnly, cover.StartDate)
	if err != nil {
		return false, fmt.Errorf("Invalid start date  %w", err)
	}
	newDate := t.AddDate(0, 0, cover.Period)

	pendDate := newDate.Format("02/01/2006")
	endDateFormated := t.Format("02/01/2006")

	validationReq := &dmvic.DoubleInsuranceRequest{
		PolicyStartDate:           endDateFormated,
		PolicyEndDate:             pendDate,
		VehicleRegistrationNumber: risk.RegistrationNumber,
		ChassisNumber:             risk.ChassisNumber,
	}
	dmvicResp, err := qval.DmvicClient.ValidateDoubleInsurance(validationReq)
	if err != nil {
		return false, fmt.Errorf("failed to validate dmvic request  %w", err)
	}

	if dmvicResp.Success {
		// later check the resp
		return true, nil
	}

	return true, nil
}

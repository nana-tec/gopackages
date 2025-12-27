package quotation

import (
	"context"
	"fmt"
	"time"

	dmvic "github.com/nana-tec/gopackages/Dmvic"
)

type quotationValidatorInstance struct {
	dmvicService dmvic.DmvicService
}

func NewQuotationValidatorInstance(DmvicService dmvic.DmvicService) (QuotationValidator, error) {

	return &quotationValidatorInstance{
		dmvicService: DmvicService,
	}, nil
}

func (qval *quotationValidatorInstance) ValidateQuotationRequest(ctx context.Context, cover *CoverDetails, risk *RiskDetails, client *ClientDetails) (bool, error) {
	return true, nil
}

func (qval *quotationValidatorInstance) ValidateDmvicRiskRequest(ctx context.Context, cover *CoverDetails, risk *dmvic.RiskDetails) (dmvic.MotorCoverValidationResponse, error) {
	t, err := time.Parse(time.DateOnly, cover.StartDate)
	if err != nil {
		return dmvic.MotorCoverValidationResponse{}, fmt.Errorf("Invalid start date  %w", err)
	}

	startDateFormated := t.Format("02/01/2006")
	newDate := t.AddDate(0, 0, cover.Period)
	endDateFormated := newDate.Format("02/01/2006")

	reqCoverDet := dmvic.CoverDetails{
		StartDate: startDateFormated,
		EndDate:   endDateFormated,
	}
	return qval.dmvicService.MotorCoverValidation(ctx, reqCoverDet, risk)

}

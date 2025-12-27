package dmvic

import (
	"context"
	"errors"
	"fmt"
)

type CoverDetails struct {
	StartDate string
	EndDate   string
}

type RiskDetails struct {
	RegistrationNumber string
	ChassisNumber      string
}

type MotorCoverValidationResponse struct {
	HasActiveCover    bool
	ValidationMessage string
}

type DmvicService interface {
	MotorCoverValidation(ctx context.Context, coverdet CoverDetails, riskDet *RiskDetails) (MotorCoverValidationResponse, error)
	GetToken(ctx context.Context) (string, error)
}

type dmvicServiceInstance struct {
	dmvicClient Client
}

func NewDmvicServiceInstance(dmvicClient Client) (DmvicService, error) {
	return &dmvicServiceInstance{
		dmvicClient: dmvicClient,
	}, nil
}

func (ds *dmvicServiceInstance) MotorCoverValidation(ctx context.Context, coverdet CoverDetails, riskDet *RiskDetails) (MotorCoverValidationResponse, error) {

	var motorValidationResponse MotorCoverValidationResponse
	validationReq := &DoubleInsuranceRequest{
		PolicyStartDate:           coverdet.StartDate,
		PolicyEndDate:             coverdet.EndDate,
		VehicleRegistrationNumber: riskDet.RegistrationNumber,
		ChassisNumber:             riskDet.ChassisNumber,
	}
	dmvicResp, err := ds.dmvicClient.ValidateDoubleInsurance(validationReq)
	if err != nil {
		var appErr *ClientError // Target variable for the type assertion
		if errors.As(err, &appErr) {
			if appErr.DMVICCode == "ER001" {
				return MotorCoverValidationResponse{HasActiveCover: false, ValidationMessage: appErr.Message}, nil
			}
		} else {
			// this is an error we dont know about yet
			return motorValidationResponse, fmt.Errorf("failed to motor Vehicle cover  %w", err) // later return message for them to retry later
		}
		// we got here
		return motorValidationResponse, fmt.Errorf("failed to validate dmvic response  %w", err)
	}

	// no errors during validation
	if len(dmvicResp.CallbackObj.DoubleInsurance) > 0 {
		var doubleInDet DoubleInsuranceDetails = dmvicResp.CallbackObj.DoubleInsurance[0]
		if doubleInDet.ChassisNumber != "" {
			// later check iff not equal risk chassis num
			riskDet.ChassisNumber = doubleInDet.ChassisNumber
		}
		if doubleInDet.RegistrationNumber != "" {
			riskDet.RegistrationNumber = doubleInDet.RegistrationNumber
		}
		if doubleInDet.CertificateStatus == "Active" {
			valMessage := fmt.Sprintf("The Motor Has an active cover with %s ,Ending %s , Insurance Policy Number  %s", doubleInDet.MemberCompanyName, doubleInDet.CoverEndDate, doubleInDet.InsurancePolicyNo)
			return MotorCoverValidationResponse{HasActiveCover: true, ValidationMessage: valMessage}, nil
		}

		return MotorCoverValidationResponse{HasActiveCover: false, ValidationMessage: "No Active Cover"}, nil
	}

	return MotorCoverValidationResponse{HasActiveCover: false, ValidationMessage: "No Active Cover"}, nil

}

func (ds *dmvicServiceInstance) GetToken(ctx context.Context) (string, error) {

	tkn := ds.dmvicClient.GetToken()
	if tkn == "" {
		return "", fmt.Errorf("Unable to get Dmvic Token")
	}
	return tkn, nil
}

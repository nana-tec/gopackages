package risk

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	dmvic "github.com/nana-tec/gopackages/Dmvic"
	ntlogger "github.com/nana-tec/gopackages/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

type riskUsecase struct {
	repo   RiskRepository
	dmvic  dmvic.Client
	logger *ntlogger.Logger
}

func NewRiskUsecase(repo RiskRepository, dmvic dmvic.Client, logger *ntlogger.Logger) *riskUsecase {
	return &riskUsecase{
		repo:   repo,
		dmvic:  dmvic,
		logger: logger,
	}
}

func (uc *riskUsecase) motorRiskModelFromRisk(risk *MotorRisk) *MotorRiskModel {
	// generrate unique risk system ref
	uuid := uuid.New()
	return &MotorRiskModel{
		RiskSystemRef:      uuid.String(),
		RegistrationNumber: risk.RegistrationNumber,
		ChassisNumber:      risk.ChassisNumber,
		CarMake:            risk.CarMake,
		CarModel:           risk.CarModel,
		SeatingCapacity:    risk.SeatingCapacity,
		Tonnage:            risk.Tonnage,
		YearOfManufacture:  risk.YearOfManufacture,
		CubicCapacity:      risk.CubicCapacity,
		VehicleType:        risk.VehicleType,
		BodyType:           risk.BodyType,
		NameOfSacco:        risk.NameOfSacco,
	}
}

func (uc *riskUsecase) CreateUpdateRisk(ctx context.Context, motorRisk *MotorRisk) (string, error) {
	var rsk = uc.motorRiskModelFromRisk(motorRisk)
	_, err := uc.repo.GetMotorRiskByRegistrationNumberOrChassis(ctx, motorRisk.RegistrationNumber, motorRisk.ChassisNumber)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// create new risk

			err = uc.repo.SaveMotorRisk(ctx, rsk)
			if err != nil {
				return "", err
			}
			return rsk.RiskSystemRef, nil
		}
		return "", err
	}

	// update risk
	err = uc.repo.UpdateMotorRisk(ctx, rsk)
	if err != nil {
		return "", err
	}

	return "", nil
}

func (uc *riskUsecase) ValidateRiskDoubleInsurance(ctx context.Context, riskRef string, PolicyStartDate string, PolicyEndDate string) (riskValidateDoubleInsuranceResponse, error) {
	// this riskref can be registration number or chassis number

	riskDetail, err := uc.repo.GetMotorRiskByRef(ctx, riskRef)
	if err != nil {
		return riskValidateDoubleInsuranceResponse{}, err
	}

	if riskDetail == nil {
		return riskValidateDoubleInsuranceResponse{}, fmt.Errorf("risk not found: %s", riskRef)
	}

	// validate double insurance
	validationReq := &dmvic.DoubleInsuranceRequest{
		VehicleRegistrationNumber: riskDetail.RegistrationNumber,
		ChassisNumber:             riskDetail.ChassisNumber,
		PolicyStartDate:           PolicyStartDate,
		PolicyEndDate:             PolicyEndDate,
	}
	validationResponse, err := uc.dmvic.ValidateDoubleInsurance(validationReq)
	if err != nil {
		return riskValidateDoubleInsuranceResponse{}, err
	}

	if !validationResponse.Success {
		return riskValidateDoubleInsuranceResponse{}, fmt.Errorf("validation failed: %s", validationResponse.Error[0].ErrorText)
	}

	//return uc.repo.ValidateRiskDoubleInsurance(ctx, riskRef, PolicyStartDate, PolicyEndDate)
	return riskValidateDoubleInsuranceResponse{}, nil
}

func (uc *riskUsecase) GetRiskByRef(ctx context.Context, riskRef string) (*MotorRiskModel, error) {

	//return uc.repo.GetMotorRiskByRiskSystemRef(ctx, riskRef)

	return nil, nil
}

func (uc *riskUsecase) UpdateRisk(ctx context.Context, motorRisk *MotorRiskModel) error {

	//return uc.repo.GetMotorRiskByChassisNumber(ctx, chassisNumber)
	return nil
}

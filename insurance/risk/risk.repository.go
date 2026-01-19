package risk

import (
	"context"
	"fmt"

	ntlogger "github.com/nana-tec/gopackages/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// impliment risk repository interface in mongo db

type riskMongoRepository struct {
	db     *mongo.Database
	risks  *mongo.Collection
	logger *ntlogger.Logger
}

func NewRiskMongoRepository(db *mongo.Database, logger *ntlogger.Logger) *riskMongoRepository {
	repo := &riskMongoRepository{
		db:     db,
		risks:  db.Collection("risks"),
		logger: logger,
	}
	return repo
}

func (repo *riskMongoRepository) GetMotorRiskByRegistrationNumber(ctx context.Context, registrationNumber string) (*MotorRiskModel, error) {

	var rsk MotorRiskModel
	err := repo.risks.FindOne(ctx, bson.M{"registration_number": registrationNumber}).Decode(&rsk)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("risk not found: %s", registrationNumber)
		}
		return nil, err
	}
	return &rsk, nil
}

func (repo *riskMongoRepository) GetMotorRiskByChassisNumber(ctx context.Context, chassisNumber string) (*MotorRiskModel, error) {
	var rsk MotorRiskModel
	err := repo.risks.FindOne(ctx, bson.M{"chassis_number": chassisNumber}).Decode(&rsk)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("risk not found: %s", chassisNumber)
		}
		return nil, err
	}
	return &rsk, nil
}

func (repo *riskMongoRepository) GetMotorRiskByRiskSystemRef(ctx context.Context, riskSystemRef string) (*MotorRiskModel, error) {
	var rsk MotorRiskModel
	err := repo.risks.FindOne(ctx, bson.M{"risk_system_ref": riskSystemRef}).Decode(&rsk)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("risk not found: %s", riskSystemRef)
		}
		return nil, err
	}
	return &rsk, nil
}

func (repo *riskMongoRepository) GetMotorRiskByRef(ctx context.Context, riskRef string) (*MotorRiskModel, error) {
	var rsk MotorRiskModel

	filter := bson.D{
		{"$or", bson.A{
			bson.D{{"registration_number", riskRef}},
			bson.D{{"chassis_number", riskRef}},
		}},
	}
	err := repo.risks.FindOne(ctx, filter).Decode(&rsk)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("risk not found: %s", riskRef)
		}
		return nil, err
	}
	return &rsk, nil
}

func (repo *riskMongoRepository) GetMotorRiskByRegistrationNumberOrChassis(ctx context.Context, registrationNumber string, chassisNumber string) (*MotorRiskModel, error) {

	var rsk MotorRiskModel
	filter := bson.D{
		{"$or", bson.A{
			bson.D{{"registration_number", registrationNumber}},
			bson.D{{"chassis_number", chassisNumber}},
		}},
	}
	err := repo.risks.FindOne(ctx, filter).Decode(&rsk)
	if err != nil {
		return nil, err
	}
	return &rsk, nil

}
func (repo *riskMongoRepository) SaveMotorRisk(ctx context.Context, motorRisk *MotorRiskModel) error {

	_, err := repo.risks.InsertOne(ctx, motorRisk)
	if err != nil {
		return err
	}

	return nil
}
func (repo *riskMongoRepository) UpdateMotorRisk(ctx context.Context, motorRisk *MotorRiskModel) error {

	_, err := repo.risks.UpdateOne(ctx, bson.M{"risk_system_ref": motorRisk.RiskSystemRef}, bson.M{"$set": motorRisk})
	if err != nil {
		return err
	}

	return nil
}
func (repo *riskMongoRepository) DeleteMotorRisk(ctx context.Context, motorRisk *MotorRiskModel) error {

	_, err := repo.risks.DeleteOne(ctx, bson.M{"risk_system_ref": motorRisk.RiskSystemRef})
	if err != nil {
		return err
	}
	return nil
}

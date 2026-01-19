package risk

import (
	dmvic "github.com/nana-tec/gopackages/Dmvic"
	ntlogger "github.com/nana-tec/gopackages/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

func NewRiskService(db *mongo.Database, dmvic dmvic.Client, logger *ntlogger.Logger) (*riskUsecase, error) {

	repo := NewRiskMongoRepository(db, logger)
	riskUsecase := NewRiskUsecase(repo, dmvic, logger)
	return riskUsecase, nil
}

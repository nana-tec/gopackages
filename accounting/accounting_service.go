package accounting

// should expose an instance of accounting service

import (
	"go.mongodb.org/mongo-driver/mongo"
)

func NewAccountingService(db *mongo.Database) *AccountingService {

	return &AccountingService{
		db:       db,
		accounts: db.Collection("accounts"),
		journals: db.Collection("journals"),
	}
}

package accounting

import (
	"context"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --------------------------
//  Account CRUD
// --------------------------

func (s *AccountingService) CreateAccount(ctx context.Context, accType AccountType, initialBalance decimal.Decimal, name string) (*Account, error) {
	acc := &Account{
		ID:        primitive.NewObjectID(),
		Type:      accType,
		Name:      name,
		CreatedAt: time.Now(),
	}
	acc.SetBalance(initialBalance)

	_, err := s.accounts.InsertOne(ctx, acc)
	if err != nil {
		return nil, err
	}
	return acc, nil
}

func (s *AccountingService) GetAccountByID(ctx context.Context, accountID primitive.ObjectID) (*Account, error) {
	var acc Account
	err := s.accounts.FindOne(ctx, bson.M{"_id": accountID}).Decode(&acc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("account not found: %s", accountID.Hex())
		}
		return nil, err
	}
	return &acc, nil
}

func (s *AccountingService) GetAccountBalance(ctx context.Context, accountID primitive.ObjectID) (decimal.Decimal, error) {
	acc, err := s.GetAccountByID(ctx, accountID)
	if err != nil {
		return decimal.Zero, err
	}
	return acc.GetBalance(), nil
}

// --------------------------
//  Double-Entry Posting
// --------------------------

func (s *AccountingService) postDoubleEntry(
	ctx context.Context,
	txType TransactionType,
	amount decimal.Decimal,
	debitAccID, creditAccID primitive.ObjectID,
	tranRef string,
) error {
	if amount.LessThanOrEqual(decimal.Zero) {
		return fmt.Errorf("amount must be > 0")
	}

	return s.runInTransaction(ctx, func(sc mongo.SessionContext) error {
		// 1. Update account balances
		if err := s.incrementBalance(sc, debitAccID, amount.Neg()); err != nil {
			return err
		}
		if err := s.incrementBalance(sc, creditAccID, amount); err != nil {
			return err
		}

		// 2. Insert journal entry (double-entry)
		entry := &JournalEntry{
			ID:            primitive.NewObjectID(),
			Type:          txType,
			Amount:        amount.String(),
			DebitAccount:  debitAccID,
			CreditAccount: creditAccID,
			CreatedAt:     time.Now(),
			TranRef:       tranRef,
		}
		_, err := s.journals.InsertOne(sc, entry)
		return err
	})
}

// Client Top-Up: Debit Gateway (asset), Credit Client (liability)
func (s *AccountingService) ClientAccountTopUp(ctx context.Context, clientAccID, gatewayAccID primitive.ObjectID, amount decimal.Decimal, tranRef string) error {
	return s.postDoubleEntry(ctx, TopUp, amount, gatewayAccID, clientAccID, tranRef)
}

// Premium Payment: Debit Client (liability), Credit Underwriter (liability)
func (s *AccountingService) ClientPremiumPayment(ctx context.Context, clientAccID, underwriterAccID primitive.ObjectID, amount decimal.Decimal, tranRef string) error {
	return s.postDoubleEntry(ctx, PremiumPayment, amount, clientAccID, underwriterAccID, tranRef)
}

// Commission: Debit Underwriter (expense), Credit Agent (revenue)
func (s *AccountingService) PostAgentCommission(ctx context.Context, underwriterAccID, agentAccID primitive.ObjectID, amount decimal.Decimal, tranRef string) error {
	return s.postDoubleEntry(ctx, CommissionPayment, amount, underwriterAccID, agentAccID, tranRef)
}

// Helper: increment balance atomically
func (s *AccountingService) incrementBalance(sc mongo.SessionContext, accountID primitive.ObjectID, delta decimal.Decimal) error {
	acc, err := s.getAccountInSession(sc, accountID)
	if err != nil {
		return err
	}
	newBal := acc.GetBalance().Add(delta)
	filter := bson.M{"_id": accountID}
	update := bson.M{"$set": bson.M{"balance": newBal.String()}}
	_, err = s.accounts.UpdateOne(sc, filter, update)
	return err
}

func (s *AccountingService) getAccountInSession(sc mongo.SessionContext, accountID primitive.ObjectID) (*Account, error) {
	var acc Account
	err := s.accounts.FindOne(sc, bson.M{"_id": accountID}).Decode(&acc)
	if err != nil {
		return nil, err
	}
	return &acc, nil
}

// --------------------------
//  Journal History
// --------------------------

func (s *AccountingService) GetJournalEntries(ctx context.Context, limit, skip int64) ([]JournalEntry, error) {
	if limit <= 0 {
		limit = 50
	}
	if skip < 0 {
		skip = 0
	}

	opts := options.Find().
		SetSort(bson.M{"created_at": -1}).
		SetLimit(limit).
		SetSkip(skip)

	cursor, err := s.journals.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []JournalEntry
	if err = cursor.All(ctx, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *AccountingService) GetJournalEntriesByRef(ctx context.Context, tranRef string) ([]JournalEntry, error) {
	filter := bson.M{"tranref": tranRef}
	cursor, err := s.journals.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []JournalEntry
	if err = cursor.All(ctx, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// --------------------------
//  LEDGER RECONCILIATION (Double-Entry)
// --------------------------

func (s *AccountingService) ReconcileAccount(ctx context.Context, accountID primitive.ObjectID) (*ReconciliationResult, error) {
	acc, err := s.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	// Fetch all journal legs affecting this account
	filter := bson.M{
		"$or": []bson.M{
			{"debit_account": accountID},
			{"credit_account": accountID},
		},
	}
	cursor, err := s.journals.Find(ctx, filter, options.Find().SetSort(bson.M{"created_at": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entries []JournalEntry
	if err = cursor.All(ctx, &entries); err != nil {
		return nil, err
	}

	var computed decimal.Decimal
	for _, e := range entries {
		amt := e.GetAmount()
		if e.DebitAccount == accountID {
			computed = computed.Add(amt)
		}
		if e.CreditAccount == accountID {
			computed = computed.Sub(amt)
		}
	}

	stored := acc.GetBalance()
	discrepancy := computed.Sub(stored)
	status := Reconciled
	if len(entries) == 0 {
		status = NoTransactions
	} else if !discrepancy.IsZero() {
		status = Discrepancy
	}

	return &ReconciliationResult{
		AccountID:       accountID,
		AccountType:     acc.Type,
		StoredBalance:   stored,
		ComputedBalance: computed,
		Discrepancy:     discrepancy,
		Status:          status,
		JournalCount:    len(entries),
	}, nil
}

func (s *AccountingService) GetReconciliationReport(ctx context.Context) ([]ReconciliationResult, error) {
	cursor, err := s.accounts.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []Account
	if err = cursor.All(ctx, &accounts); err != nil {
		return nil, err
	}

	var report []ReconciliationResult
	for _, acc := range accounts {
		res, err := s.ReconcileAccount(ctx, acc.ID)
		if err != nil {
			return nil, err
		}
		report = append(report, *res)
	}
	return report, nil
}

// --------------------------
//  Helpers
// --------------------------

func (s *AccountingService) runInTransaction(ctx context.Context, fn func(mongo.SessionContext) error) error {
	session, err := s.db.Client().StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(ctx)

	_, err = session.WithTransaction(ctx, func(sc mongo.SessionContext) (interface{}, error) {
		return nil, fn(sc)
	})
	return err
}

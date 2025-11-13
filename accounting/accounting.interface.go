package accounting

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// --------------------------
//  Types & Constants
// --------------------------

type AccountType string

const (
	UnderwriterPremiumPayable AccountType = "UnderwriterPremiumPayable"
	AgentCommissionEarned     AccountType = "AgentCommissionEarned"
	PaymentGateway            AccountType = "PaymentGateway"
	ClientInsurance           AccountType = "ClientInsurance"
)

type TransactionType string

const (
	TopUp             TransactionType = "TopUp"
	PremiumPayment    TransactionType = "PremiumPayment"
	CommissionPayment TransactionType = "CommissionPayment"
)

// --------------------------
//  Models
// --------------------------

type Account struct {
	ID        primitive.ObjectID `bson:"_id"`
	Type      AccountType        `bson:"type"`
	Balance   string             `bson:"balance"` // decimal string
	Name      string             `bson:"name"`
	CreatedAt time.Time          `bson:"created_at"`
}

func (a *Account) GetBalance() decimal.Decimal {
	d, _ := decimal.NewFromString(a.Balance)
	return d
}

func (a *Account) SetBalance(d decimal.Decimal) {
	a.Balance = d.String()
}

// JournalEntry: One transaction = two legs (debit + credit)
type JournalEntry struct {
	ID            primitive.ObjectID `bson:"_id"`
	TransactionID primitive.ObjectID `bson:"transaction_id"` // optional group
	Type          TransactionType    `bson:"type"`
	Amount        string             `bson:"amount"`
	TranRef       string             `bson:"tranref"` // external reference
	DebitAccount  primitive.ObjectID `bson:"debit_account"`
	CreditAccount primitive.ObjectID `bson:"credit_account"`
	CreatedAt     time.Time          `bson:"created_at"`
}

func (j JournalEntry) GetAmount() decimal.Decimal {
	d, _ := decimal.NewFromString(j.Amount)
	return d
}

func (j JournalEntry) String() string {
	return fmt.Sprintf("[%s] %s | %s | Dr:%s | Cr:%s | %s | Tranref: %s",
		j.ID.Hex()[:8],
		j.Type,
		j.GetAmount().StringFixed(2),
		j.DebitAccount.Hex()[:8],
		j.CreditAccount.Hex()[:8],
		j.CreatedAt.Format("2006-01-02 15:04:05"),
		j.TranRef,
	)
}

// --------------------------
//  Reconciliation Result
// --------------------------

type ReconciliationStatus string

const (
	Reconciled     ReconciliationStatus = "RECONCILED"
	Discrepancy    ReconciliationStatus = "DISCREPANCY"
	NoTransactions ReconciliationStatus = "NO_TRANSACTIONS"
)

type ReconciliationResult struct {
	AccountID       primitive.ObjectID   `json:"account_id"`
	AccountType     AccountType          `json:"account_type"`
	StoredBalance   decimal.Decimal      `json:"stored_balance"`
	ComputedBalance decimal.Decimal      `json:"computed_balance"`
	Discrepancy     decimal.Decimal      `json:"discrepancy"`
	Status          ReconciliationStatus `json:"status"`
	JournalCount    int                  `json:"journal_count"`
}

// --------------------------
//  Service
// --------------------------

type AccountingService struct {
	db       *mongo.Database
	accounts *mongo.Collection
	journals *mongo.Collection
}

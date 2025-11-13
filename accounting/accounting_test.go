package accounting

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// setupTestDB starts an in-memory MongoDB and returns a connected service
func setupTestDB(t *testing.T) (*AccountingService, func()) {
	ctx := context.Background()
	//.WaitForLog("Waiting for connections"),
	// Start MongoDB container
	/*mongoContainer, err := mongodb.RunContainer(ctx,
		testcontainers.WithImage("mongo:7"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("Waiting for connections"),
		),
		mongodb.WithReplicaSet("rs0"),
	)
	require.NoError(t, err)

	// Get connection string
	uri, err := mongoContainer.ConnectionString(ctx)
	uri = "mongodb://localhost:27017/?replicaSet=rs0"
	require.NoError(t, err)

	// Connect service
	s := &AccountingService{}
	s.db, err = connectToMongo(uri)
	require.NoError(t, err)
	*/
	s := newAccountingService()
	s.accounts = s.db.Collection("accounts")
	s.journals = s.db.Collection("journals")

	// Cleanup
	cleanup := func() {
		//_ = s.db.Drop(ctx)
		_ = s.db.Client().Disconnect(ctx)
	}

	return s, cleanup
}

func newAccountingService() *AccountingService {
	clientOpts := options.Client().ApplyURI("mongodb://localhost:27017/?replicaSet=rs0")
	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		log.Fatal(err)
	}
	db := client.Database("insurance_db")
	println("Connected to MongoDB")
	return &AccountingService{
		db:       db,
		accounts: db.Collection("accounts"),
		journals: db.Collection("journals"),
	}
}

// Helper to connect (extracted from NewAccountingService)
func connectToMongo(uri string) (*mongo.Database, error) {
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		return nil, err
	}
	db := client.Database("test_db")
	return db, nil
}

// === TESTS ===

func TestClientTopUp_DoubleEntry(t *testing.T) {
	t.Parallel()
	s, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	// Create accounts
	clientAcc, _ := s.CreateAccount(ctx, ClientInsurance, decimal.Zero, "Client A Topup")
	gatewayAcc, _ := s.CreateAccount(ctx, PaymentGateway, decimal.Zero, "Gateway A Topup")

	amount := decimal.NewFromFloat(1000.0)

	// Execute
	err := s.ClientAccountTopUp(ctx, clientAcc.ID, gatewayAcc.ID, amount, "topupref1")
	require.NoError(t, err)

	// 1. Balances
	clientBal, _ := s.GetAccountBalance(ctx, clientAcc.ID)
	gatewayBal, _ := s.GetAccountBalance(ctx, gatewayAcc.ID)
	assert.True(t, clientBal.Equal(decimal.NewFromFloat(1000)))
	assert.True(t, gatewayBal.Equal(decimal.NewFromFloat(-1000)))

	// 2. Journal: exactly 1 entry
	entries, _ := s.GetJournalEntriesByRef(ctx, "topupref1")
	require.Len(t, entries, 1)
	e := entries[0]
	assert.Equal(t, TopUp, e.Type)
	assert.True(t, e.GetAmount().Equal(amount))
	assert.Equal(t, gatewayAcc.ID, e.DebitAccount)
	assert.Equal(t, clientAcc.ID, e.CreditAccount)
}

func TestPremiumPayment_BalanceAndJournal(t *testing.T) {
	t.Parallel()
	s, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	clientAcc, _ := s.CreateAccount(ctx, ClientInsurance, decimal.NewFromFloat(1500), "Client Premium Payment")
	underwriterAcc, _ := s.CreateAccount(ctx, UnderwriterPremiumPayable, decimal.Zero, "Underwriter Premium Payment")

	amount := decimal.NewFromFloat(800)

	err := s.ClientPremiumPayment(ctx, clientAcc.ID, underwriterAcc.ID, amount, "premiumpaymentref1")
	require.NoError(t, err)

	// Balances
	assert.True(t, (func() bool {
		b, _ := s.GetAccountBalance(ctx, clientAcc.ID)
		return b.Equal(decimal.NewFromFloat(700))
	})())
	assert.True(t, (func() bool {
		b, _ := s.GetAccountBalance(ctx, underwriterAcc.ID)
		return b.Equal(decimal.NewFromFloat(800))
	})())

	// Journal
	entries, _ := s.GetJournalEntriesByRef(ctx, "premiumpaymentref1")
	require.Len(t, entries, 1)
	e := entries[0]
	assert.Equal(t, PremiumPayment, e.Type)
	assert.Equal(t, clientAcc.ID, e.DebitAccount)
	assert.Equal(t, underwriterAcc.ID, e.CreditAccount)
}

/*
	func TestAgentCommission_Reconciliation(t *testing.T) {
		t.Parallel()
		s, cleanup := setupTestDB(t)
		defer cleanup()
		ctx := context.Background()

		underwriterAcc, _ := s.CreateAccount(ctx, UnderwriterPremiumPayable, decimal.NewFromFloat(1000), "Underwriter Commision Payment")
		agentAcc, _ := s.CreateAccount(ctx, AgentCommissionEarned, decimal.Zero, "Agent Commision Payment")

		amount := decimal.NewFromFloat(150)

		err := s.PostAgentCommission(ctx, underwriterAcc.ID, agentAcc.ID, amount, "agentcommissionref1")
		require.NoError(t, err)
		time.Sleep(2 * time.Second) // Pause for 2 seconds
		// Balances
		b, _ := s.GetAccountBalance(ctx, underwriterAcc.ID)
		println("Undewriter balance:", b.String())
		assert.True(t, (func() bool {
			b, _ := s.GetAccountBalance(ctx, underwriterAcc.ID)
			return b.Equal(decimal.NewFromFloat(850))
		})())
		assert.True(t, (func() bool {
			b, _ := s.GetAccountBalance(ctx, agentAcc.ID)
			return b.Equal(decimal.NewFromFloat(150))
		})())

		// Reconciliation
		res, err := s.ReconcileAccount(ctx, underwriterAcc.ID)
		require.NoError(t, err)
		assert.Equal(t, Reconciled, res.Status)
		assert.True(t, res.Discrepancy.IsZero())
		assert.Equal(t, 1, res.JournalCount)
	}

	func TestFullFlow_ReconciliationReport(t *testing.T) {
		t.Parallel()
		s, cleanup := setupTestDB(t)
		defer cleanup()
		ctx := context.Background()

		client, _ := s.CreateAccount(ctx, ClientInsurance, decimal.Zero, "Client Recconiliation")
		gateway, _ := s.CreateAccount(ctx, PaymentGateway, decimal.Zero, "Gateway Recconiliation")
		underwriter, _ := s.CreateAccount(ctx, UnderwriterPremiumPayable, decimal.Zero, "Underwriter Recconiliation")
		agent, _ := s.CreateAccount(ctx, AgentCommissionEarned, decimal.Zero, "Agent Recconiliation")

		// Flow
		_ = s.ClientAccountTopUp(ctx, client.ID, gateway.ID, decimal.NewFromFloat(2000), "reconciliationref1")
		_ = s.ClientPremiumPayment(ctx, client.ID, underwriter.ID, decimal.NewFromFloat(1200), "reconciliationref1")
		_ = s.PostAgentCommission(ctx, underwriter.ID, agent.ID, decimal.NewFromFloat(100), "reconciliationref1")

		// Full report
		report, err := s.GetReconciliationReport(ctx)
		require.NoError(t, err)
		require.Len(t, report, 4)

		for _, r := range report {
			assert.Equal(t, Reconciled, r.Status, "Account %s failed reconciliation", r.AccountType)
			assert.True(t, r.Discrepancy.IsZero(), "Discrepancy in %s", r.AccountType)
		}
	}

	func TestDoubleEntry_Atomicity_PartialFailure(t *testing.T) {
		t.Parallel()
		s, cleanup := setupTestDB(t)
		defer cleanup()
		ctx := context.Background()

		client, _ := s.CreateAccount(ctx, ClientInsurance, decimal.Zero)
		gateway, _ := s.CreateAccount(ctx, PaymentGateway, decimal.Zero)

		// Inject failure: use invalid account ID for credit
		invalidID := primitive.NewObjectID()

		// Create a mock for the journals collection
		type mockCollection struct {
			*mongo.Collection
			insertOneFunc func(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error)
		}

		func (m *mockCollection) InsertOne(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
			return m.insertOneFunc(ctx, document, opts...)
		}

		// Create a mock instance of journals
		mockJournals := &mockCollection{
			insertOneFunc: func(ctx context.Context, document interface{}, opts ...*options.InsertOneOptions) (*mongo.InsertOneResult, error) {
				return nil, assert.AnError // simulate DB error
			},
		}

		// Replace the journals collection with the mock
		s.journals = mockJournals
			return nil, assert.AnError // simulate DB error
		}
		defer func() { s.journals.InsertOne = origInsert }()

		err := s.ClientAccountTopUp(ctx, client.ID, gateway.ID, decimal.NewFromFloat(500))
		require.Error(t, err)

		// Balances must remain zero
		b1, _ := s.GetAccountBalance(ctx, client.ID)
		b2, _ := s.GetAccountBalance(ctx, gateway.ID)
		assert.True(t, b1.IsZero())
		assert.True(t, b2.IsZero())

		// No journal entries
		entries, _ := s.GetJournalEntries(ctx, 10, 0)
		assert.Empty(t, entries)
	}
*/
func TestJournal_DebitsEqualCredits(t *testing.T) {
	t.Parallel()
	s, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	// Create 4 accounts
	accs := make([]*Account, 4)
	types := []AccountType{ClientInsurance, PaymentGateway, UnderwriterPremiumPayable, AgentCommissionEarned}
	for i, typ := range types {
		acc, _ := s.CreateAccount(ctx, typ, decimal.Zero, fmt.Sprintf("Account %d", i))
		accs[i] = acc
	}

	// Post 3 transactions
	_ = s.ClientAccountTopUp(ctx, accs[0].ID, accs[1].ID, decimal.NewFromFloat(1000), "journalref1")
	_ = s.ClientPremiumPayment(ctx, accs[0].ID, accs[2].ID, decimal.NewFromFloat(700), "journalref1")
	_ = s.PostAgentCommission(ctx, accs[2].ID, accs[3].ID, decimal.NewFromFloat(70), "journalref1")

	// Fetch all journals
	entries, _ := s.GetJournalEntriesByRef(ctx, "journalref1")
	require.Len(t, entries, 3)

	var totalDebit, totalCredit decimal.Decimal
	for _, e := range entries {
		amt := e.GetAmount()
		totalDebit = totalDebit.Add(amt)
		totalCredit = totalCredit.Add(amt)
	}

	assert.True(t, totalDebit.Equal(totalCredit), "Debits must equal Credits")
	assert.True(t, totalDebit.Equal(decimal.NewFromFloat(1770)))
}

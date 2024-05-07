package port

import (
	"time"

	"github.com/Just-Goo/grpc-go-server/internal/application/domain/bank"
	"github.com/google/uuid"
)

// ports are basically Go interfaces where the implementation is on the application layer
type HelloServicePort interface {
	GenerateHello(name string) string
}

type BankServicePort interface {
	FindCurrentBalance(account string) (float64, error)
	CreateExchangeRate(r bank.ExchangeRate) (uuid.UUID, error)
	FindExchangeRate(fromCur string, toCur string, ts time.Time) (float64, error)
	CreateTransaction(acct string, t bank.Transaction) (uuid.UUID, error)
	CalculateTransactionSummary(tcur *bank.TransactionSummary, trans bank.Transaction) error
	Transfer(tt bank.TransferTransaction) (uuid.UUID, bool, error)
}

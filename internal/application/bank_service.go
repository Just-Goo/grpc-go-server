package application

import (
	"fmt"
	"log"
	"time"

	"github.com/Just-Goo/grpc-go-server/internal/adapter/database"
	"github.com/Just-Goo/grpc-go-server/internal/application/domain/bank"
	"github.com/Just-Goo/grpc-go-server/internal/port"
	"github.com/google/uuid"
)

type BankService struct {
	db port.BankDatabasePort
}

func NewBankService(dbPort port.BankDatabasePort) *BankService {
	return &BankService{
		db: dbPort,
	}
}

func (b *BankService) FindCurrentBalance(account string) (float64, error) {
	bankAccount, err := b.db.GetBankAccountByAccountNumber(account)
	if err != nil {
		log.Println("error on 'find current balance'", err)
		return 0, err
	}

	return bankAccount.CurrentBalance, nil
}

func (b *BankService) CreateExchangeRate(r bank.ExchangeRate) (uuid.UUID, error) {
	newUuid := uuid.New()
	now := time.Now()

	exchangeRateOrm := database.BankExchangeRateOrm{
		ExchangeRateUuid:   newUuid,
		FromCurrency:       r.FromCurrency,
		ToCurrency:         r.ToCurrency,
		Rate:               r.Rate,
		ValidFromTimestamp: r.ValidFromTimestamp,
		ValidToTimestamp:   r.ValidToTimestamp,
		CreatedAt:          now,
	}
	return b.db.CreateExchangeRate(&exchangeRateOrm)
}

func (b *BankService) FindExchangeRate(fromCur string, toCur string, ts time.Time) (float64, error) {
	exchangeRate, err := b.db.GetExchangeRateAtTimestamp(fromCur, toCur, ts)
	if err != nil {
		return 0, err
	}

	return float64(exchangeRate.Rate), nil
}

func (b *BankService) CreateTransaction(acct string, t bank.Transaction) (uuid.UUID, error) {
	newUUID := uuid.New()
	now := time.Now()

	bankAccountOrm, err := b.db.GetBankAccountByAccountNumber(acct)
	if err != nil {
		return uuid.Nil, fmt.Errorf("can't find account number %v : %v", acct, err.Error())
	}

	if t.TransactionType == bank.TransactionTypeOUT && bankAccountOrm.CurrentBalance < t.Amount {
		return bankAccountOrm.AccountUuid, fmt.Errorf(
			"insufficient account balance %v for [out] transaction amount %v",
			bankAccountOrm.CurrentBalance, t.Amount,
		)
	}

	transactionOrm := database.BankTransactionOrm{
		TransactionUuid:      newUUID,
		AccountUuid:          bankAccountOrm.AccountUuid,
		TransactionTimestamp: now,
		Amount:               t.Amount,
		TransactionType:      t.TransactionType,
		Notes:                t.Notes,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	savedUUID, err := b.db.CreateTransaction(bankAccountOrm, transactionOrm)

	return savedUUID, err
}

func (b *BankService) CalculateTransactionSummary(tcur *bank.TransactionSummary, trans bank.Transaction) error {

	switch trans.TransactionType {
	case bank.TransactionTypeIN:
		tcur.SumIn += trans.Amount
	case bank.TransactionTypeOUT:
		tcur.SumOut += trans.Amount
	default:
		return fmt.Errorf("unknown transaction type %v", trans.TransactionType)
	}

	tcur.SumTotal = tcur.SumIn - tcur.SumOut

	return nil
}

func (b *BankService) Transfer(tt bank.TransferTransaction) (uuid.UUID, bool, error) {
	now := time.Now()

	fromAccountOrm, err := b.db.GetBankAccountByAccountNumber(tt.FromAccountNumber)
	if err != nil {
		log.Printf("can't find account for this account number %v : %v", tt.FromAccountNumber, err)
		return uuid.Nil, false, bank.ErrTransferSourceAccountNotFound
	}

	if fromAccountOrm.CurrentBalance < tt.Amount {
		return uuid.Nil, false, bank.ErrTransferTransactionPair
	}

	toAccountOrm, err := b.db.GetBankAccountByAccountNumber(tt.ToAccountNumber)
	if err != nil {
		log.Printf("can't find account for this account number %v : %v", tt.ToAccountNumber, err)
		return uuid.Nil, false, bank.ErrTransferDestinationAccountNotFound
	}

	fromTransactionOrm := database.BankTransactionOrm{
		TransactionUuid:      uuid.New(),
		TransactionTimestamp: now,
		TransactionType:      bank.TransactionTypeOUT,
		AccountUuid:          fromAccountOrm.AccountUuid,
		Amount:               tt.Amount,
		Notes:                "Transfer out to " + tt.ToAccountNumber,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	toTransactionOrm := database.BankTransactionOrm{
		TransactionUuid:      uuid.New(),
		TransactionTimestamp: now,
		TransactionType:      bank.TransactionTypeIN,
		AccountUuid:          toAccountOrm.AccountUuid,
		Amount:               tt.Amount,
		Notes:                "Transfer in from " + tt.FromAccountNumber,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	// create transfer request
	newTransferUuid := uuid.New()

	// 'TransferSuccess - false' => assume transfer failed when creating record for the first time
	transferOrm := database.BankTransferOrm{
		TransferUuid:      newTransferUuid,
		FromAccountUuid:   fromAccountOrm.AccountUuid,
		ToAccountUuid:     toAccountOrm.AccountUuid,
		Currency:          tt.Currency,
		Amount:            tt.Amount,
		TransferTimestamp: now,
		TransferSuccess:   false,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if _, err := b.db.CreateTransfer(transferOrm); err != nil {
		log.Printf("can't create transfer from %v to %v : %v", tt.FromAccountNumber, tt.ToAccountNumber, err)
		return uuid.Nil, false, bank.ErrTransferRecordFailed
	}

	// create transaction pair. If transaction pair created successfully, update 'TransferSuccess - true' otherwise update it to false
	if transferPairSuccess, _ := b.db.CreateTransferTransactionPair(fromAccountOrm, toAccountOrm, fromTransactionOrm,
		toTransactionOrm); transferPairSuccess {
		b.db.UpdateTransferStatus(transferOrm, true) // handle error
		return newTransferUuid, true, nil
	} else {
		return uuid.Nil, false, bank.ErrTransferTransactionPair
	}

}

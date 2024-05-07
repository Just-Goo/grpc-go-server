package database

import (
	"fmt"
	"log"
	"time"

	"github.com/Just-Goo/grpc-go-server/internal/application/domain/bank"
	"github.com/google/uuid"
)

func (d *DatabaseAdapter) GetBankAccountByAccountNumber(acct string) (BankAccountOrm, error) {
	var bankAccountOrm BankAccountOrm

	if err := d.db.First(&bankAccountOrm, "account_number = ?", acct).Error; err != nil {
		log.Printf("can't find bank account number %s : %v\n", acct, err)
		return bankAccountOrm, err
	}

	return bankAccountOrm, nil
}

func (d *DatabaseAdapter) CreateExchangeRate(r *BankExchangeRateOrm) (uuid.UUID, error) {
	if err := d.db.Create(r).Error; err != nil {
		return uuid.Nil, err
	}

	return r.ExchangeRateUuid, nil
}

func (d *DatabaseAdapter) GetExchangeRateAtTimestamp(fromCur string, toCur string, ts time.Time) (BankExchangeRateOrm, error) {
	var exchangeRateOrm BankExchangeRateOrm
	err := d.db.First(&exchangeRateOrm, "from_currency = ? "+" AND to_currency = ? "+
		" AND (? BETWEEN valid_from_timestamp and valid_to_timestamp)", fromCur, toCur, ts).Error

	return exchangeRateOrm, err
}

func (d *DatabaseAdapter) CreateTransaction(acct BankAccountOrm, t BankTransactionOrm) (uuid.UUID, error) {
	tx := d.db.Begin()

	if err := tx.Create(t).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}

	// calculate new account balance
	newAmount := t.Amount

	if t.TransactionType == bank.TransactionTypeOUT {
		newAmount = -1 * t.Amount
	}

	newAccountBalance := acct.CurrentBalance + newAmount

	if err := tx.Model(&acct).Updates(
		map[string]interface{}{
			"current_balance": newAccountBalance,
			"updated_at":      time.Now(),
		},
	).Error; err != nil {
		tx.Rollback()
		return uuid.Nil, err
	}

	tx.Commit()

	return t.TransactionUuid, nil
}

func (d *DatabaseAdapter) CreateTransfer(transfer BankTransferOrm) (uuid.UUID, error) {
	if err := d.db.Create(transfer).Error; err != nil {
		return uuid.Nil, err
	}

	return transfer.TransferUuid, nil
}

func (d *DatabaseAdapter) CreateTransferTransactionPair(fromAccountOrm BankAccountOrm, toAccountOrm BankAccountOrm,
	fromTransactionOrm BankTransactionOrm, toTransactionOrm BankTransactionOrm) (bool, error) {

	tx := d.db.Begin()

	// create from account transaction
	if err := d.db.Create(fromTransactionOrm).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	// create from to transaction
	if err := d.db.Create(toTransactionOrm).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	// check if customer wants to send money to him/herself
	if fromAccountOrm.AccountNumber == toAccountOrm.AccountNumber {
		return false, fmt.Errorf("cannot send money to yourself")
	}

	// calculate new account balance (fromAccount)
	fromAccountBalanceNew := fromAccountOrm.CurrentBalance - fromTransactionOrm.Amount

	// check if fromAccount has enough money to send
	if fromAccountBalanceNew < 0 {
		return false, fmt.Errorf("insufficient funds")
	}

	if err := tx.Model(&fromAccountOrm).Updates(
		map[string]interface{}{
			"current_balance": fromAccountBalanceNew,
			"updated_at":      time.Now(),
		},
	).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	// calculate new account balance (toAccount)
	toAccountBalanceNew := toAccountOrm.CurrentBalance + toTransactionOrm.Amount

	if err := tx.Model(&toAccountOrm).Updates(
		map[string]interface{}{
			"current_balance": toAccountBalanceNew,
			"updated_at":      time.Now(),
		},
	).Error; err != nil {
		tx.Rollback()
		return false, err
	}

	tx.Commit()

	return true, nil
}

func (d *DatabaseAdapter) UpdateTransferStatus(transfer BankTransferOrm, status bool) error {
	if err := d.db.Model(&transfer).Updates(
		map[string]interface{}{
			"transfer_success": status,
			"updated_at":       time.Now(),
		},
	).Error; err != nil {
		return err
	}

	return nil
}

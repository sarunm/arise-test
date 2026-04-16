package transaction

import (
	"errors"
	"time"
)

type Transaction struct {
	ID            int       `gorm:"primaryKey"`
	FromAccountID int       `gorm:"index"`
	ToAccountID   int       `gorm:"index"`
	Amount        int64     `gorm:"not null"`
	Type          Type      `gorm:"type:transaction_type;not null"`
	CreatedAt     time.Time
}

type Type string

const (
	TypeDeposit  Type = "DEPOSIT"
	TypeWithdraw Type = "WITHDRAW"
	TypeTransfer Type = "TRANSFER"
)

var (
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrAccountNotActive    = errors.New("account is not active")
	ErrSameAccount         = errors.New("cannot transfer to same account")
)

type DepositRequest struct {
	AccountID int   `json:"account_id" binding:"required"`
	Amount    int64 `json:"amount"     binding:"required,gt=0"`
}

type WithdrawRequest struct {
	AccountID int   `json:"account_id" binding:"required"`
	Amount    int64 `json:"amount"     binding:"required,gt=0"`
}

type TransferRequest struct {
	FromAccountID int   `json:"from_account_id" binding:"required"`
	ToAccountID   int   `json:"to_account_id"   binding:"required"`
	Amount        int64 `json:"amount"          binding:"required,gt=0"`
}

type TransactionResponse struct {
	ID            int       `json:"id"`
	FromAccountID int       `json:"from_account_id,omitempty"`
	ToAccountID   int       `json:"to_account_id,omitempty"`
	Amount        float64   `json:"amount"`
	Type          Type      `json:"type"`
	CreatedAt     time.Time `json:"created_at"`
}

func (t *Transaction) ToResponse() TransactionResponse {
	return TransactionResponse{
		ID:            t.ID,
		FromAccountID: t.FromAccountID,
		ToAccountID:   t.ToAccountID,
		Amount:        float64(t.Amount) / 100,
		Type:          t.Type,
		CreatedAt:     t.CreatedAt,
	}
}

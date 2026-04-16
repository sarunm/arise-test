package account

import (
	"errors"
	"time"
)

type Account struct {
	ID         int       `gorm:"column:id" gorm:"primaryKey"`
	CustomerID int       `gorm:"column:customer_id"`
	Number     string    `gorm:"column:number"`
	Balance    int64     `gorm:"column:balance"`
	Status     Status    `gorm:"column:status"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

type Status string

const (
	StatusActive   Status = "ACTIVE"
	StatusInactive Status = "INACTIVE"
	StatusFrozen   Status = "FROZEN"
)

var (
	ErrAccountNotFound = errors.New("account not found")
)

type CreateAccountRequest struct {
	CustomerID int `json:"customer_id"`
}

type AccountResponse struct {
	ID         int       `json:"id"`
	CustomerID int       `json:"customer_id"`
	Number     string    `json:"number"`
	Balance    float64   `json:"balance"`
	Status     Status    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

func (a *Account) ToResponse() AccountResponse {
	return AccountResponse{
		ID:         a.ID,
		CustomerID: a.CustomerID,
		Number:     a.Number,
		Balance:    float64(a.Balance) / 100,
		Status:     a.Status,
		CreatedAt:  a.CreatedAt,
	}
}

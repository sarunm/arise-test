package account

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type repo struct {
	db *gorm.DB
}

type Repo interface {
	GetByID(ctx context.Context, id int) (*Account, error)
	GetByCustomerID(ctx context.Context, customerID int) ([]Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (*Account, error)
	Create(ctx context.Context, account Account) error
	Update(ctx context.Context, account Account) error
}

func newRepo(db *gorm.DB) Repo {
	return &repo{db: db}

}

func (r *repo) GetByID(ctx context.Context, id int) (*Account, error) {
	var account Account

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return &account, nil
}

func (r *repo) GetByCustomerID(ctx context.Context, customerID int) ([]Account, error) {
	var accounts []Account
	err := r.db.WithContext(ctx).Where("customer_id = ?", customerID).Find(&accounts).Error
	return accounts, err
}

func (r *repo) GetByAccountNumber(ctx context.Context, accountNumber string) (*Account, error) {
	var account Account
	if err := r.db.WithContext(ctx).Where("account_number = ?", accountNumber).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAccountNotFound
		}
		return nil, err
	}
	return &account, nil
}

func (r *repo) Create(ctx context.Context, account Account) error {
	return r.db.WithContext(ctx).Create(&account).Error
}

func (r *repo) Update(ctx context.Context, account Account) error {
	return r.db.WithContext(ctx).Save(&account).Error
}

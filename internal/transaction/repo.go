package transaction

import (
	"context"
	"errors"

	"github.com/sarunm/arise-test/internal/account"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo interface {
	Deposit(ctx context.Context, req DepositRequest) (*Transaction, error)
	Withdraw(ctx context.Context, req WithdrawRequest) (*Transaction, error)
	Transfer(ctx context.Context, req TransferRequest) (*Transaction, error)
	GetByAccountID(ctx context.Context, accountID int) ([]Transaction, error)
}

type repo struct {
	db *gorm.DB
}

func newRepo(db *gorm.DB) Repo {
	return &repo{db: db}
}

func (r *repo) Deposit(ctx context.Context, req DepositRequest) (*Transaction, error) {
	var tx Transaction

	err := r.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		var acc account.Account
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&acc, req.AccountID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return account.ErrAccountNotFound
			}
			return err
		}

		if acc.Status != account.StatusActive {
			return ErrAccountNotActive
		}

		acc.Balance += req.Amount
		if err := db.Model(&acc).Update("balance", acc.Balance).Error; err != nil {
			return err
		}

		tx = Transaction{
			ToAccountID: req.AccountID,
			Amount:      req.Amount,
			Type:        TypeDeposit,
		}
		return db.Create(&tx).Error
	})

	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *repo) Withdraw(ctx context.Context, req WithdrawRequest) (*Transaction, error) {
	var tx Transaction

	err := r.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		var acc account.Account
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&acc, req.AccountID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return account.ErrAccountNotFound
			}
			return err
		}

		if acc.Status != account.StatusActive {
			return ErrAccountNotActive
		}
		if acc.Balance < req.Amount {
			return ErrInsufficientBalance
		}

		acc.Balance -= req.Amount
		if err := db.Model(&acc).Update("balance", acc.Balance).Error; err != nil {
			return err
		}

		tx = Transaction{
			FromAccountID: req.AccountID,
			Amount:        req.Amount,
			Type:          TypeWithdraw,
		}
		return db.Create(&tx).Error
	})

	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *repo) Transfer(ctx context.Context, req TransferRequest) (*Transaction, error) {
	var tx Transaction

	err := r.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		// lock ทีละ row ตาม id น้อย → มาก ป้องกัน deadlock
		firstID, secondID := req.FromAccountID, req.ToAccountID
		if firstID > secondID {
			firstID, secondID = secondID, firstID
		}

		var first, second account.Account
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&first, firstID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return account.ErrAccountNotFound
			}
			return err
		}
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&second, secondID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return account.ErrAccountNotFound
			}
			return err
		}

		accMap := map[int]*account.Account{
			first.ID:  &first,
			second.ID: &second,
		}
		from, to := accMap[req.FromAccountID], accMap[req.ToAccountID]
		if from.Status != account.StatusActive || to.Status != account.StatusActive {
			return ErrAccountNotActive
		}
		if from.Balance < req.Amount {
			return ErrInsufficientBalance
		}

		from.Balance -= req.Amount
		to.Balance += req.Amount

		if err := db.Model(from).Update("balance", from.Balance).Error; err != nil {
			return err
		}
		if err := db.Model(to).Update("balance", to.Balance).Error; err != nil {
			return err
		}

		tx = Transaction{
			FromAccountID: req.FromAccountID,
			ToAccountID:   req.ToAccountID,
			Amount:        req.Amount,
			Type:          TypeTransfer,
		}
		return db.Create(&tx).Error
	})

	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *repo) GetByAccountID(ctx context.Context, accountID int) ([]Transaction, error) {
	var txs []Transaction
	err := r.db.WithContext(ctx).
		Where("from_account_id = ? OR to_account_id = ?", accountID, accountID).
		Order("created_at DESC").
		Find(&txs).Error
	return txs, err
}

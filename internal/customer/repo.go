package customer

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type Repo interface {
	GetByID(ctx context.Context, id int) (*Customer, error)
	GetByEmail(ctx context.Context, email string) (*Customer, error)
	Create(ctx context.Context, customer Customer) error
	Update(ctx context.Context, customer Customer) error
}

type repo struct {
	db *gorm.DB
}

func newRepo(db *gorm.DB) Repo {
	return &repo{db: db}
}

func (r *repo) GetByID(ctx context.Context, id int) (*Customer, error) {
	var customer Customer
	if err := r.db.WithContext(ctx).First(&customer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	return &customer, nil
}

func (r *repo) GetByEmail(ctx context.Context, email string) (*Customer, error) {
	var customer Customer
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	return &customer, nil
}

func (r *repo) Create(ctx context.Context, customer Customer) error {
	return r.db.WithContext(ctx).Create(&customer).Error
}

func (r *repo) Update(ctx context.Context, customer Customer) error {
	return r.db.WithContext(ctx).Save(&customer).Error
}

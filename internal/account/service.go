package account

import (
	"context"
	"fmt"
	"math/rand"
)

type Service interface {
	GetByID(ctx context.Context, id int) (*AccountResponse, error)
	GetByCustomerID(ctx context.Context, customerID int) ([]AccountResponse, error)
	CreateAccount(ctx context.Context, req CreateAccountRequest) (*AccountResponse, error)
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, id int) (*AccountResponse, error) {
	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	response := account.ToResponse()
	return &response, nil
}

func (s *service) GetByCustomerID(ctx context.Context, customerID int) ([]AccountResponse, error) {
	account, err := s.repo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}
	response := make([]AccountResponse, len(account))
	for i, v := range account {
		response[i] = v.ToResponse()
	}
	return response, nil

}

func (s *service) CreateAccount(ctx context.Context, req CreateAccountRequest) (*AccountResponse, error) {
	account := &Account{
		CustomerID: req.CustomerID,
		Number:     generateAccountNumber(),
		Status:     StatusActive,
	}

	if err := s.repo.Create(ctx, account); err != nil {
		return nil, err
	}

	response := account.ToResponse()
	return &response, nil
}

func generateAccountNumber() string {
	return fmt.Sprintf("%010d", rand.Intn(9_000_000_000)+1_000_000_000)
}

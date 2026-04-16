package customer

import (
	"context"
	"errors"
)

type Service interface {
	GetByID(ctx context.Context, id int) (*CustomerResponse, error)
	CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*CustomerResponse, error)
	UpdateCustomer(ctx context.Context, id int, req UpdateCustomerRequest) (*CustomerResponse, error)
}

type service struct {
	repo Repo
}

func NewService(repo Repo) Service {
	return &service{repo: repo}
}

func (s *service) GetByID(ctx context.Context, id int) (*CustomerResponse, error) {
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	res := customer.ToResponse()
	return &res, nil
}

func (s *service) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*CustomerResponse, error) {
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, ErrCustomerNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	customer := Customer{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	}

	if err := s.repo.Create(ctx, customer); err != nil {
		return nil, err
	}

	res := customer.ToResponse()
	return &res, nil
}

func (s *service) UpdateCustomer(ctx context.Context, id int, req UpdateCustomerRequest) (*CustomerResponse, error) {
	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.FirstName != "" {
		customer.FirstName = req.FirstName
	}
	if req.LastName != "" {
		customer.LastName = req.LastName
	}

	if err := s.repo.Update(ctx, *customer); err != nil {
		return nil, err
	}

	res := customer.ToResponse()
	return &res, nil
}

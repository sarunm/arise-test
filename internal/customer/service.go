package customer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sarunm/arise-test/pkg/cache"
)

const (
	customerCacheTTL = 10 * time.Minute
	customerCacheKey = "customer:%d"
)

type Service interface {
	GetByID(ctx context.Context, id int) (*CustomerResponse, error)
	CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*CustomerResponse, error)
	UpdateCustomer(ctx context.Context, id int, req UpdateCustomerRequest) (*CustomerResponse, error)
}

type service struct {
	repo  Repo
	cache cache.Cache
}

func NewService(repo Repo, cache cache.Cache) Service {
	return &service{repo: repo, cache: cache}
}

func (s *service) GetByID(ctx context.Context, id int) (*CustomerResponse, error) {
	key := fmt.Sprintf(customerCacheKey, id)

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var response CustomerResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return &response, nil
		}
	}

	customer, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := customer.ToResponse()

	if data, err := json.Marshal(response); err == nil {
		s.cache.Set(ctx, key, data, customerCacheTTL)
	}

	return &response, nil
}

func (s *service) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*CustomerResponse, error) {
	existing, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, ErrCustomerNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrEmailAlreadyExists
	}

	customer := &Customer{
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

	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, err
	}

	s.cache.Del(ctx, fmt.Sprintf(customerCacheKey, id))

	res := customer.ToResponse()
	return &res, nil
}

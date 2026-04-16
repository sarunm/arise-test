package account

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/sarunm/arise-test/pkg/cache"
)

const (
	accountCacheTTL     = 5 * time.Minute
	accountCacheKey     = "account:%d"
	accountListCacheKey = "account:customer:%d"
	accountAllCacheKey  = "account:all"
)

type Service interface {
	GetAll(ctx context.Context) ([]AccountResponse, error)
	GetByID(ctx context.Context, id int) (*AccountResponse, error)
	GetByCustomerID(ctx context.Context, customerID int) ([]AccountResponse, error)
	CreateAccount(ctx context.Context, req CreateAccountRequest) (*AccountResponse, error)
	InvalidateCache(ctx context.Context, accountID int) error
}

type service struct {
	repo  Repo
	cache cache.Cache
}

func NewService(repo Repo, cache cache.Cache) Service {
	return &service{repo: repo, cache: cache}
}

func (s *service) GetAll(ctx context.Context) ([]AccountResponse, error) {
	if cached, err := s.cache.Get(ctx, accountAllCacheKey); err == nil {
		var responses []AccountResponse
		if err := json.Unmarshal([]byte(cached), &responses); err == nil {
			return responses, nil
		}
	}

	accounts, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]AccountResponse, len(accounts))
	for i, v := range accounts {
		responses[i] = v.ToResponse()
	}

	if data, err := json.Marshal(responses); err == nil {
		s.cache.Set(ctx, accountAllCacheKey, data, accountCacheTTL)
	}

	return responses, nil
}

func (s *service) GetByID(ctx context.Context, id int) (*AccountResponse, error) {
	key := fmt.Sprintf(accountCacheKey, id)

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var response AccountResponse
		if err := json.Unmarshal([]byte(cached), &response); err == nil {
			return &response, nil
		}
	}

	account, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := account.ToResponse()

	if data, err := json.Marshal(response); err == nil {
		s.cache.Set(ctx, key, data, accountCacheTTL)
	}

	return &response, nil
}

func (s *service) GetByCustomerID(ctx context.Context, customerID int) ([]AccountResponse, error) {
	key := fmt.Sprintf(accountListCacheKey, customerID)

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var responses []AccountResponse
		if err := json.Unmarshal([]byte(cached), &responses); err == nil {
			return responses, nil
		}
	}

	accounts, err := s.repo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	responses := make([]AccountResponse, len(accounts))
	for i, v := range accounts {
		responses[i] = v.ToResponse()
	}

	if data, err := json.Marshal(responses); err == nil {
		s.cache.Set(ctx, key, data, accountCacheTTL)
	}

	return responses, nil
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

	s.cache.Del(ctx, fmt.Sprintf(accountListCacheKey, req.CustomerID), accountAllCacheKey)

	response := account.ToResponse()
	return &response, nil
}

func (s *service) InvalidateCache(ctx context.Context, accountID int) error {
	return s.cache.Del(ctx, fmt.Sprintf(accountCacheKey, accountID))
}

func generateAccountNumber() string {
	return fmt.Sprintf("%010d", rand.Intn(9_000_000_000)+1_000_000_000)
}

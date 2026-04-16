package transaction

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/sarunm/arise-test/internal/account"
	"github.com/sarunm/arise-test/pkg/cache"
)

const (
	txCacheTTL = 5 * time.Minute
	txCacheKey = "transaction:account:%d"
)

type Service interface {
	Deposit(ctx context.Context, req DepositRequest) (*TransactionResponse, error)
	Withdraw(ctx context.Context, req WithdrawRequest) (*TransactionResponse, error)
	Transfer(ctx context.Context, req TransferRequest) (*TransactionResponse, error)
	GetByAccountID(ctx context.Context, accountID int) ([]TransactionResponse, error)
}

type service struct {
	repo       Repo
	cache      cache.Cache
	accountSvc account.Service
}

func NewService(repo Repo, cache cache.Cache, accountSvc account.Service) Service {
	return &service{repo: repo, cache: cache, accountSvc: accountSvc}
}

func (s *service) Deposit(ctx context.Context, req DepositRequest) (*TransactionResponse, error) {
	tx, err := s.repo.Deposit(ctx, req)
	if err != nil {
		return nil, err
	}

	s.cache.Del(ctx, fmt.Sprintf(txCacheKey, req.AccountID))
	s.accountSvc.InvalidateCache(ctx, req.AccountID)

	res := tx.ToResponse()
	return &res, nil
}

func (s *service) Withdraw(ctx context.Context, req WithdrawRequest) (*TransactionResponse, error) {
	tx, err := s.repo.Withdraw(ctx, req)
	if err != nil {
		return nil, err
	}

	s.cache.Del(ctx, fmt.Sprintf(txCacheKey, req.AccountID))
	s.accountSvc.InvalidateCache(ctx, req.AccountID)

	res := tx.ToResponse()
	return &res, nil
}

func (s *service) Transfer(ctx context.Context, req TransferRequest) (*TransactionResponse, error) {
	if req.FromAccountID == req.ToAccountID {
		return nil, ErrSameAccount
	}

	tx, err := s.repo.Transfer(ctx, req)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		s.cache.Del(ctx,
			fmt.Sprintf(txCacheKey, req.FromAccountID),
			fmt.Sprintf(txCacheKey, req.ToAccountID),
		)
	}()
	go func() {
		defer wg.Done()
		s.accountSvc.InvalidateCache(ctx, req.FromAccountID)
	}()
	go func() {
		defer wg.Done()
		s.accountSvc.InvalidateCache(ctx, req.ToAccountID)
	}()
	wg.Wait()

	res := tx.ToResponse()
	return &res, nil
}

func (s *service) GetByAccountID(ctx context.Context, accountID int) ([]TransactionResponse, error) {
	key := fmt.Sprintf(txCacheKey, accountID)

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var responses []TransactionResponse
		if err := json.Unmarshal([]byte(cached), &responses); err == nil {
			return responses, nil
		}
	}

	txs, err := s.repo.GetByAccountID(ctx, accountID)
	if err != nil {
		return nil, err
	}

	responses := make([]TransactionResponse, len(txs))
	for i, tx := range txs {
		responses[i] = tx.ToResponse()
	}

	if data, err := json.Marshal(responses); err == nil {
		s.cache.Set(ctx, key, data, txCacheTTL)
	}

	return responses, nil
}

package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sarunm/arise-test/internal/account"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) Deposit(ctx context.Context, req DepositRequest) (*Transaction, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *mockRepo) Withdraw(ctx context.Context, req WithdrawRequest) (*Transaction, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *mockRepo) Transfer(ctx context.Context, req TransferRequest) (*Transaction, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Transaction), args.Error(1)
}

func (m *mockRepo) GetByAccountID(ctx context.Context, accountID int) ([]Transaction, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).([]Transaction), args.Error(1)
}

type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.String(0), args.Error(1)
}

func (m *mockCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *mockCache) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

type mockAccountService struct {
	mock.Mock
}

func (m *mockAccountService) GetAll(ctx context.Context) ([]account.AccountResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).([]account.AccountResponse), args.Error(1)
}

func (m *mockAccountService) GetByID(ctx context.Context, id int) (*account.AccountResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.AccountResponse), args.Error(1)
}

func (m *mockAccountService) GetByCustomerID(ctx context.Context, customerID int) ([]account.AccountResponse, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]account.AccountResponse), args.Error(1)
}

func (m *mockAccountService) CreateAccount(ctx context.Context, req account.CreateAccountRequest) (*account.AccountResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*account.AccountResponse), args.Error(1)
}

func (m *mockAccountService) InvalidateCache(ctx context.Context, accountID int) error {
	args := m.Called(ctx, accountID)
	return args.Error(0)
}

func newTestService(repo Repo, c *mockCache, accountSvc *mockAccountService) Service {
	return &service{repo: repo, cache: c, accountSvc: accountSvc}
}

func stubTx(txType Type) *Transaction {
	return &Transaction{
		ID:     1,
		Amount: 100000,
		Type:   txType,
	}
}

func TestDeposit_Success(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)
	accountSvc := new(mockAccountService)

	req := DepositRequest{AccountID: 1, Amount: 100000}
	repo.On("Deposit", mock.Anything, req).Return(stubTx(TypeDeposit), nil)
	c.On("Del", mock.Anything, []string{"transaction:account:1"}).Return(nil)
	accountSvc.On("InvalidateCache", mock.Anything, 1).Return(nil)

	svc := newTestService(repo, c, accountSvc)
	res, err := svc.Deposit(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, TypeDeposit, res.Type)
	c.AssertCalled(t, "Del", mock.Anything, []string{"transaction:account:1"})
	accountSvc.AssertCalled(t, "InvalidateCache", mock.Anything, 1)
}

func TestDeposit_RepoError(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)
	accountSvc := new(mockAccountService)

	req := DepositRequest{AccountID: 1, Amount: 100000}
	repo.On("Deposit", mock.Anything, req).Return(nil, ErrAccountNotActive)

	svc := newTestService(repo, c, accountSvc)
	res, err := svc.Deposit(context.Background(), req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrAccountNotActive)
	c.AssertNotCalled(t, "Del")
	accountSvc.AssertNotCalled(t, "InvalidateCache")
}

func TestWithdraw_Success(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)
	accountSvc := new(mockAccountService)

	req := WithdrawRequest{AccountID: 1, Amount: 50000}
	repo.On("Withdraw", mock.Anything, req).Return(stubTx(TypeWithdraw), nil)
	c.On("Del", mock.Anything, []string{"transaction:account:1"}).Return(nil)
	accountSvc.On("InvalidateCache", mock.Anything, 1).Return(nil)

	svc := newTestService(repo, c, accountSvc)
	res, err := svc.Withdraw(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, TypeWithdraw, res.Type)
}

func TestWithdraw_InsufficientBalance(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)
	accountSvc := new(mockAccountService)

	req := WithdrawRequest{AccountID: 1, Amount: 999999999}
	repo.On("Withdraw", mock.Anything, req).Return(nil, ErrInsufficientBalance)

	svc := newTestService(repo, c, accountSvc)
	res, err := svc.Withdraw(context.Background(), req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrInsufficientBalance)
	c.AssertNotCalled(t, "Del")
}

func TestTransfer_Success(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)
	accountSvc := new(mockAccountService)

	req := TransferRequest{FromAccountID: 1, ToAccountID: 2, Amount: 10000}
	fromID, toID := 1, 2
	tx := &Transaction{ID: 1, FromAccountID: &fromID, ToAccountID: &toID, Amount: 10000, Type: TypeTransfer}

	repo.On("Transfer", mock.Anything, req).Return(tx, nil)
	c.On("Del", mock.Anything, []string{"transaction:account:1", "transaction:account:2"}).Return(nil)
	accountSvc.On("InvalidateCache", mock.Anything, 1).Return(nil)
	accountSvc.On("InvalidateCache", mock.Anything, 2).Return(nil)

	svc := newTestService(repo, c, accountSvc)
	res, err := svc.Transfer(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, TypeTransfer, res.Type)
	accountSvc.AssertCalled(t, "InvalidateCache", mock.Anything, 1)
	accountSvc.AssertCalled(t, "InvalidateCache", mock.Anything, 2)
}

func TestTransfer_SameAccount(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)
	accountSvc := new(mockAccountService)

	req := TransferRequest{FromAccountID: 1, ToAccountID: 1, Amount: 10000}

	svc := newTestService(repo, c, accountSvc)
	res, err := svc.Transfer(context.Background(), req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrSameAccount)
	repo.AssertNotCalled(t, "Transfer")
}

func TestTransfer_InsufficientBalance(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)
	accountSvc := new(mockAccountService)

	req := TransferRequest{FromAccountID: 1, ToAccountID: 2, Amount: 999999999}
	repo.On("Transfer", mock.Anything, req).Return(nil, ErrInsufficientBalance)

	svc := newTestService(repo, c, accountSvc)
	res, err := svc.Transfer(context.Background(), req)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrInsufficientBalance)
	c.AssertNotCalled(t, "Del")
}

func TestGetByAccountID_CacheHit(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)
	accountSvc := new(mockAccountService)

	cached := `[{"id":1,"amount":1000,"type":"DEPOSIT","created_at":"0001-01-01T00:00:00Z"}]`
	c.On("Get", mock.Anything, "transaction:account:1").Return(cached, nil)

	svc := newTestService(repo, c, accountSvc)
	res, err := svc.GetByAccountID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	repo.AssertNotCalled(t, "GetByAccountID")
}

func TestGetByAccountID_CacheMiss(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)
	accountSvc := new(mockAccountService)

	txs := []Transaction{*stubTx(TypeDeposit)}
	c.On("Get", mock.Anything, "transaction:account:1").Return("", errors.New("cache miss"))
	repo.On("GetByAccountID", mock.Anything, 1).Return(txs, nil)
	c.On("Set", mock.Anything, "transaction:account:1", mock.Anything, txCacheTTL).Return(nil)

	svc := newTestService(repo, c, accountSvc)
	res, err := svc.GetByAccountID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	c.AssertCalled(t, "Set", mock.Anything, "transaction:account:1", mock.Anything, txCacheTTL)
}

func TestGetByAccountID_RepoError(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)
	accountSvc := new(mockAccountService)

	c.On("Get", mock.Anything, "transaction:account:1").Return("", errors.New("cache miss"))
	repo.On("GetByAccountID", mock.Anything, 1).Return([]Transaction{}, errors.New("db error"))

	svc := newTestService(repo, c, accountSvc)
	res, err := svc.GetByAccountID(context.Background(), 1)

	assert.Nil(t, res)
	assert.Error(t, err)
}

package account

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) GetAll(ctx context.Context) ([]Account, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Account), args.Error(1)
}

func (m *mockRepo) GetByID(ctx context.Context, id int) (*Account, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Account), args.Error(1)
}

func (m *mockRepo) GetByCustomerID(ctx context.Context, customerID int) ([]Account, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]Account), args.Error(1)
}

func (m *mockRepo) GetByAccountNumber(ctx context.Context, number string) (*Account, error) {
	args := m.Called(ctx, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Account), args.Error(1)
}

func (m *mockRepo) Create(ctx context.Context, account *Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
}

func (m *mockRepo) Update(ctx context.Context, account *Account) error {
	args := m.Called(ctx, account)
	return args.Error(0)
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

func newTestService(repo Repo, cache *mockCache) Service {
	return &service{repo: repo, cache: cache}
}

func stubAccount() *Account {
	return &Account{
		ID:         1,
		CustomerID: 1,
		Number:     "1234567890",
		Balance:    100000,
		Status:     StatusActive,
	}
}

func TestGetByID_CacheHit(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	cached := `{"id":1,"customer_id":1,"number":"1234567890","balance":1000,"status":"ACTIVE","created_at":"0001-01-01T00:00:00Z"}`
	c.On("Get", mock.Anything, "account:1").Return(cached, nil)

	svc := newTestService(repo, c)
	res, err := svc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, res.ID)
	assert.Equal(t, float64(1000), res.Balance)
	repo.AssertNotCalled(t, "GetByID") // ไม่แตะ DB
}

func TestGetByID_CacheMiss(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	c.On("Get", mock.Anything, "account:1").Return("", errors.New("cache miss"))
	repo.On("GetByID", mock.Anything, 1).Return(stubAccount(), nil)
	c.On("Set", mock.Anything, "account:1", mock.Anything, accountCacheTTL).Return(nil)

	svc := newTestService(repo, c)
	res, err := svc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, res.ID)
	repo.AssertCalled(t, "GetByID", mock.Anything, 1)                                    // ต้องแตะ DB
	c.AssertCalled(t, "Set", mock.Anything, "account:1", mock.Anything, accountCacheTTL) // ต้อง save cache
}

func TestGetByID_NotFound(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	c.On("Get", mock.Anything, "account:99").Return("", errors.New("cache miss"))
	repo.On("GetByID", mock.Anything, 99).Return(nil, ErrAccountNotFound)

	svc := newTestService(repo, c)
	res, err := svc.GetByID(context.Background(), 99)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrAccountNotFound)
}

func TestCreateAccount_Success(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*account.Account")).
		Return(nil).
		Run(func(args mock.Arguments) {
			// จำลอง GORM assign ID กลับมา
			acc := args.Get(1).(*Account)
			acc.ID = 1
		})
	c.On("Del", mock.Anything, []string{"account:customer:1", "account:all"}).Return(nil)

	svc := newTestService(repo, c)
	res, err := svc.CreateAccount(context.Background(), CreateAccountRequest{CustomerID: 1})

	assert.NoError(t, err)
	assert.Equal(t, 1, res.ID)
	assert.Equal(t, StatusActive, res.Status)
	c.AssertCalled(t, "Del", mock.Anything, []string{"account:customer:1", "account:all"}) // invalidate list cache + all cache
}

func TestCreateAccount_RepoError(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	repo.On("Create", mock.Anything, mock.AnythingOfType("*account.Account")).
		Return(errors.New("db error"))

	svc := newTestService(repo, c)
	res, err := svc.CreateAccount(context.Background(), CreateAccountRequest{CustomerID: 1})

	assert.Nil(t, res)
	assert.Error(t, err)
	c.AssertNotCalled(t, "Del") // ไม่ถึง invalidate ถ้า create fail
}

func TestGetByCustomerID_CacheMiss(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	accounts := []Account{*stubAccount()}
	c.On("Get", mock.Anything, "account:customer:1").Return("", errors.New("cache miss"))
	repo.On("GetByCustomerID", mock.Anything, 1).Return(accounts, nil)
	c.On("Set", mock.Anything, "account:customer:1", mock.Anything, accountCacheTTL).Return(nil)

	svc := newTestService(repo, c)
	res, err := svc.GetByCustomerID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
}

func TestGetByCustomerID_CacheHit(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	cached := `[{"id":1,"customer_id":1,"number":"1234567890","balance":1000,"status":"ACTIVE","created_at":"0001-01-01T00:00:00Z"}]`
	c.On("Get", mock.Anything, "account:customer:1").Return(cached, nil)

	svc := newTestService(repo, c)
	res, err := svc.GetByCustomerID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	repo.AssertNotCalled(t, "GetByCustomerID")
}

func TestGetByCustomerID_RepoError(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	c.On("Get", mock.Anything, "account:customer:1").Return("", errors.New("cache miss"))
	repo.On("GetByCustomerID", mock.Anything, 1).Return([]Account{}, errors.New("db error"))

	svc := newTestService(repo, c)
	res, err := svc.GetByCustomerID(context.Background(), 1)

	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestGetAll_CacheHit(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	cached := `[{"id":1,"customer_id":1,"number":"1234567890","balance":1000,"status":"ACTIVE","created_at":"0001-01-01T00:00:00Z"}]`
	c.On("Get", mock.Anything, "account:all").Return(cached, nil)

	svc := newTestService(repo, c)
	res, err := svc.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	repo.AssertNotCalled(t, "GetAll")
}

func TestGetAll_CacheMiss(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	accounts := []Account{*stubAccount()}
	c.On("Get", mock.Anything, "account:all").Return("", errors.New("cache miss"))
	repo.On("GetAll", mock.Anything).Return(accounts, nil)
	c.On("Set", mock.Anything, "account:all", mock.Anything, accountCacheTTL).Return(nil)

	svc := newTestService(repo, c)
	res, err := svc.GetAll(context.Background())

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	repo.AssertCalled(t, "GetAll", mock.Anything)
	c.AssertCalled(t, "Set", mock.Anything, "account:all", mock.Anything, accountCacheTTL)
}

func TestGetAll_RepoError(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	c.On("Get", mock.Anything, "account:all").Return("", errors.New("cache miss"))
	repo.On("GetAll", mock.Anything).Return([]Account{}, errors.New("db error"))

	svc := newTestService(repo, c)
	res, err := svc.GetAll(context.Background())

	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestInvalidateCache(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	c.On("Del", mock.Anything, []string{"account:1"}).Return(nil)

	svc := newTestService(repo, c)
	err := svc.InvalidateCache(context.Background(), 1)

	assert.NoError(t, err)
	c.AssertCalled(t, "Del", mock.Anything, []string{"account:1"})
}

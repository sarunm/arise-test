package customer

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

func (m *mockRepo) GetByID(ctx context.Context, id int) (*Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Customer), args.Error(1)
}

func (m *mockRepo) GetByEmail(ctx context.Context, email string) (*Customer, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Customer), args.Error(1)
}

func (m *mockRepo) Create(ctx context.Context, customer *Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *mockRepo) Update(ctx context.Context, customer *Customer) error {
	args := m.Called(ctx, customer)
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

func stubCustomer() *Customer {
	return &Customer{
		ID:        1,
		FirstName: "สมชาย",
		LastName:  "ใจดี",
		Email:     "somchai@test.com",
	}
}

func TestGetByID_CacheHit(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	cached := `{"id":1,"first_name":"สมชาย","last_name":"ใจดี","email":"somchai@test.com","created_at":"0001-01-01T00:00:00Z"}`
	c.On("Get", mock.Anything, "customer:1").Return(cached, nil)

	svc := newTestService(repo, c)
	res, err := svc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, res.ID)
	assert.Equal(t, "สมชาย", res.FirstName)
	repo.AssertNotCalled(t, "GetByID") // ไม่แตะ DB
}

func TestGetByID_CacheMiss(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	c.On("Get", mock.Anything, "customer:1").Return("", errors.New("cache miss"))
	repo.On("GetByID", mock.Anything, 1).Return(stubCustomer(), nil)
	c.On("Set", mock.Anything, "customer:1", mock.Anything, customerCacheTTL).Return(nil)

	svc := newTestService(repo, c)
	res, err := svc.GetByID(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, 1, res.ID)
	repo.AssertCalled(t, "GetByID", mock.Anything, 1)
	c.AssertCalled(t, "Set", mock.Anything, "customer:1", mock.Anything, customerCacheTTL)
}

func TestGetByID_NotFound(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	c.On("Get", mock.Anything, "customer:99").Return("", errors.New("cache miss"))
	repo.On("GetByID", mock.Anything, 99).Return(nil, ErrCustomerNotFound)

	svc := newTestService(repo, c)
	res, err := svc.GetByID(context.Background(), 99)

	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrCustomerNotFound)
}

func TestCreateCustomer_Success(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	repo.On("GetByEmail", mock.Anything, "somchai@test.com").Return(nil, ErrCustomerNotFound)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*customer.Customer")).
		Return(nil).
		Run(func(args mock.Arguments) {
			customer := args.Get(1).(*Customer)
			customer.ID = 1
		})

	svc := newTestService(repo, c)
	res, err := svc.CreateCustomer(context.Background(), CreateCustomerRequest{
		FirstName: "สมชาย",
		LastName:  "ใจดี",
		Email:     "somchai@test.com",
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, res.ID)
	assert.Equal(t, "somchai@test.com", res.Email)
}

func TestCreateCustomer_EmailAlreadyExists(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	repo.On("GetByEmail", mock.Anything, "somchai@test.com").Return(stubCustomer(), nil)

	svc := newTestService(repo, c)
	res, err := svc.CreateCustomer(context.Background(), CreateCustomerRequest{
		FirstName: "อื่น",
		LastName:  "อื่น",
		Email:     "somchai@test.com",
	})

	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrEmailAlreadyExists)
	repo.AssertNotCalled(t, "Create") // ไม่ถึง create ถ้า email ซ้ำ
}

func TestCreateCustomer_RepoError(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	repo.On("GetByEmail", mock.Anything, "somchai@test.com").Return(nil, ErrCustomerNotFound)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*customer.Customer")).
		Return(errors.New("db error"))

	svc := newTestService(repo, c)
	res, err := svc.CreateCustomer(context.Background(), CreateCustomerRequest{
		FirstName: "สมชาย",
		LastName:  "ใจดี",
		Email:     "somchai@test.com",
	})

	assert.Nil(t, res)
	assert.Error(t, err)
}

func TestUpdateCustomer_Success(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	repo.On("GetByID", mock.Anything, 1).Return(stubCustomer(), nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*customer.Customer")).Return(nil)
	c.On("Del", mock.Anything, []string{"customer:1"}).Return(nil)

	svc := newTestService(repo, c)
	res, err := svc.UpdateCustomer(context.Background(), 1, UpdateCustomerRequest{
		FirstName: "สมชาย2",
	})

	assert.NoError(t, err)
	assert.Equal(t, "สมชาย2", res.FirstName)
	assert.Equal(t, "ใจดี", res.LastName) // LastName ไม่เปลี่ยน
	c.AssertCalled(t, "Del", mock.Anything, []string{"customer:1"})
}

func TestUpdateCustomer_NotFound(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	repo.On("GetByID", mock.Anything, 99).Return(nil, ErrCustomerNotFound)

	svc := newTestService(repo, c)
	res, err := svc.UpdateCustomer(context.Background(), 99, UpdateCustomerRequest{
		FirstName: "ใหม่",
	})

	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrCustomerNotFound)
	c.AssertNotCalled(t, "Del") // ไม่ invalidate ถ้า customer ไม่มี
}

func TestUpdateCustomer_PartialUpdate(t *testing.T) {
	repo := new(mockRepo)
	c := new(mockCache)

	repo.On("GetByID", mock.Anything, 1).Return(stubCustomer(), nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*customer.Customer")).Return(nil)
	c.On("Del", mock.Anything, []string{"customer:1"}).Return(nil)

	svc := newTestService(repo, c)
	// ส่งมาแค่ LastName
	res, err := svc.UpdateCustomer(context.Background(), 1, UpdateCustomerRequest{
		LastName: "มั่งมี",
	})

	assert.NoError(t, err)
	assert.Equal(t, "สมชาย", res.FirstName) // FirstName คงเดิม
	assert.Equal(t, "มั่งมี", res.LastName)
}

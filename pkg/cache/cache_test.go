package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newTestCache(t *testing.T) (Cache, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(mr.Close)

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return &redisCache{client: client}, mr
}

func TestSet_Get(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	err := c.Set(ctx, "key1", "value1", time.Minute)
	assert.NoError(t, err)

	val, err := c.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "value1", val)
}

func TestGet_NotFound(t *testing.T) {
	c, _ := newTestCache(t)

	val, err := c.Get(context.Background(), "not_exist")
	assert.Error(t, err)
	assert.Empty(t, val)
}

func TestSet_TTLExpired(t *testing.T) {
	c, mr := newTestCache(t)
	ctx := context.Background()

	c.Set(ctx, "key_ttl", "value", time.Second)

	mr.FastForward(2 * time.Second)

	val, err := c.Get(ctx, "key_ttl")
	assert.Error(t, err)
	assert.Empty(t, val)
}

func TestDel_SingleKey(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	c.Set(ctx, "key1", "value1", time.Minute)

	err := c.Del(ctx, "key1")
	assert.NoError(t, err)

	val, err := c.Get(ctx, "key1")
	assert.Error(t, err)
	assert.Empty(t, val)
}

func TestDel_MultipleKeys(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	c.Set(ctx, "key1", "value1", time.Minute)
	c.Set(ctx, "key2", "value2", time.Minute)

	err := c.Del(ctx, "key1", "key2")
	assert.NoError(t, err)

	_, err1 := c.Get(ctx, "key1")
	_, err2 := c.Get(ctx, "key2")
	assert.Error(t, err1)
	assert.Error(t, err2)
}

func TestDel_KeyNotExist(t *testing.T) {
	c, _ := newTestCache(t)

	err := c.Del(context.Background(), "not_exist")
	assert.NoError(t, err)
}

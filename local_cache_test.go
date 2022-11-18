package cache_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/longyue0521/cache"
	"github.com/stretchr/testify/assert"
)

func TestLocalCache(t *testing.T) {

	t.Run("Get", func(t *testing.T) {

		t.Run("key不存在", func(t *testing.T) {
			c := newLocalCache()
			_, err := c.Get(context.Background(), "whatever")
			assert.ErrorIs(t, err, cache.ErrKeyNotFound)
		})

		t.Run("key已存在", func(t *testing.T) {
			c := newLocalCache()
			key, value := "Key", "Value"
			assert.NoError(t, c.Set(context.Background(), key, value, time.Minute))
			val, err := c.Get(context.Background(), key)
			assert.NoError(t, err)
			assert.Equal(t, value, val)
		})

		t.Run("key已过期", func(t *testing.T) {
			c := newLocalCache()
			key, value, expiration := "Key1", "Value1", time.Microsecond
			assert.NoError(t, c.Set(context.Background(), key, value, expiration))
			<-time.After(expiration)
			_, err := c.Get(context.Background(), key)
			assert.ErrorIs(t, err, cache.ErrKeyNotFound)
		})

		t.Run("与Set并发", func(t *testing.T) {
			t.Skip("并发测试用例不稳定，需要手动运行多次才能出现期望的场景：key过期，多个Set与Get并发调用，Set将key成功续约，Get返回val且未过期")

			c := newLocalCache()
			key, value, expiration := "Key1", "Value1", time.Microsecond
			assert.NoError(t, c.Set(context.Background(), key, value, expiration))

			// 前置条件：key过期，用多个Set并发续约，Get返回能拿到key且未过期
			var wg sync.WaitGroup
			n := 5
			wg.Add(n)

			go func() {
				defer wg.Done()
				val, err := c.Get(context.Background(), key)
				assert.NoError(t, err)
				assert.Equal(t, value, val)
			}()

			for i := 0; i < n-1; i++ {
				go func(i int) {
					defer wg.Done()
					assert.NoError(t, c.Set(context.Background(), key, value, time.Duration(i)+1*expiration))
				}(i)
			}

			<-time.After(time.Millisecond * 10)
			wg.Wait()
		})
	})

	t.Run("Set", func(t *testing.T) {

		t.Run("多次Set，过期时间逐渐减小", func(t *testing.T) {
			c := newLocalCache()
			n := 10

			key, value, expiration := "Key2", "Value2", time.Millisecond

			for i := n; i > 0; i-- {
				assert.NoError(t, c.Set(context.Background(), key, value, time.Duration(i)*expiration))

				val, err := c.Get(context.Background(), key)
				assert.NoError(t, err)
				assert.Equal(t, value, val)
			}

			<-time.After(expiration)

			_, err := c.Get(context.Background(), key)
			assert.ErrorIs(t, err, cache.ErrKeyNotFound)
		})

		t.Run("多次Set，过期时间逐渐增大", func(t *testing.T) {
			c := newLocalCache()

			key, value, expiration := "Key2", "Value2", time.Millisecond
			assert.NoError(t, c.Set(context.Background(), key, value, expiration))

			n := 10
			for i := 2; i <= n; i++ {
				assert.NoError(t, c.Set(context.Background(), key, value, time.Duration(i)*expiration))

				val, err := c.Get(context.Background(), key)
				assert.NoError(t, err)
				assert.Equal(t, value, val)
			}

			<-time.After(time.Duration(n) * expiration)

			_, err := c.Get(context.Background(), key)
			assert.ErrorIs(t, err, cache.ErrKeyNotFound)
		})
	})

	t.Run("Del", func(t *testing.T) {

		t.Run("key不存在", func(t *testing.T) {
			c := newLocalCache()
			err := c.Del(context.Background(), "whatever")
			assert.NoError(t, err)
		})

		t.Run("key已存在", func(t *testing.T) {
			c := newLocalCache()
			key, value, expiration := "Key1", "Value1", time.Millisecond
			assert.NoError(t, c.Set(context.Background(), key, value, expiration))

			val, err := c.Get(context.Background(), key)
			assert.NoError(t, err)
			assert.Equal(t, value, val)

			err = c.Del(context.Background(), key)
			assert.NoError(t, err)

			_, err = c.Get(context.Background(), key)
			assert.ErrorIs(t, err, cache.ErrKeyNotFound)
		})

		t.Run("Key已过期", func(t *testing.T) {

			c := newLocalCache()
			key, value, expiration := "Key1", "Value1", time.Millisecond
			assert.NoError(t, c.Set(context.Background(), key, value, expiration))

			val, err := c.Get(context.Background(), key)
			assert.NoError(t, err)
			assert.Equal(t, value, val)

			<-time.After(expiration)

			err = c.Del(context.Background(), key)
			assert.NoError(t, err)

			_, err = c.Get(context.Background(), key)
			assert.ErrorIs(t, err, cache.ErrKeyNotFound)
		})
	})
}

func newLocalCache() cache.Cache {
	return cache.NewLocalCache()
}

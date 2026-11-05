package storage

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisStorageStrategyWhenRequestIsAllowed(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()
	defer clearRedis(rdb)

	storageStrategy := NewRedisStorageStrategy(rdb)

	allowed, err := storageStrategy.AllowRequest(context.Background(), "test-id", 5, time.Second, time.Second)
	assert.NoError(t, err)
	assert.True(t, allowed, "expected request to be allowed")

}

func TestRedisStorageStrategyWhenRequestIsNotAllowed(t *testing.T) {
	id := "test-id"
	requestLimit := 5
	windowDuration := 1 * time.Second
	blockDuration := 2 * time.Second

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()
	defer clearRedis(rdb)

	storageStrategy := NewRedisStorageStrategy(rdb)

	var wg sync.WaitGroup

	// Allow 5 requests
	for i := 0; i < 5; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			allowed, err := storageStrategy.AllowRequest(context.Background(), id, requestLimit, windowDuration, blockDuration)
			assert.NoError(t, err)
			assert.True(t, allowed, "expected request to be allowed")
		}()
	}

	wg.Wait()

	// 6th request should not be allowed
	allowed, err := storageStrategy.AllowRequest(context.Background(), id, requestLimit, windowDuration, blockDuration)
	assert.NoError(t, err)
	assert.False(t, allowed, "expected request to be not allowed")

	// Wait for block duration
	time.Sleep(blockDuration)

	// Request should be allowed after block duration
	allowed, err = storageStrategy.AllowRequest(context.Background(), id, requestLimit, windowDuration, blockDuration)
	assert.NoError(t, err)
	assert.True(t, allowed, "expected request to be allowed after block duration")
}

func TestRedisStorageStrategyWhenMultipleRequestsAreMadeConcurrently(t *testing.T) {
	id := "test-id"
	requestLimit := 5
	windowDuration := 1 * time.Second
	blockDuration := 2 * time.Second

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()
	defer clearRedis(rdb)

	var wg sync.WaitGroup

	totalRequests := 10
	chAllowed := make(chan bool, totalRequests)
	chErrors := make(chan error, totalRequests)

	storageStrategy := NewRedisStorageStrategy(rdb)

	// Make concurrent requests and collect results in channels
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			allowed, err := storageStrategy.AllowRequest(context.Background(), id, requestLimit, windowDuration, blockDuration)
			if err != nil {
				chErrors <- err
				return
			}

			chAllowed <- allowed
		}()
	}

	wg.Wait()
	close(chAllowed)
	close(chErrors)

	var allowedCount int
	var notAllowedCount int

	for err := range chErrors {
		require.NoError(t, err)
	}

	for allowed := range chAllowed {
		if allowed {
			allowedCount++
		} else {
			notAllowedCount++
		}
	}

	require.Equal(t, totalRequests, allowedCount+notAllowedCount, "expected total requests to match allowed and not allowed counts")
	assert.Equal(t, requestLimit, allowedCount, "expected allowed count to be equal to request limit")
	assert.Equal(t, totalRequests-requestLimit, notAllowedCount, "expected not allowed count to be equal to total requests minus request limit")
}

func clearRedis(rdb *redis.Client) {
	rdb.FlushDB(context.Background())
}

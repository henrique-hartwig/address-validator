package services

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupRedisContainer(t *testing.T) (*CacheService, func()) {
	ctx := context.Background()

	redisContainer, err := redis.Run(ctx,
		"redis:7-alpine",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
	)
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get Redis host: %v", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("Failed to get Redis port: %v", err)
	}

	cache, err := NewCacheService(host, port.Int(), "", 0, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to create cache service: %v", err)
	}

	cleanup := func() {
		cache.Close()
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}

	return cache, cleanup
}

func TestCacheSetAndGet(t *testing.T) {
	cache, cleanup := setupRedisContainer(t)
	defer cleanup()

	cache.Set("test_key", "test_value")

	value, found := cache.Get("test_key")
	if !found {
		t.Error("Expected to find cached value")
	}

	if value != "test_value" {
		t.Errorf("Expected 'test_value', got %v", value)
	}
}

func TestCacheExpiration(t *testing.T) {
	ctx := context.Background()

	redisContainer, err := redis.Run(ctx,
		"redis:7-alpine",
		redis.WithSnapshotting(10, 1),
		redis.WithLogLevel(redis.LogLevelVerbose),
	)
	if err != nil {
		t.Fatalf("Failed to start Redis container: %v", err)
	}
	defer func() {
		if err := testcontainers.TerminateContainer(redisContainer); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}()

	host, err := redisContainer.Host(ctx)
	if err != nil {
		t.Fatalf("Failed to get Redis host: %v", err)
	}

	port, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		t.Fatalf("Failed to get Redis port: %v", err)
	}

	shortTTLCache, err := NewCacheService(host, port.Int(), "", 0, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to create cache with short TTL: %v", err)
	}
	defer shortTTLCache.Close()

	shortTTLCache.Set("test_key", "test_value")

	_, found := shortTTLCache.Get("test_key")
	if !found {
		t.Error("Expected to find cached value immediately")
	}

	time.Sleep(200 * time.Millisecond)

	_, found = shortTTLCache.Get("test_key")
	if found {
		t.Error("Expected cache to have expired")
	}
}

func TestCacheDelete(t *testing.T) {
	cache, cleanup := setupRedisContainer(t)
	defer cleanup()

	cache.Set("test_key", "test_value")
	cache.Delete("test_key")

	_, found := cache.Get("test_key")
	if found {
		t.Error("Expected key to be deleted")
	}
}

func TestCacheFlush(t *testing.T) {
	cache, cleanup := setupRedisContainer(t)
	defer cleanup()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	if cache.ItemCount() != 2 {
		t.Errorf("Expected 2 items, got %d", cache.ItemCount())
	}

	cache.Flush()

	if cache.ItemCount() != 0 {
		t.Errorf("Expected 0 items after flush, got %d", cache.ItemCount())
	}
}

func TestCacheWithComplexData(t *testing.T) {
	cache, cleanup := setupRedisContainer(t)
	defer cleanup()

	data := map[string]interface{}{
		"name":  "John Doe",
		"age":   30,
		"email": "john@example.com",
		"addresses": []string{
			"123 Main St",
			"456 Oak Ave",
		},
	}

	cache.Set("user:123", data)

	retrieved, found := cache.Get("user:123")
	if !found {
		t.Fatal("Expected to find cached complex data")
	}

	retrievedMap, ok := retrieved.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map[string]interface{}, got %T", retrieved)
	}

	if retrievedMap["name"] != "John Doe" {
		t.Errorf("Expected name 'John Doe', got %v", retrievedMap["name"])
	}

	if retrievedMap["age"].(float64) != 30 {
		t.Errorf("Expected age 30, got %v", retrievedMap["age"])
	}
}

func TestCacheConnectionError(t *testing.T) {
	_, err := NewCacheService("invalid-host", 9999, "", 0, 1*time.Hour)
	if err == nil {
		t.Error("Expected error when connecting to invalid host")
	}
}

func TestCacheItemCount(t *testing.T) {
	cache, cleanup := setupRedisContainer(t)
	defer cleanup()

	if cache.ItemCount() != 0 {
		t.Errorf("Expected 0 items initially, got %d", cache.ItemCount())
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	if cache.ItemCount() != 3 {
		t.Errorf("Expected 3 items, got %d", cache.ItemCount())
	}

	cache.Delete("key2")

	if cache.ItemCount() != 2 {
		t.Errorf("Expected 2 items after delete, got %d", cache.ItemCount())
	}
}

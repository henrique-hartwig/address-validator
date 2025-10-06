package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCacheService(host string, port int, password string, db int, ttl time.Duration) (*CacheService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &CacheService{
		client: client,
		ttl:    ttl,
	}, nil
}

func (s *CacheService) Get(key string) (interface{}, bool) {
	ctx := context.Background()

	val, err := s.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, false
	}
	if err != nil {
		return nil, false
	}

	var result interface{}
	if err := json.Unmarshal([]byte(val), &result); err != nil {
		return val, true
	}

	return result, true
}

func (s *CacheService) Set(key string, value interface{}) {
	ctx := context.Background()

	data, err := json.Marshal(value)
	if err != nil {
		return
	}

	s.client.Set(ctx, key, data, s.ttl)
}

func (s *CacheService) Delete(key string) {
	ctx := context.Background()
	s.client.Del(ctx, key)
}

func (s *CacheService) Flush() {
	ctx := context.Background()
	s.client.FlushDB(ctx)
}

func (s *CacheService) ItemCount() int {
	ctx := context.Background()
	size, err := s.client.DBSize(ctx).Result()
	if err != nil {
		return 0
	}
	return int(size)
}

func (s *CacheService) Close() error {
	return s.client.Close()
}

package services

import (
	"sync"
	"time"
)

type MockCacheService struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

func NewMockCacheService() *MockCacheService {
	return &MockCacheService{
		data: make(map[string]interface{}),
	}
}

func (m *MockCacheService) Get(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

func (m *MockCacheService) Set(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

func (m *MockCacheService) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

func (m *MockCacheService) Flush() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]interface{})
}

func (m *MockCacheService) ItemCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

func (m *MockCacheService) Close() error {
	return nil
}

func NewMockCacheServiceWithTTL(ttl time.Duration) *MockCacheService {
	return NewMockCacheService()
}

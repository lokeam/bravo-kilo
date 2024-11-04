package services

import (
	"sync"
	"time"
)

type BookCacheMetrics struct {
    mu            sync.RWMutex
    Hits          int64
    Misses        int64
    Errors        int64
    UnmarshalErrs int64
    Operations    map[string]int64
    LastAccess    time.Time
}

func NewBookCacheMetrics() *BookCacheMetrics {
    return &BookCacheMetrics{
        Operations: make(map[string]int64),
    }
}

func (m *BookCacheMetrics) RecordCacheHit(key string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.Hits++
    m.Operations["hit"]++
    m.LastAccess = time.Now()
}

func (m *BookCacheMetrics) RecordCacheMiss(key string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.Misses++
    m.Operations["miss"]++
    m.LastAccess = time.Now()
}

func (m *BookCacheMetrics) RecordUnmarshalError(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UnmarshalErrs++
	m.Errors++
	m.Operations["unmarshal_error"]++
	m.LastAccess = time.Now()
}

package storage

import (
	"context"
	"sync"
	"time"
)

type FixedWindow struct {
	count       int
	windowStart time.Time
	blockUntil  *time.Time
}

type MemoryStorageStrategy struct {
	mu           sync.Mutex
	fixedWindows map[string]*FixedWindow
}

func NewMemoryStorageStrategy() *MemoryStorageStrategy {
	return &MemoryStorageStrategy{
		fixedWindows: make(map[string]*FixedWindow),
	}
}

func (ms *MemoryStorageStrategy) AllowRequest(ctx context.Context, id string, requestLimit int, windowDuration time.Duration, blockDuration time.Duration) (bool, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	fixedWindow, exists := ms.fixedWindows[id]
	if !exists {
		ms.fixedWindows[id] = &FixedWindow{
			count:       1,
			windowStart: time.Now(),
			blockUntil:  nil,
		}

		return true, nil
	}

	now := time.Now()

	if fixedWindow.blockUntil != nil && fixedWindow.blockUntil.After(now) {
		return false, nil
	}

	if fixedWindow.windowStart.Add(windowDuration).Before(now) {
		fixedWindow.count = 0
		fixedWindow.windowStart = now
		fixedWindow.blockUntil = nil
	}

	if fixedWindow.count >= requestLimit {
		blockUntil := now.Add(blockDuration)
		fixedWindow.blockUntil = &blockUntil
		return false, nil
	}

	fixedWindow.count++
	return true, nil
}

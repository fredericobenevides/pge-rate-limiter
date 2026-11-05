package storage

import (
	"context"
	"time"
)

type StorageStrategy interface {
	AllowRequest(ctx context.Context, id string, requestLimit int, windowDuration time.Duration, blockDuration time.Duration) (bool, error)
}

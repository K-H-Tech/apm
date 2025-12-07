package contract

import (
	"context"
	"time"
)

// Lock will wrap the locking mechanism methods and behavior.
type Lock interface {
	// Lock will acquire a lock.
	//
	// if it was successful it will return true and a function for releasing the lock after job is done.
	// 	example:
	// 		locked, releaseLock, err := lock.Lock(context.Background(), "user:10", time.Hour)
	// 		// handle error
	// 		defer releaseLock(context.Background())
	Lock(ctx context.Context, lockName string, ttl time.Duration) (bool, func(context.Context) error, error)
}

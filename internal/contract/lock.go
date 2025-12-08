package contract

import (
	"context"
	"time"
)

// Lock will wrap the locking mechanism methods and behavior.
type Lock interface {
	// Lock will acquire a lock.
	//
	// Returns:
	//   - bool: true if the lock was acquired, false otherwise
	//   - func: release function to unlock; nil if error != nil, no-op if bool == false
	//   - error: any error that occurred during lock acquisition
	//
	// Callers should check error first, then check bool before calling the release function.
	// If it was successful it will return true and a function for releasing the lock after job is done.
	//
	// Example:
	//
	//	locked, releaseLock, err := lock.Lock(context.Background(), "user:10", time.Hour)
	//	if err != nil {
	//	    return err
	//	}
	//	if locked {
	//	    defer releaseLock(context.Background())
	//	}
	Lock(ctx context.Context, lockName string, ttl time.Duration) (bool, func(context.Context) error, error)
}

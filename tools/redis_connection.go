package tools

import (
	"context"

	"github.com/K-H-Tech/apm/internal/config"
	"github.com/redis/go-redis/v9"
)

// ConnectToRedis will finalize a redis config
func ConnectToRedis(redisCfg config.Redis) (*redis.Client, error) {
	var rdb *redis.Client

	if redisCfg.SentinelEnabled {
		rdb = redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName:    redisCfg.MasterName,
			SentinelAddrs: redisCfg.SentinelAddrs,
			MaxRetries:    redisCfg.MaxRetries,
			Password:      redisCfg.Password,
			DB:            redisCfg.DB,
			PoolSize:      redisCfg.PoolSize,
			DialTimeout:   redisCfg.DialTimeout,
			ReadTimeout:   redisCfg.ReadTimeout,
			WriteTimeout:  redisCfg.WriteTimeout,
			PoolTimeout:   redisCfg.PoolTimeout,
		})
	} else {
		rdb = redis.NewClient(&redis.Options{
			Addr:         redisCfg.Addr,
			Password:     redisCfg.Password,
			DB:           redisCfg.DB,
			MaxRetries:   redisCfg.MaxRetries,
			PoolSize:     redisCfg.PoolSize,
			DialTimeout:  redisCfg.DialTimeout,
			ReadTimeout:  redisCfg.ReadTimeout,
			WriteTimeout: redisCfg.WriteTimeout,
			PoolTimeout:  redisCfg.PoolTimeout,
		})
	}

	// Use timeout for Ping to avoid hanging indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), redisCfg.DialTimeout)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}

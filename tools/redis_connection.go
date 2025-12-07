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
			Addr:     redisCfg.Addr,
			Password: redisCfg.Password,
			DB:       redisCfg.DB,
		})

	}

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}

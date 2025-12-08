# Redis (go-redis) Best Practices

## Overview

go-redis is the official Redis client for Go, offering a type-safe API with support for Redis Cluster, Sentinel, and all Redis data types.

## Installation

```bash
go get github.com/redis/go-redis/v9
```

## Client Initialization

### Basic Client

```go
import (
    "context"
    "time"
    "github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr:         "localhost:6379",
        Password:     "",           // No password
        DB:           0,            // Default DB

        // Connection pool settings
        PoolSize:     10,
        MinIdleConns: 5,
        PoolTimeout:  30 * time.Second,

        // Timeouts
        DialTimeout:  10 * time.Second,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,

        // Retry settings
        MaxRetries:      3,
        MinRetryBackoff: 8 * time.Millisecond,
        MaxRetryBackoff: 512 * time.Millisecond,
    })

    return rdb
}
```

### URL-Based Configuration

```go
func NewRedisClientFromURL(url string) (*redis.Client, error) {
    opts, err := redis.ParseURL(url)
    if err != nil {
        return nil, err
    }

    // Override specific options
    opts.PoolSize = 10
    opts.DialTimeout = 10 * time.Second

    return redis.NewClient(opts), nil
}

// Usage
// redis://user:password@localhost:6379/0?protocol=3
```

### Connection Verification

```go
func VerifyConnection(ctx context.Context, rdb *redis.Client) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    if err := rdb.Ping(ctx).Err(); err != nil {
        return fmt.Errorf("redis connection failed: %w", err)
    }
    return nil
}
```

## Connection Pool

### Pool Configuration

```go
rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",

    // Pool size = expected concurrent connections
    PoolSize:     10,                  // Max connections
    MinIdleConns: 5,                   // Keep warm connections
    PoolTimeout:  4 * time.Second,     // Wait for connection

    // Connection lifecycle
    ConnMaxIdleTime: 5 * time.Minute,  // Close idle connections
    ConnMaxLifetime: 0,                // No max lifetime (0 = forever)
})
```

### Pool Monitoring

```go
func LogPoolStats(rdb *redis.Client) {
    stats := rdb.PoolStats()
    log.Printf("Redis Pool Stats - Hits: %d, Misses: %d, Timeouts: %d, Total: %d, Idle: %d",
        stats.Hits,
        stats.Misses,
        stats.Timeouts,
        stats.TotalConns,
        stats.IdleConns,
    )
}
```

## Error Handling

### Common Error Patterns

```go
func GetValue(ctx context.Context, rdb *redis.Client, key string) (string, error) {
    val, err := rdb.Get(ctx, key).Result()

    // Key doesn't exist - not an error, just nil value
    if err == redis.Nil {
        return "", nil
    }

    // Actual error
    if err != nil {
        return "", fmt.Errorf("redis get failed: %w", err)
    }

    return val, nil
}
```

### Handling Specific Errors

```go
func HandleRedisError(err error) error {
    if err == nil {
        return nil
    }

    // Key not found
    if err == redis.Nil {
        return ErrNotFound
    }

    // Connection errors
    if errors.Is(err, context.DeadlineExceeded) {
        return ErrTimeout
    }

    // Check for network errors
    var netErr net.Error
    if errors.As(err, &netErr) {
        if netErr.Timeout() {
            return ErrTimeout
        }
        return ErrConnectionFailed
    }

    return err
}
```

## Common Operations

### String Operations

```go
// Set with expiration
err := rdb.Set(ctx, "key", "value", time.Hour).Err()

// Set only if not exists
wasSet, err := rdb.SetNX(ctx, "key", "value", time.Hour).Result()

// Set only if exists
wasSet, err := rdb.SetXX(ctx, "key", "value", time.Hour).Result()

// Get
val, err := rdb.Get(ctx, "key").Result()

// Get with type conversion
intVal, err := rdb.Get(ctx, "counter").Int()
```

### Hash Operations

```go
// Set hash fields
err := rdb.HSet(ctx, "user:1", map[string]interface{}{
    "name":  "John",
    "email": "john@example.com",
    "age":   30,
}).Err()

// Get single field
name, err := rdb.HGet(ctx, "user:1", "name").Result()

// Get all fields
user, err := rdb.HGetAll(ctx, "user:1").Result()

// Scan into struct
type User struct {
    Name  string `redis:"name"`
    Email string `redis:"email"`
    Age   int    `redis:"age"`
}

var user User
err := rdb.HGetAll(ctx, "user:1").Scan(&user)
```

### List Operations

```go
// Push to list
rdb.RPush(ctx, "queue", "item1", "item2")

// Pop from list (blocking)
result, err := rdb.BLPop(ctx, 0, "queue").Result()

// Get range
items, err := rdb.LRange(ctx, "list", 0, -1).Result()
```

### Set Operations

```go
// Add members
rdb.SAdd(ctx, "tags", "go", "redis", "backend")

// Check membership
isMember, err := rdb.SIsMember(ctx, "tags", "go").Result()

// Get all members
tags, err := rdb.SMembers(ctx, "tags").Result()
```

## Pipelining

### Basic Pipeline

```go
func GetMultipleKeys(ctx context.Context, rdb *redis.Client, keys []string) ([]string, error) {
    pipe := rdb.Pipeline()

    cmds := make([]*redis.StringCmd, len(keys))
    for i, key := range keys {
        cmds[i] = pipe.Get(ctx, key)
    }

    _, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return nil, err
    }

    results := make([]string, len(keys))
    for i, cmd := range cmds {
        val, err := cmd.Result()
        if err == redis.Nil {
            results[i] = ""
        } else if err != nil {
            return nil, err
        } else {
            results[i] = val
        }
    }

    return results, nil
}
```

### Transaction Pipeline

```go
func TransferPoints(ctx context.Context, rdb *redis.Client, from, to string, amount int) error {
    txf := func(tx *redis.Tx) error {
        // Get current values
        fromVal, err := tx.Get(ctx, from).Int()
        if err != nil && err != redis.Nil {
            return err
        }

        if fromVal < amount {
            return errors.New("insufficient points")
        }

        // Execute transaction
        _, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
            pipe.DecrBy(ctx, from, int64(amount))
            pipe.IncrBy(ctx, to, int64(amount))
            return nil
        })
        return err
    }

    // Retry on WATCH conflict
    for i := 0; i < 3; i++ {
        err := rdb.Watch(ctx, txf, from)
        if err == nil {
            return nil
        }
        if err == redis.TxFailedErr {
            continue
        }
        return err
    }

    return errors.New("transaction failed after retries")
}
```

## Cluster Configuration

```go
func NewClusterClient() *redis.ClusterClient {
    return redis.NewClusterClient(&redis.ClusterOptions{
        Addrs: []string{
            "localhost:7000",
            "localhost:7001",
            "localhost:7002",
        },

        // Load balancing
        RouteByLatency: true,
        RouteRandomly:  false,

        // Pool settings per node
        PoolSize:     10,
        MinIdleConns: 5,

        // Timeouts
        DialTimeout:  10 * time.Second,
        ReadTimeout:  30 * time.Second,
        WriteTimeout: 30 * time.Second,
    })
}

// Verify cluster
func VerifyCluster(ctx context.Context, rdb *redis.ClusterClient) error {
    return rdb.ForEachShard(ctx, func(ctx context.Context, shard *redis.Client) error {
        return shard.Ping(ctx).Err()
    })
}
```

## Sentinel (High Availability)

```go
func NewSentinelClient() *redis.Client {
    return redis.NewFailoverClient(&redis.FailoverOptions{
        MasterName:    "mymaster",
        SentinelAddrs: []string{
            "localhost:26379",
            "localhost:26380",
            "localhost:26381",
        },

        // Authentication
        Password:         "redis-password",
        SentinelPassword: "sentinel-password",

        DB:       0,
        PoolSize: 10,

        // Read from replicas
        RouteRandomly: true,
    })
}
```

## Pub/Sub

```go
func Subscribe(ctx context.Context, rdb *redis.Client, channel string) error {
    pubsub := rdb.Subscribe(ctx, channel)
    defer pubsub.Close()

    // Wait for subscription confirmation
    _, err := pubsub.Receive(ctx)
    if err != nil {
        return err
    }

    // Listen for messages
    ch := pubsub.Channel()
    for msg := range ch {
        log.Printf("Received: %s from %s", msg.Payload, msg.Channel)
    }

    return nil
}

func Publish(ctx context.Context, rdb *redis.Client, channel, message string) error {
    return rdb.Publish(ctx, channel, message).Err()
}
```

## Caching Patterns

### Cache-Aside Pattern

```go
func GetUserWithCache(ctx context.Context, rdb *redis.Client, db *sql.DB, userID string) (*User, error) {
    cacheKey := fmt.Sprintf("user:%s", userID)

    // Try cache first
    cached, err := rdb.Get(ctx, cacheKey).Bytes()
    if err == nil {
        var user User
        if err := json.Unmarshal(cached, &user); err == nil {
            return &user, nil
        }
    }

    // Cache miss - fetch from database
    user, err := getUserFromDB(db, userID)
    if err != nil {
        return nil, err
    }

    // Store in cache
    data, _ := json.Marshal(user)
    rdb.Set(ctx, cacheKey, data, time.Hour)

    return user, nil
}
```

### Cache Invalidation

```go
func InvalidateUserCache(ctx context.Context, rdb *redis.Client, userID string) error {
    keys := []string{
        fmt.Sprintf("user:%s", userID),
        fmt.Sprintf("user:%s:posts", userID),
        fmt.Sprintf("user:%s:followers", userID),
    }
    return rdb.Del(ctx, keys...).Err()
}
```

## Metrics & Monitoring

### Prometheus Integration

```go
import "github.com/redis/go-redis/extra/redisprometheus/v9"

func NewInstrumentedClient() *redis.Client {
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    collector := redisprometheus.NewCollector("myapp", "redis", rdb)
    prometheus.MustRegister(collector)

    return rdb
}
```

## Testing

### Mock Client

```go
import "github.com/go-redis/redismock/v9"

func TestGetUser(t *testing.T) {
    db, mock := redismock.NewClientMock()

    mock.ExpectGet("user:123").SetVal(`{"id":"123","name":"John"}`)

    result, err := db.Get(context.Background(), "user:123").Result()

    assert.NoError(t, err)
    assert.Contains(t, result, "John")
    assert.NoError(t, mock.ExpectationsWereMet())
}
```

## Anti-Patterns to Avoid

1. **Don't ignore redis.Nil** - It's not an error, it means key doesn't exist
2. **Don't create client per request** - Reuse the client (it's thread-safe)
3. **Don't use unbounded keys** - Set TTL or implement eviction
4. **Avoid large values** - Keep values under 100KB for best performance
5. **Don't block on BLPOP without timeout** - Always set reasonable timeouts
6. **Avoid KEYS command in production** - Use SCAN instead

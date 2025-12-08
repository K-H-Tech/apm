# ClickHouse Go Driver Best Practices

## Overview

ClickHouse-go is the official Go driver for ClickHouse, supporting both native protocol and HTTP interface. It's optimized for high-throughput data insertion and analytical queries.

## Installation

```bash
go get github.com/ClickHouse/clickhouse-go/v2
```

## Connection Setup

### Native Protocol (Recommended)

```go
import (
    "context"
    "time"
    "github.com/ClickHouse/clickhouse-go/v2"
)

func NewClickHouseConn() (clickhouse.Conn, error) {
    conn, err := clickhouse.Open(&clickhouse.Options{
        Addr: []string{"127.0.0.1:9000"},
        Auth: clickhouse.Auth{
            Database: "default",
            Username: "default",
            Password: "",
        },

        // Connection settings
        DialTimeout:      10 * time.Second,
        MaxOpenConns:     10,
        MaxIdleConns:     5,
        ConnMaxLifetime:  time.Hour,
        ConnOpenStrategy: clickhouse.ConnOpenInOrder,

        // Compression (recommended for large data)
        Compression: &clickhouse.Compression{
            Method: clickhouse.CompressionLZ4,
        },

        // Query settings
        Settings: clickhouse.Settings{
            "max_execution_time": 60,
        },

        // Client identification
        ClientInfo: clickhouse.ClientInfo{
            Products: []struct {
                Name    string
                Version string
            }{
                {Name: "my-app", Version: "1.0"},
            },
        },
    })

    if err != nil {
        return nil, err
    }

    // Verify connection
    if err := conn.Ping(context.Background()); err != nil {
        return nil, err
    }

    return conn, nil
}
```

### HTTP Protocol

```go
func NewClickHouseHTTPConn() (clickhouse.Conn, error) {
    return clickhouse.Open(&clickhouse.Options{
        Addr:     []string{"127.0.0.1:8123"},
        Protocol: clickhouse.HTTP,
        Auth: clickhouse.Auth{
            Database: "default",
            Username: "default",
            Password: "",
        },
        Settings: clickhouse.Settings{
            "max_execution_time": 60,
        },
        Compression: &clickhouse.Compression{
            Method: clickhouse.CompressionLZ4,
        },
    })
}
```

### database/sql Interface

```go
import (
    "database/sql"
    _ "github.com/ClickHouse/clickhouse-go/v2"
)

func NewClickHouseDB() (*sql.DB, error) {
    db := clickhouse.OpenDB(&clickhouse.Options{
        Addr: []string{"127.0.0.1:9000"},
        Auth: clickhouse.Auth{
            Database: "default",
            Username: "default",
            Password: "",
        },
        Settings: clickhouse.Settings{
            "max_execution_time": 60,
        },
    })

    // Configure pool
    db.SetMaxIdleConns(5)
    db.SetMaxOpenConns(10)
    db.SetConnMaxLifetime(time.Hour)

    return db, nil
}
```

### TLS Configuration

```go
func NewSecureConn() (clickhouse.Conn, error) {
    return clickhouse.Open(&clickhouse.Options{
        Addr: []string{"clickhouse.example.com:9440"},
        Auth: clickhouse.Auth{
            Database: "default",
            Username: "default",
            Password: "secret",
        },
        TLS: &tls.Config{
            InsecureSkipVerify: false,
            // Or load certificates
            // RootCAs: certPool,
        },
    })
}
```

## Batch Inserts

### Native Protocol Batch

```go
func BatchInsert(ctx context.Context, conn clickhouse.Conn, events []Event) error {
    batch, err := conn.PrepareBatch(ctx, `
        INSERT INTO events (
            timestamp, event_type, user_id, data
        )
    `)
    if err != nil {
        return err
    }

    for _, event := range events {
        err := batch.Append(
            event.Timestamp,
            event.Type,
            event.UserID,
            event.Data,
        )
        if err != nil {
            return err
        }
    }

    return batch.Send()
}
```

### Struct-Based Batch

```go
type Event struct {
    Timestamp time.Time `ch:"timestamp"`
    EventType string    `ch:"event_type"`
    UserID    uint64    `ch:"user_id"`
    Data      string    `ch:"data"`
}

func BatchInsertStructs(ctx context.Context, conn clickhouse.Conn, events []Event) error {
    batch, err := conn.PrepareBatch(ctx, `
        INSERT INTO events
    `)
    if err != nil {
        return err
    }

    for _, event := range events {
        if err := batch.AppendStruct(&event); err != nil {
            return err
        }
    }

    return batch.Send()
}
```

### Async Inserts

```go
func AsyncInsert(ctx context.Context, conn clickhouse.Conn, event Event) error {
    // Async insert - data buffered on server
    ctx = clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{
        "async_insert":          1,
        "wait_for_async_insert": 0,
    }))

    return conn.Exec(ctx, `
        INSERT INTO events (timestamp, event_type, user_id, data)
        VALUES (?, ?, ?, ?)
    `, event.Timestamp, event.Type, event.UserID, event.Data)
}
```

## Querying Data

### Basic Query

```go
func QueryEvents(ctx context.Context, conn clickhouse.Conn) ([]Event, error) {
    rows, err := conn.Query(ctx, `
        SELECT timestamp, event_type, user_id, data
        FROM events
        WHERE timestamp > ?
        ORDER BY timestamp DESC
        LIMIT 100
    `, time.Now().Add(-24*time.Hour))

    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var events []Event
    for rows.Next() {
        var e Event
        if err := rows.Scan(&e.Timestamp, &e.Type, &e.UserID, &e.Data); err != nil {
            return nil, err
        }
        events = append(events, e)
    }

    return events, rows.Err()
}
```

### Scan to Struct

```go
func QueryEventsToStruct(ctx context.Context, conn clickhouse.Conn) ([]Event, error) {
    rows, err := conn.Query(ctx, `
        SELECT timestamp, event_type, user_id, data
        FROM events
        LIMIT 100
    `)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var events []Event
    for rows.Next() {
        var e Event
        if err := rows.ScanStruct(&e); err != nil {
            return nil, err
        }
        events = append(events, e)
    }

    return events, nil
}
```

### Single Row Query

```go
func GetEventCount(ctx context.Context, conn clickhouse.Conn, eventType string) (uint64, error) {
    var count uint64
    err := conn.QueryRow(ctx, `
        SELECT count()
        FROM events
        WHERE event_type = ?
    `, eventType).Scan(&count)

    return count, err
}
```

### Named Parameters

```go
func QueryWithNamedParams(ctx context.Context, conn clickhouse.Conn) error {
    rows, err := conn.Query(ctx, `
        SELECT *
        FROM events
        WHERE event_type = @type
        AND timestamp > @start
    `, clickhouse.Named("type", "click"), clickhouse.Named("start", time.Now().Add(-24*time.Hour)))

    if err != nil {
        return err
    }
    defer rows.Close()

    // Process rows...
    return nil
}
```

## Data Types

### Common Type Mappings

```go
type AllTypes struct {
    // Integers
    Int8Val   int8   `ch:"int8_col"`
    Int64Val  int64  `ch:"int64_col"`
    UInt64Val uint64 `ch:"uint64_col"`

    // Floats
    Float32Val float32 `ch:"float32_col"`
    Float64Val float64 `ch:"float64_col"`

    // Strings
    StringVal     string `ch:"string_col"`
    FixedStr      string `ch:"fixed_string_col"`

    // Date/Time
    DateVal     time.Time `ch:"date_col"`
    DateTimeVal time.Time `ch:"datetime_col"`
    DateTime64  time.Time `ch:"datetime64_col"`

    // Complex types
    ArrayVal  []string          `ch:"array_col"`
    MapVal    map[string]string `ch:"map_col"`
    TupleVal  []interface{}     `ch:"tuple_col"`

    // Nullable
    NullableInt *int64 `ch:"nullable_int"`

    // UUID
    UUIDVal uuid.UUID `ch:"uuid_col"`

    // Decimal
    DecimalVal decimal.Decimal `ch:"decimal_col"`
}
```

### Working with Arrays

```go
// Insert array
batch.Append([]string{"tag1", "tag2", "tag3"})

// Query array
var tags []string
rows.Scan(&tags)
```

### Working with Maps

```go
// Insert map
batch.Append(map[string]int64{
    "clicks": 100,
    "views":  1000,
})

// Query map
var metrics map[string]int64
rows.Scan(&metrics)
```

## Connection Pooling

### Pool Configuration

```go
conn, err := clickhouse.Open(&clickhouse.Options{
    Addr: []string{"127.0.0.1:9000"},

    // Pool settings
    MaxOpenConns:     20,              // Max concurrent connections
    MaxIdleConns:     10,              // Keep connections warm
    ConnMaxLifetime:  time.Hour,       // Recycle connections

    // Connection strategy
    ConnOpenStrategy: clickhouse.ConnOpenInOrder,  // or ConnOpenRoundRobin

    // Block buffer for streaming
    BlockBufferSize: 10,
})
```

### Multiple Hosts (Failover)

```go
conn, err := clickhouse.Open(&clickhouse.Options{
    Addr: []string{
        "clickhouse-1:9000",
        "clickhouse-2:9000",
        "clickhouse-3:9000",
    },
    ConnOpenStrategy: clickhouse.ConnOpenRoundRobin,
})
```

## Error Handling

```go
func HandleClickHouseError(err error) {
    if err == nil {
        return
    }

    // Check for specific ClickHouse errors
    var exception *clickhouse.Exception
    if errors.As(err, &exception) {
        log.Printf("ClickHouse Exception [%d]: %s",
            exception.Code,
            exception.Message,
        )

        // Handle specific error codes
        switch exception.Code {
        case 60: // TABLE_ALREADY_EXISTS
            // Handle accordingly
        case 81: // DATABASE_NOT_FOUND
            // Handle accordingly
        }
    }
}
```

## Performance Optimization

### Batch Size

```go
const OptimalBatchSize = 100000

func BatchInsertOptimized(ctx context.Context, conn clickhouse.Conn, events []Event) error {
    for i := 0; i < len(events); i += OptimalBatchSize {
        end := i + OptimalBatchSize
        if end > len(events) {
            end = len(events)
        }

        if err := insertBatch(ctx, conn, events[i:end]); err != nil {
            return err
        }
    }
    return nil
}
```

### Compression Settings

```go
// LZ4 (default, fastest)
Compression: &clickhouse.Compression{
    Method: clickhouse.CompressionLZ4,
}

// ZSTD (better compression, slower)
Compression: &clickhouse.Compression{
    Method: clickhouse.CompressionZSTD,
    Level:  3,  // 1-22, higher = better compression
}
```

### Query Settings

```go
ctx = clickhouse.Context(ctx, clickhouse.WithSettings(clickhouse.Settings{
    "max_execution_time":     60,
    "max_memory_usage":       10000000000,  // 10GB
    "max_threads":            8,
    "prefer_localhost_replica": 1,
}))
```

## Testing

### Integration Test Setup

```go
func setupTestDB(t *testing.T) clickhouse.Conn {
    conn, err := clickhouse.Open(&clickhouse.Options{
        Addr: []string{"localhost:9000"},
        Auth: clickhouse.Auth{
            Database: "test",
        },
    })
    require.NoError(t, err)

    // Create test table
    err = conn.Exec(context.Background(), `
        CREATE TABLE IF NOT EXISTS test_events (
            timestamp DateTime,
            event_type String,
            user_id UInt64
        ) ENGINE = Memory
    `)
    require.NoError(t, err)

    t.Cleanup(func() {
        conn.Exec(context.Background(), "DROP TABLE IF EXISTS test_events")
        conn.Close()
    })

    return conn
}
```

## Anti-Patterns to Avoid

1. **Don't insert row-by-row** - Always use batch inserts
2. **Don't use small batch sizes** - Aim for 10k-100k rows per batch
3. **Avoid SELECT *** - Always specify columns explicitly
4. **Don't ignore compression** - Enable LZ4 at minimum
5. **Avoid frequent ALTER TABLE** - ClickHouse is optimized for append
6. **Don't use JOINs on large tables without proper keys**
7. **Avoid nullable columns when possible** - They add overhead

## Recommended Table Settings

```sql
CREATE TABLE events (
    timestamp DateTime,
    event_type LowCardinality(String),  -- For repeated values
    user_id UInt64,
    data String CODEC(ZSTD(3))          -- Compress large columns
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)        -- Monthly partitions
ORDER BY (event_type, timestamp)        -- Optimize for common queries
TTL timestamp + INTERVAL 90 DAY;        -- Auto-cleanup old data
```

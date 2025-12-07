# PostgreSQL (pgx) Best Practices

## Overview

pgx is a pure Go driver and toolkit for PostgreSQL, offering high performance and full PostgreSQL feature support. It's the recommended driver for Go applications using PostgreSQL.

## Installation

```bash
go get github.com/jackc/pgx/v5
go get github.com/jackc/pgx/v5/pgxpool
```

## Connection Setup

### Connection Pool (Recommended)

```go
import (
    "context"
    "os"
    "time"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context) (*pgxpool.Pool, error) {
    config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
    if err != nil {
        return nil, err
    }

    // Pool configuration
    config.MaxConns = 20
    config.MinConns = 5
    config.MaxConnLifetime = time.Hour
    config.MaxConnIdleTime = 30 * time.Minute
    config.HealthCheckPeriod = time.Minute

    // Connection configuration
    config.ConnConfig.ConnectTimeout = 5 * time.Second

    // After connect hook (session setup)
    config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
        // Set session parameters
        _, err := conn.Exec(ctx, "SET timezone = 'UTC'")
        return err
    }

    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, err
    }

    // Verify connection
    if err := pool.Ping(ctx); err != nil {
        return nil, err
    }

    return pool, nil
}
```

### Single Connection

```go
import "github.com/jackc/pgx/v5"

func NewConnection(ctx context.Context) (*pgx.Conn, error) {
    conn, err := pgx.Connect(ctx, os.Getenv("DATABASE_URL"))
    if err != nil {
        return nil, err
    }

    return conn, nil
}
```

### Connection String Formats

```go
// URL format
"postgres://user:password@localhost:5432/dbname?sslmode=disable"

// DSN format
"host=localhost port=5432 user=postgres password=secret dbname=mydb sslmode=disable"

// With options
"postgres://user:pass@host:5432/db?pool_max_conns=10&pool_min_conns=5"
```

## Query Execution

### Query Single Row

```go
func GetUserByID(ctx context.Context, pool *pgxpool.Pool, id int64) (*User, error) {
    var user User
    err := pool.QueryRow(ctx, `
        SELECT id, email, name, created_at
        FROM users
        WHERE id = $1
    `, id).Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)

    if err != nil {
        if errors.Is(err, pgx.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, err
    }

    return &user, nil
}
```

### Query Multiple Rows

```go
func GetUsers(ctx context.Context, pool *pgxpool.Pool, limit int) ([]User, error) {
    rows, err := pool.Query(ctx, `
        SELECT id, email, name, created_at
        FROM users
        ORDER BY created_at DESC
        LIMIT $1
    `, limit)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var users []User
    for rows.Next() {
        var user User
        err := rows.Scan(&user.ID, &user.Email, &user.Name, &user.CreatedAt)
        if err != nil {
            return nil, err
        }
        users = append(users, user)
    }

    if err := rows.Err(); err != nil {
        return nil, err
    }

    return users, nil
}
```

### Using pgx.CollectRows

```go
func GetUsersSimpler(ctx context.Context, pool *pgxpool.Pool) ([]User, error) {
    rows, err := pool.Query(ctx, `
        SELECT id, email, name, created_at
        FROM users
    `)
    if err != nil {
        return nil, err
    }

    return pgx.CollectRows(rows, pgx.RowToStructByName[User])
}
```

### Execute (Insert/Update/Delete)

```go
func CreateUser(ctx context.Context, pool *pgxpool.Pool, user *User) error {
    _, err := pool.Exec(ctx, `
        INSERT INTO users (email, name, created_at)
        VALUES ($1, $2, $3)
    `, user.Email, user.Name, time.Now())

    return err
}

func CreateUserReturning(ctx context.Context, pool *pgxpool.Pool, user *User) (int64, error) {
    var id int64
    err := pool.QueryRow(ctx, `
        INSERT INTO users (email, name, created_at)
        VALUES ($1, $2, $3)
        RETURNING id
    `, user.Email, user.Name, time.Now()).Scan(&id)

    return id, err
}
```

## Transactions

### Basic Transaction

```go
func TransferFunds(ctx context.Context, pool *pgxpool.Pool, from, to int64, amount float64) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx) // No-op if committed

    // Deduct from sender
    result, err := tx.Exec(ctx, `
        UPDATE accounts
        SET balance = balance - $1
        WHERE id = $2 AND balance >= $1
    `, amount, from)
    if err != nil {
        return err
    }

    if result.RowsAffected() == 0 {
        return errors.New("insufficient funds")
    }

    // Add to receiver
    _, err = tx.Exec(ctx, `
        UPDATE accounts
        SET balance = balance + $1
        WHERE id = $2
    `, amount, to)
    if err != nil {
        return err
    }

    return tx.Commit(ctx)
}
```

### Transaction with Options

```go
func TransactionWithIsolation(ctx context.Context, pool *pgxpool.Pool) error {
    tx, err := pool.BeginTx(ctx, pgx.TxOptions{
        IsoLevel:   pgx.Serializable,
        AccessMode: pgx.ReadWrite,
    })
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    // Execute queries...

    return tx.Commit(ctx)
}
```

### Transaction Helper

```go
func WithTransaction(ctx context.Context, pool *pgxpool.Pool, fn func(pgx.Tx) error) error {
    tx, err := pool.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    if err := fn(tx); err != nil {
        return err
    }

    return tx.Commit(ctx)
}

// Usage
err := WithTransaction(ctx, pool, func(tx pgx.Tx) error {
    // All operations in transaction
    return nil
})
```

## Bulk Operations

### COPY Protocol (Fastest)

```go
func BulkInsertUsers(ctx context.Context, pool *pgxpool.Pool, users []User) error {
    rows := make([][]interface{}, len(users))
    for i, u := range users {
        rows[i] = []interface{}{u.Email, u.Name, u.CreatedAt}
    }

    _, err := pool.CopyFrom(
        ctx,
        pgx.Identifier{"users"},
        []string{"email", "name", "created_at"},
        pgx.CopyFromRows(rows),
    )

    return err
}
```

### Batch Queries

```go
func BatchOperations(ctx context.Context, pool *pgxpool.Pool, users []User) error {
    batch := &pgx.Batch{}

    for _, user := range users {
        batch.Queue(`
            INSERT INTO users (email, name)
            VALUES ($1, $2)
        `, user.Email, user.Name)
    }

    br := pool.SendBatch(ctx, batch)
    defer br.Close()

    for range users {
        _, err := br.Exec()
        if err != nil {
            return err
        }
    }

    return nil
}
```

## Prepared Statements

### Named Prepared Statements

```go
func PreparedQuery(ctx context.Context, conn *pgx.Conn) error {
    // Prepare once
    _, err := conn.Prepare(ctx, "get_user", `
        SELECT id, email, name FROM users WHERE id = $1
    `)
    if err != nil {
        return err
    }

    // Execute many times
    for _, id := range userIDs {
        rows, err := conn.Query(ctx, "get_user", id)
        // Process rows...
    }

    return nil
}
```

## LISTEN/NOTIFY

### Pub/Sub with PostgreSQL

```go
func ListenForNotifications(ctx context.Context, pool *pgxpool.Pool, channel string) error {
    conn, err := pool.Acquire(ctx)
    if err != nil {
        return err
    }
    defer conn.Release()

    _, err = conn.Exec(ctx, "LISTEN "+channel)
    if err != nil {
        return err
    }

    for {
        notification, err := conn.Conn().WaitForNotification(ctx)
        if err != nil {
            return err
        }

        log.Printf("Received: channel=%s payload=%s",
            notification.Channel,
            notification.Payload,
        )
    }
}

func Notify(ctx context.Context, pool *pgxpool.Pool, channel, payload string) error {
    _, err := pool.Exec(ctx, "SELECT pg_notify($1, $2)", channel, payload)
    return err
}
```

## Custom Types

### UUID Support

```go
import (
    "github.com/google/uuid"
    pgxuuid "github.com/jackc/pgx-gofrs-uuid"
)

// Register in AfterConnect
config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
    pgxuuid.Register(conn.TypeMap())
    return nil
}
```

### JSON/JSONB

```go
type Metadata map[string]interface{}

func (m *Metadata) ScanPgx(src interface{}) error {
    return json.Unmarshal(src.([]byte), m)
}

// Or use pgtype.JSON
import "github.com/jackc/pgx/v5/pgtype"

var meta pgtype.JSON
rows.Scan(&meta)
```

### Arrays

```go
// Query array column
var tags []string
row.Scan(&tags)

// Insert array
pool.Exec(ctx, "INSERT INTO posts (tags) VALUES ($1)", []string{"go", "postgres"})
```

## Error Handling

### PostgreSQL Error Codes

```go
import "github.com/jackc/pgx/v5/pgconn"

func HandlePgError(err error) error {
    if err == nil {
        return nil
    }

    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        switch pgErr.Code {
        case "23505": // unique_violation
            return ErrDuplicateKey
        case "23503": // foreign_key_violation
            return ErrForeignKeyViolation
        case "23502": // not_null_violation
            return ErrNullConstraint
        case "22P02": // invalid_text_representation
            return ErrInvalidInput
        default:
            log.Printf("PostgreSQL Error [%s]: %s", pgErr.Code, pgErr.Message)
        }
    }

    if errors.Is(err, pgx.ErrNoRows) {
        return ErrNotFound
    }

    return err
}
```

## Pool Monitoring

```go
func LogPoolStats(pool *pgxpool.Pool) {
    stat := pool.Stat()
    log.Printf("Pool Stats - Total: %d, Acquired: %d, Idle: %d, MaxConns: %d",
        stat.TotalConns(),
        stat.AcquiredConns(),
        stat.IdleConns(),
        stat.MaxConns(),
    )
}
```

## Query Building

### Safe Query Building

```go
import "github.com/jackc/pgx/v5"

func SearchUsers(ctx context.Context, pool *pgxpool.Pool, filters map[string]interface{}) ([]User, error) {
    query := "SELECT id, email, name FROM users WHERE 1=1"
    args := pgx.NamedArgs{}

    if email, ok := filters["email"]; ok {
        query += " AND email = @email"
        args["email"] = email
    }

    if name, ok := filters["name"]; ok {
        query += " AND name ILIKE @name"
        args["name"] = "%" + name.(string) + "%"
    }

    rows, err := pool.Query(ctx, query, args)
    if err != nil {
        return nil, err
    }

    return pgx.CollectRows(rows, pgx.RowToStructByName[User])
}
```

## Testing

### Test Helpers

```go
func setupTestDB(t *testing.T) *pgxpool.Pool {
    pool, err := pgxpool.New(context.Background(), os.Getenv("TEST_DATABASE_URL"))
    require.NoError(t, err)

    t.Cleanup(func() {
        pool.Close()
    })

    return pool
}

func withTestTx(t *testing.T, pool *pgxpool.Pool, fn func(pgx.Tx)) {
    tx, err := pool.Begin(context.Background())
    require.NoError(t, err)
    defer tx.Rollback(context.Background())

    fn(tx)
    // Transaction automatically rolled back
}
```

### Mock with pgxmock

```go
import "github.com/pashagolub/pgxmock/v3"

func TestGetUser(t *testing.T) {
    mock, err := pgxmock.NewPool()
    require.NoError(t, err)
    defer mock.Close()

    rows := pgxmock.NewRows([]string{"id", "email", "name"}).
        AddRow(int64(1), "test@example.com", "Test User")

    mock.ExpectQuery("SELECT").
        WithArgs(int64(1)).
        WillReturnRows(rows)

    // Use mock as pool...
}
```

## Performance Tips

1. **Use Connection Pool** - Never create connections per request
2. **Use COPY for Bulk Inserts** - 10x faster than individual inserts
3. **Use Prepared Statements** - For frequently executed queries
4. **Batch Queries** - Reduce round trips
5. **Index Appropriately** - Profile slow queries
6. **Use EXPLAIN ANALYZE** - Understand query plans

## Anti-Patterns to Avoid

1. **Don't create pool per request** - Pool is thread-safe
2. **Don't ignore context cancellation** - Pass context through
3. **Don't use string concatenation for queries** - SQL injection risk
4. **Don't forget rows.Close()** - Leaks connections
5. **Don't ignore rows.Err()** - May contain errors
6. **Avoid `SELECT *`** - Specify columns explicitly

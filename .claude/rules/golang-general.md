# Go General Best Practices

## Overview

This document outlines general best practices for Go programming, covering project structure, error handling, concurrency, interface design, and testing.

## Project Structure

### Standard Layout

```
project/
├── cmd/                    # Main applications
│   └── myapp/
│       └── main.go
├── internal/               # Private code (not importable)
│   ├── domain/            # Business logic
│   ├── repository/        # Data access
│   └── service/           # Application services
├── pkg/                    # Public, reusable packages
├── api/                    # API definitions (OpenAPI, protobuf)
├── configs/               # Configuration files
├── scripts/               # Build and utility scripts
├── test/                  # Additional test data and helpers
├── go.mod
├── go.sum
└── README.md
```

### Package Naming

```go
// Good: Short, lowercase, no underscores
package user
package httputil
package testdata

// Bad: Avoid these patterns
package user_service    // No underscores
package UserService     // No camelCase
package util           // Too generic
```

## Error Handling

### Always Handle Errors

```go
// Good: Check and handle every error
result, err := doSomething()
if err != nil {
    return fmt.Errorf("doSomething failed: %w", err)
}

// Bad: Ignoring errors
result, _ := doSomething()  // Silent failure
```

### Error Wrapping

```go
// Good: Wrap errors with context
if err := db.Query(query); err != nil {
    return fmt.Errorf("querying users: %w", err)
}

// Check wrapped errors
if errors.Is(err, sql.ErrNoRows) {
    return nil, ErrNotFound
}

// Extract wrapped errors
var pgErr *pgconn.PgError
if errors.As(err, &pgErr) {
    log.Printf("PostgreSQL error code: %s", pgErr.Code)
}
```

### Custom Error Types

```go
// Define sentinel errors
var (
    ErrNotFound     = errors.New("resource not found")
    ErrUnauthorized = errors.New("unauthorized access")
)

// Custom error type with context
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error on %s: %s", e.Field, e.Message)
}
```

## Concurrency

### Goroutine Management

```go
// Good: Use errgroup for coordinated goroutines
func processItems(ctx context.Context, items []Item) error {
    g, ctx := errgroup.WithContext(ctx)

    for _, item := range items {
        item := item // Capture loop variable
        g.Go(func() error {
            return processItem(ctx, item)
        })
    }

    return g.Wait()
}
```

### Channel Patterns

```go
// Good: Always close channels from the sender side
func producer(ctx context.Context) <-chan int {
    ch := make(chan int)
    go func() {
        defer close(ch)
        for i := 0; ; i++ {
            select {
            case <-ctx.Done():
                return
            case ch <- i:
            }
        }
    }()
    return ch
}

// Good: Use buffered channels to prevent blocking
ch := make(chan Result, 100)
```

### Context Usage

```go
// Good: Pass context as first parameter
func DoWork(ctx context.Context, input string) (Result, error) {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return Result{}, ctx.Err()
    default:
    }

    // Use context for timeouts
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    return doExpensiveWork(ctx, input)
}
```

### Mutex Usage

```go
// Good: Keep mutex close to the data it protects
type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

// Good: Use RWMutex for read-heavy workloads
type Cache struct {
    mu    sync.RWMutex
    items map[string]Item
}

func (c *Cache) Get(key string) (Item, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item, ok := c.items[key]
    return item, ok
}
```

## Interface Design

### Small Interfaces

```go
// Good: Small, focused interfaces
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// Compose when needed
type ReadWriter interface {
    Reader
    Writer
}
```

### Accept Interfaces, Return Structs

```go
// Good: Accept interface, return concrete type
func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

// Interface defined by consumer
type Repository interface {
    Get(ctx context.Context, id string) (*Entity, error)
    Save(ctx context.Context, entity *Entity) error
}
```

## Struct Design

### Constructor Functions

```go
// Good: Use constructor functions
func NewServer(addr string, opts ...Option) *Server {
    s := &Server{
        addr:    addr,
        timeout: 30 * time.Second, // Default
    }
    for _, opt := range opts {
        opt(s)
    }
    return s
}

// Functional options pattern
type Option func(*Server)

func WithTimeout(d time.Duration) Option {
    return func(s *Server) {
        s.timeout = d
    }
}
```

### Struct Tags

```go
type User struct {
    ID        int64     `json:"id" db:"id"`
    Email     string    `json:"email" db:"email" validate:"required,email"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}
```

## Testing

### Table-Driven Tests

```go
func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive numbers", 1, 2, 3},
        {"negative numbers", -1, -2, -3},
        {"zero", 0, 0, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d",
                    tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

### Test Helpers

```go
// t.Helper() marks function as test helper
func assertNoError(t *testing.T, err error) {
    t.Helper()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func assertEqual[T comparable](t *testing.T, got, want T) {
    t.Helper()
    if got != want {
        t.Errorf("got %v, want %v", got, want)
    }
}
```

### Mocking with Interfaces

```go
// Define interface for dependency
type EmailSender interface {
    Send(to, subject, body string) error
}

// Production implementation
type SMTPSender struct { /* ... */ }

// Test mock
type MockSender struct {
    SendFunc func(to, subject, body string) error
}

func (m *MockSender) Send(to, subject, body string) error {
    return m.SendFunc(to, subject, body)
}
```

## Performance Tips

### Preallocate Slices

```go
// Good: Preallocate when size is known
items := make([]Item, 0, len(input))
for _, v := range input {
    items = append(items, transform(v))
}

// Bad: Repeated allocations
var items []Item
for _, v := range input {
    items = append(items, transform(v))
}
```

### String Building

```go
// Good: Use strings.Builder for concatenation
var b strings.Builder
for _, s := range parts {
    b.WriteString(s)
}
result := b.String()

// Bad: String concatenation in loop
result := ""
for _, s := range parts {
    result += s  // Creates new string each time
}
```

### Avoid Unnecessary Allocations

```go
// Good: Reuse byte slices
buf := make([]byte, 4096)
for {
    n, err := r.Read(buf)
    // process buf[:n]
}

// Good: Use sync.Pool for frequently allocated objects
var bufPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 4096)
    },
}

buf := bufPool.Get().([]byte)
defer bufPool.Put(buf)
```

## Code Style

### Naming Conventions

```go
// Exported names: CamelCase
func CalculateTotal() int { }
type UserService struct { }

// Unexported names: camelCase
func calculateSubtotal() int { }
type userRepository struct { }

// Acronyms: Keep consistent case
var httpClient *http.Client  // Not hTTPClient
type JSONParser struct { }    // Not JsonParser

// Interface names: -er suffix for single-method
type Reader interface { Read([]byte) (int, error) }
type Stringer interface { String() string }
```

### Comments

```go
// Package user provides user management functionality.
package user

// User represents a system user with authentication credentials.
type User struct {
    // ID is the unique identifier for the user.
    ID int64
    // Email is the user's email address, used for login.
    Email string
}

// NewUser creates a new user with the given email.
// It returns an error if the email is invalid.
func NewUser(email string) (*User, error) {
    // ...
}
```

## Anti-Patterns to Avoid

1. **Naked returns**: Always use explicit return values
2. **init() abuse**: Minimize use of init(), prefer explicit initialization
3. **Package-level state**: Avoid global mutable state
4. **Ignoring errors**: Always handle or explicitly ignore with `_ =`
5. **Deep nesting**: Refactor to reduce nesting, use early returns
6. **Large packages**: Split into smaller, focused packages
7. **Circular dependencies**: Use interfaces to break cycles

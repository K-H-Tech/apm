# Gin Web Framework Best Practices

## Overview

Gin is a high-performance HTTP web framework for Go, offering a Martini-like API with up to 40x better performance. It's ideal for building REST APIs and web applications.

## Installation

```bash
go get -u github.com/gin-gonic/gin
```

## Router Setup

### Basic vs Default Router

```go
// gin.Default() includes Logger and Recovery middleware
r := gin.Default()

// gin.New() creates blank router (production recommended)
r := gin.New()
r.Use(gin.Logger())
r.Use(gin.Recovery())
```

### Production Mode

```go
// Set before creating router
gin.SetMode(gin.ReleaseMode)

// Or via environment variable
// GIN_MODE=release
```

## Middleware Patterns

### Global Middleware

```go
func main() {
    r := gin.New()

    // Applied to all routes
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    r.Use(CORSMiddleware())
    r.Use(RequestIDMiddleware())

    r.Run(":8080")
}
```

### Group Middleware

```go
func main() {
    r := gin.Default()

    // Public routes
    public := r.Group("/")
    {
        public.GET("/health", healthCheck)
        public.POST("/login", login)
    }

    // Protected routes
    api := r.Group("/api")
    api.Use(AuthMiddleware())
    {
        api.GET("/users", getUsers)
        api.POST("/users", createUser)
    }

    // Admin routes with multiple middleware
    admin := r.Group("/admin")
    admin.Use(AuthMiddleware(), AdminMiddleware())
    {
        admin.DELETE("/users/:id", deleteUser)
    }
}
```

### Custom Middleware

```go
func RequestIDMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := c.GetHeader("X-Request-ID")
        if requestID == "" {
            requestID = uuid.New().String()
        }

        c.Set("request_id", requestID)
        c.Header("X-Request-ID", requestID)

        c.Next()
    }
}

func LoggingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        c.Next()

        latency := time.Since(start)
        status := c.Writer.Status()

        log.Printf("[%d] %s %s - %v",
            status, c.Request.Method, path, latency)
    }
}
```

## Route Handling

### Path Parameters

```go
// Single parameter
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
})

// Catch-all parameter
r.GET("/files/*filepath", func(c *gin.Context) {
    filepath := c.Param("filepath")
    c.String(200, "File: %s", filepath)
})

// Multiple parameters
r.GET("/users/:userID/posts/:postID", func(c *gin.Context) {
    userID := c.Param("userID")
    postID := c.Param("postID")
    // ...
})
```

### Query Parameters

```go
r.GET("/search", func(c *gin.Context) {
    query := c.Query("q")                    // Empty string if not present
    page := c.DefaultQuery("page", "1")      // With default
    limit, exists := c.GetQuery("limit")     // Check existence

    c.JSON(200, gin.H{
        "query": query,
        "page":  page,
        "limit": limit,
    })
})
```

### Request Binding

```go
type CreateUserRequest struct {
    Name     string `json:"name" binding:"required"`
    Email    string `json:"email" binding:"required,email"`
    Age      int    `json:"age" binding:"gte=0,lte=130"`
    Password string `json:"password" binding:"required,min=8"`
}

r.POST("/users", func(c *gin.Context) {
    var req CreateUserRequest

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{
            "error": "Validation failed",
            "details": err.Error(),
        })
        return
    }

    // Process valid request
    c.JSON(201, gin.H{"message": "User created"})
})
```

### Multiple Binding Types

```go
type QueryParams struct {
    Page  int `form:"page" binding:"gte=1"`
    Limit int `form:"limit" binding:"gte=1,lte=100"`
}

type PathParams struct {
    ID string `uri:"id" binding:"required,uuid"`
}

r.GET("/users/:id", func(c *gin.Context) {
    var path PathParams
    var query QueryParams

    if err := c.ShouldBindUri(&path); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    if err := c.ShouldBindQuery(&query); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // Use path.ID, query.Page, query.Limit
})
```

## Response Handling

### JSON Responses

```go
// Simple response
c.JSON(200, gin.H{
    "status": "success",
    "data":   user,
})

// With struct
type Response struct {
    Status  string      `json:"status"`
    Data    interface{} `json:"data,omitempty"`
    Error   string      `json:"error,omitempty"`
}

c.JSON(200, Response{
    Status: "success",
    Data:   users,
})
```

### Error Responses

```go
// Standard error response
func ErrorResponse(c *gin.Context, code int, message string) {
    c.JSON(code, gin.H{
        "status":  "error",
        "code":    code,
        "message": message,
    })
}

// Usage
r.GET("/users/:id", func(c *gin.Context) {
    user, err := getUserByID(c.Param("id"))
    if err != nil {
        if errors.Is(err, ErrNotFound) {
            ErrorResponse(c, 404, "User not found")
            return
        }
        ErrorResponse(c, 500, "Internal server error")
        return
    }
    c.JSON(200, user)
})
```

## Authentication

### JWT Middleware

```go
func JWTAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Missing authorization header"})
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid authorization format"})
            return
        }

        claims, err := ValidateToken(parts[1])
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
            return
        }

        c.Set("user_id", claims.UserID)
        c.Set("user_role", claims.Role)
        c.Next()
    }
}

// Usage in handler
func GetCurrentUser(c *gin.Context) {
    userID := c.GetString("user_id")
    // ...
}
```

### Basic Auth

```go
authorized := r.Group("/admin", gin.BasicAuth(gin.Accounts{
    "admin": "secret",
    "user":  "password",
}))

authorized.GET("/secrets", func(c *gin.Context) {
    user := c.MustGet(gin.AuthUserKey).(string)
    c.JSON(200, gin.H{"user": user})
})
```

## Error Handling

### Recovery Middleware

```go
r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
    if err, ok := recovered.(string); ok {
        c.JSON(500, gin.H{
            "error": err,
        })
    }
    c.AbortWithStatus(500)
}))
```

### Centralized Error Handler

```go
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func (e *AppError) Error() string {
    return e.Message
}

func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err

            if appErr, ok := err.(*AppError); ok {
                c.JSON(appErr.Code, appErr)
                return
            }

            c.JSON(500, gin.H{"error": "Internal server error"})
        }
    }
}
```

## Performance Tips

### Use Appropriate Buffer Sizes

```go
// For large file uploads
r.MaxMultipartMemory = 8 << 20 // 8 MiB
```

### Avoid Allocations in Hot Paths

```go
// Reuse gin.H for static responses
var successResponse = gin.H{"status": "ok"}

r.GET("/health", func(c *gin.Context) {
    c.JSON(200, successResponse)
})
```

### Structured Logging

```go
r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
    return fmt.Sprintf(`{"time":"%s","status":%d,"latency":"%s","client_ip":"%s","method":"%s","path":"%s"}` + "\n",
        param.TimeStamp.Format(time.RFC3339),
        param.StatusCode,
        param.Latency,
        param.ClientIP,
        param.Method,
        param.Path,
    )
}))
```

## Testing

### Handler Testing

```go
func TestGetUser(t *testing.T) {
    gin.SetMode(gin.TestMode)

    r := gin.New()
    r.GET("/users/:id", GetUser)

    req, _ := http.NewRequest("GET", "/users/123", nil)
    w := httptest.NewRecorder()

    r.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)

    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "123", response["id"])
}
```

### Testing with Middleware

```go
func TestAuthenticatedEndpoint(t *testing.T) {
    r := setupRouter()

    req, _ := http.NewRequest("GET", "/api/users", nil)
    req.Header.Set("Authorization", "Bearer valid-token")

    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
}
```

## Anti-Patterns to Avoid

1. **Don't use c.JSON after c.Abort()** - The response is already committed
2. **Don't share context across goroutines** - Context is not thread-safe
3. **Avoid global middleware for route-specific logic**
4. **Don't ignore binding errors** - Always validate and return appropriate errors
5. **Don't use Default() in production without customization**

## Recommended Project Structure

```
/api
├── handlers/
│   ├── user_handler.go
│   └── product_handler.go
├── middleware/
│   ├── auth.go
│   ├── cors.go
│   └── logging.go
├── routes/
│   └── routes.go
├── validators/
│   └── custom_validators.go
└── main.go
```

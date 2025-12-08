# Viper Configuration Best Practices

## Overview

Viper is a complete configuration solution for Go applications. It supports reading from JSON, TOML, YAML, HCL, envfile, and Java properties files, as well as environment variables and command-line flags.

## Installation

```bash
go get github.com/spf13/viper
```

## Basic Setup

### Simple Configuration

```go
import "github.com/spf13/viper"

func initConfig() {
    // Set config file name (without extension)
    viper.SetConfigName("config")

    // Set config type
    viper.SetConfigType("yaml")

    // Add config paths (searched in order)
    viper.AddConfigPath(".")
    viper.AddConfigPath("./config")
    viper.AddConfigPath("$HOME/.myapp")
    viper.AddConfigPath("/etc/myapp")

    // Read config
    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); ok {
            // Config file not found, use defaults
            log.Println("No config file found, using defaults")
        } else {
            log.Fatalf("Error reading config: %v", err)
        }
    }
}
```

### Explicit Config File

```go
viper.SetConfigFile("/path/to/config.yaml")

if err := viper.ReadInConfig(); err != nil {
    log.Fatalf("Error reading config: %v", err)
}
```

## Setting Defaults

```go
func setDefaults() {
    // Simple values
    viper.SetDefault("server.port", 8080)
    viper.SetDefault("server.host", "localhost")
    viper.SetDefault("log.level", "info")

    // Duration
    viper.SetDefault("server.timeout", "30s")

    // Nested values
    viper.SetDefault("database.host", "localhost")
    viper.SetDefault("database.port", 5432)
    viper.SetDefault("database.pool.max", 10)
    viper.SetDefault("database.pool.min", 2)

    // Slice
    viper.SetDefault("allowed.origins", []string{"http://localhost:3000"})

    // Map
    viper.SetDefault("features", map[string]bool{
        "feature-a": true,
        "feature-b": false,
    })
}
```

## Reading Values

### Basic Value Types

```go
// String
host := viper.GetString("server.host")

// Integer
port := viper.GetInt("server.port")

// Boolean
debug := viper.GetBool("debug")

// Float
rate := viper.GetFloat64("rate.limit")

// Duration
timeout := viper.GetDuration("server.timeout")

// Time
startTime := viper.GetTime("maintenance.start")

// String slice
origins := viper.GetStringSlice("allowed.origins")

// String map
headers := viper.GetStringMapString("custom.headers")
```

### Checking Existence

```go
if viper.IsSet("database.password") {
    password := viper.GetString("database.password")
}
```

### Getting All Settings

```go
all := viper.AllSettings()
fmt.Printf("%+v\n", all)
```

## Environment Variables

### Basic Environment Binding

```go
func initEnv() {
    // Set prefix for all env vars (e.g., MYAPP_SERVER_PORT)
    viper.SetEnvPrefix("MYAPP")

    // Replace . with _ in env var names
    viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

    // Automatically read matching env vars
    viper.AutomaticEnv()
}

// With MYAPP_SERVER_PORT=3000
port := viper.GetInt("server.port") // Returns 3000
```

### Explicit Binding

```go
// Bind specific key to env var
viper.BindEnv("database.password", "DB_PASSWORD")

// Now viper.GetString("database.password") reads DB_PASSWORD
```

### Environment Variable Priority

```go
// Priority (highest to lowest):
// 1. Explicit Set()
// 2. Flags
// 3. Environment variables
// 4. Config file
// 5. Key/value store
// 6. Defaults
```

## Configuration Structs

### Unmarshaling to Struct

```go
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
    Host    string        `mapstructure:"host"`
    Port    int           `mapstructure:"port"`
    Timeout time.Duration `mapstructure:"timeout"`
}

type DatabaseConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
    Name     string `mapstructure:"name"`
    Pool     PoolConfig `mapstructure:"pool"`
}

type PoolConfig struct {
    MaxConns int `mapstructure:"max"`
    MinConns int `mapstructure:"min"`
}

type LogConfig struct {
    Level  string `mapstructure:"level"`
    Format string `mapstructure:"format"`
}

func LoadConfig() (*Config, error) {
    var cfg Config

    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("unmarshal config: %w", err)
    }

    return &cfg, nil
}
```

### Unmarshal Options

```go
// Custom decode hooks
cfg := Config{}
err := viper.Unmarshal(&cfg, viper.DecodeHook(
    mapstructure.ComposeDecodeHookFunc(
        mapstructure.StringToTimeDurationHookFunc(),
        mapstructure.StringToSliceHookFunc(","),
    ),
))
```

## Cobra Integration

### Binding Flags

```go
import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
    Use: "myapp",
    PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
        return initConfig()
    },
}

func init() {
    // Define flag
    rootCmd.PersistentFlags().String("config", "", "config file path")
    rootCmd.PersistentFlags().Int("port", 8080, "server port")

    // Bind flag to viper
    viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
    viper.BindPFlag("server.port", rootCmd.PersistentFlags().Lookup("port"))
}

func initConfig() error {
    cfgFile := viper.GetString("config")

    if cfgFile != "" {
        viper.SetConfigFile(cfgFile)
    } else {
        viper.SetConfigName("config")
        viper.AddConfigPath(".")
    }

    viper.AutomaticEnv()

    return viper.ReadInConfig()
}
```

## Configuration File Watching

### Live Reload

```go
import "github.com/fsnotify/fsnotify"

func watchConfig() {
    viper.OnConfigChange(func(e fsnotify.Event) {
        log.Printf("Config file changed: %s", e.Name)

        // Reload configuration
        var cfg Config
        if err := viper.Unmarshal(&cfg); err != nil {
            log.Printf("Error reloading config: %v", err)
            return
        }

        // Apply new configuration
        applyConfig(cfg)
    })

    viper.WatchConfig()
}
```

### Safe Reload Pattern

```go
type ConfigManager struct {
    mu  sync.RWMutex
    cfg *Config
}

func (m *ConfigManager) Get() *Config {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.cfg
}

func (m *ConfigManager) Reload() error {
    var newCfg Config
    if err := viper.Unmarshal(&newCfg); err != nil {
        return err
    }

    m.mu.Lock()
    defer m.mu.Unlock()
    m.cfg = &newCfg
    return nil
}
```

## Sub-Configurations

### Extracting Config Subsets

```go
// config.yaml:
// cache:
//   redis:
//     host: localhost
//     port: 6379
//   memory:
//     size: 1000

func NewRedisClient() *redis.Client {
    // Get sub-configuration
    redisCfg := viper.Sub("cache.redis")

    if redisCfg == nil {
        log.Fatal("Redis configuration not found")
    }

    return redis.NewClient(&redis.Options{
        Addr: fmt.Sprintf("%s:%d",
            redisCfg.GetString("host"),
            redisCfg.GetInt("port"),
        ),
    })
}
```

## Multiple Configuration Sources

### Merging Configs

```go
func loadConfigs() error {
    // Base config
    viper.SetConfigName("config")
    viper.AddConfigPath(".")
    if err := viper.ReadInConfig(); err != nil {
        return err
    }

    // Environment-specific config (merged)
    env := os.Getenv("APP_ENV")
    if env != "" {
        viper.SetConfigName("config." + env)
        if err := viper.MergeInConfig(); err != nil {
            if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
                return err
            }
        }
    }

    return nil
}
```

### Reading from io.Reader

```go
func loadFromBytes(data []byte) error {
    viper.SetConfigType("yaml")
    return viper.ReadConfig(bytes.NewReader(data))
}

// Useful for embedded configs
//go:embed config.yaml
var defaultConfig []byte

func init() {
    viper.SetConfigType("yaml")
    viper.ReadConfig(bytes.NewReader(defaultConfig))
}
```

## Multiple Viper Instances

### Isolated Configurations

```go
func NewDatabaseConfig() *viper.Viper {
    v := viper.New()
    v.SetConfigName("database")
    v.AddConfigPath("./config")
    v.SetEnvPrefix("DB")
    v.AutomaticEnv()

    v.SetDefault("host", "localhost")
    v.SetDefault("port", 5432)

    return v
}

func NewCacheConfig() *viper.Viper {
    v := viper.New()
    v.SetConfigName("cache")
    v.AddConfigPath("./config")
    v.SetEnvPrefix("CACHE")
    v.AutomaticEnv()

    v.SetDefault("host", "localhost")
    v.SetDefault("port", 6379)

    return v
}
```

### Custom Key Delimiter

```go
v := viper.NewWithOptions(viper.KeyDelimiter("::"))

v.Set("database::host", "localhost")
host := v.GetString("database::host")
```

## Remote Configuration

### etcd/Consul Support

```go
import _ "github.com/spf13/viper/remote"

func loadRemoteConfig() error {
    // etcd
    viper.AddRemoteProvider("etcd", "http://127.0.0.1:4001", "/config/myapp.json")
    viper.SetConfigType("json")

    if err := viper.ReadRemoteConfig(); err != nil {
        return err
    }

    // Watch for changes
    go func() {
        for {
            time.Sleep(5 * time.Second)
            if err := viper.WatchRemoteConfig(); err != nil {
                log.Printf("Error watching remote config: %v", err)
            }
        }
    }()

    return nil
}
```

## Validation

### Config Validation

```go
type Config struct {
    Server ServerConfig `mapstructure:"server"`
}

func (c *Config) Validate() error {
    if c.Server.Port < 1 || c.Server.Port > 65535 {
        return fmt.Errorf("invalid port: %d", c.Server.Port)
    }

    if c.Server.Host == "" {
        return errors.New("host is required")
    }

    return nil
}

func LoadConfig() (*Config, error) {
    var cfg Config

    if err := viper.Unmarshal(&cfg); err != nil {
        return nil, err
    }

    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    return &cfg, nil
}
```

## Testing

### Test Configuration

```go
func TestWithConfig(t *testing.T) {
    // Create isolated viper for testing
    v := viper.New()
    v.SetConfigType("yaml")

    testConfig := `
server:
  port: 9999
  host: testhost
`
    v.ReadConfig(strings.NewReader(testConfig))

    assert.Equal(t, 9999, v.GetInt("server.port"))
    assert.Equal(t, "testhost", v.GetString("server.host"))
}

func TestConfigFromEnv(t *testing.T) {
    v := viper.New()
    v.SetEnvPrefix("TEST")
    v.AutomaticEnv()

    t.Setenv("TEST_DATABASE_HOST", "testdb")

    assert.Equal(t, "testdb", v.GetString("database.host"))
}
```

## Best Practices

1. **Set Defaults First** - Always set sensible defaults before reading config
2. **Use Struct Unmarshaling** - Type-safe access to configuration
3. **Validate Configuration** - Check values after loading
4. **Use Environment Variables for Secrets** - Never commit secrets to config files
5. **Separate Configs by Environment** - Use config.dev.yaml, config.prod.yaml
6. **Document Configuration** - Keep a sample/template config file

## Anti-Patterns to Avoid

1. **Don't use global viper** in libraries - Accept config as parameters
2. **Don't read config in init()** - Explicit initialization is better
3. **Don't ignore errors** from ReadInConfig
4. **Avoid hardcoding config paths** - Use AddConfigPath for flexibility
5. **Don't mix viper.Get* with struct fields** - Choose one approach

## Example Config File

```yaml
# config.yaml
server:
  host: localhost
  port: 8080
  timeout: 30s

database:
  host: localhost
  port: 5432
  user: postgres
  name: myapp
  pool:
    max: 10
    min: 2

log:
  level: info
  format: json

features:
  new-ui: true
  beta-api: false
```

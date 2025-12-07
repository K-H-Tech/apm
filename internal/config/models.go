package config

import "time"

type (
	App struct {
		DebugMode bool `json:"DEBUG_MODE" yaml:"DEBUG_MODE"`
	}

	PostgreSQL struct {
		Username           string        `json:"USERNAME" yaml:"USERNAME"`
		Password           string        `json:"PASSWORD" yaml:"PASSWORD"`
		Port               string        `json:"PORT" yaml:"PORT"`
		Host               string        `json:"HOST" yaml:"HOST"`
		DBName             string        `json:"DB_NAME" yaml:"DB_NAME"`
		MaxOpenConnections int           `json:"MAX_OPEN_CONNECTIONS" yaml:"MAX_OPEN_CONNECTIONS"`
		MaxIdleConnections int           `json:"MAX_IDLE_CONNECTIONS" yaml:"MAX_IDLE_CONNECTIONS"`
		ConnMaxLifetime    time.Duration `json:"CONN_MAX_LIFETIME" yaml:"CONN_MAX_LIFETIME"`
	}

	MongoDB struct {
		Username string        `json:"USERNAME" yaml:"USERNAME"`
		Password string        `json:"PASSWORD" yaml:"PASSWORD"`
		Host     string        `json:"HOST" yaml:"HOST"`
		DBName   string        `json:"DB_NAME" yaml:"DB_NAME"`
		Timeout  time.Duration `json:"TIMEOUT" yaml:"TIMEOUT"`
		URI      string        `json:"URI" yaml:"URI"`
		Database string        `json:"DATABASE" yaml:"DATABASE"`
	}

	Redis struct {
		Addr               string        `json:"ADDRESS" yaml:"ADDRESS"`
		DB                 int           `json:"DB" yaml:"DB"`
		Password           string        `json:"PASSWORD" yaml:"PASSWORD"`
		PoolSize           int           `json:"POOL_SIZE" yaml:"POOL_SIZE"`
		MaxRetries         int           `json:"MAX_RETRIES" yaml:"MAX_RETRIES"`
		DialTimeout        time.Duration `json:"DIAL_TIMEOUT" yaml:"DIAL_TIMEOUT"`
		ReadTimeout        time.Duration `json:"READ_TIMEOUT" yaml:"READ_TIMEOUT"`
		WriteTimeout       time.Duration `json:"WRITE_TIMEOUT" yaml:"WRITE_TIMEOUT"`
		PoolTimeout        time.Duration `json:"POOL_TIMEOUT" yaml:"POOL_TIMEOUT"`
		IdleTimeout        time.Duration `json:"IDLE_TIMEOUT" yaml:"IDLE_TIMEOUT"`
		IdleCheckFrequency time.Duration `json:"IDLE_CHECK_FREQUENCY" yaml:"IDLE_CHECK_FREQUENCY"`

		// Redis Sentinel Options
		SentinelAddrs   []string `json:"SENTINEL_ADDRS" yaml:"SENTINEL_ADDRS"`
		MasterName      string   `json:"MASTER_NAME" yaml:"MASTER_NAME"`
		SentinelEnabled bool     `json:"SENTINEL_ENABLED" yaml:"SENTINEL_ENABLED"`
	}

	RabbitMQ struct {
		Addr     string `json:"ADDRESS" yaml:"ADDRESS"`
		Username string `json:"USERNAME" yaml:"USERNAME"`
		Password string `json:"PASSWORD" yaml:"PASSWORD"`
		Channel  string `json:"CHANNEL" yaml:"CHANNEL"`
	}

	Logger struct {
		Path            string `yaml:"PATH"`
		Level           int8   `yaml:"LEVEL"`
		RotationEnabled bool   `yaml:"ROTATION_ENABLED"`
		MaxSize         int    `yaml:"MAX_SIZE"`
		MaxBackups      int    `yaml:"MAX_BACKUPS"`
		Compress        bool   `yaml:"COMPRESS"`
		MaxAge          int    `yaml:"MAX_AGE"`
	}

	KaveNegar struct {
		BaseURL string        `yaml:"BASE_URL"`
		Timeout time.Duration `yaml:"TIMEOUT"`
		APIKey  string        `yaml:"API_KEY"`
	}

	Crypt struct {
		SecretKey string `yaml:"SECRET_KEY"`
	}

	JWT struct {
		SecretKey string `yaml:"SECRET_KEY"`
	}

	ProxyCMD struct {
		ListenPort uint32 `yaml:"LISTEN_PORT"`
	}

	Notifications struct {
		TelegramToken string `yaml:"TELEGRAM_TOKEN"`

		WhatsAppSID   string `yaml:"WHATSAPP_SID"`
		WhatsAppToken string `yaml:"WHATSAPP_TOKEN"`
		WhatsAppFrom  string `yaml:"WHATSAPP_FROM"`

		SMTPServer   string `yaml:"SMTP_SERVER"`
		SMTPPort     int    `yaml:"SMTP_PORT"`
		SMTPUsername string `yaml:"SMTP_USERNAME"`
		SMTPPassword string `yaml:"SMTP_PASSWORD"`

		IgapAPIKey string `yaml:"IGAP_API_KEY"`
		BaleAPIKey string `yaml:"BALE_API_KEY"`
	}

	Auth struct {
		TestModeEnabled            bool          `json:"TEST_MODE_ENABLED" yaml:"TEST_MODE_ENABLED"`
		TestModeOTP                string        `json:"TEST_MODE_OTP" yaml:"TEST_MODE_OTP"`
		OTPExpireTime              time.Duration `json:"OTP_EXPIRE_TIME" yaml:"OTP_EXPIRE_TIME"`
		OTPMaxRetries              int           `json:"OTP_MAX_RETRIES" yaml:"OTP_MAX_RETRIES"`
		OTPBlockTime               time.Duration `json:"OTP_BLOCK_TIME" yaml:"OTP_BLOCK_TIME"`
		OTPRateLimitWindowDuration time.Duration `json:"OTP_RATE_LIMIT_WINDOW_DURATION" yaml:"OTP_RATE_LIMIT_WINDOW_DURATION"`
		OTPRateLimitCount          int           `json:"OTP_RATE_LIMIT_COUNT" yaml:"OTP_RATE_LIMIT_COUNT"`
	}
)

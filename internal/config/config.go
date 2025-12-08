package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Movadian represents the Movadian API configuration
type Movadian struct {
	ClientID                            string `yaml:"client_id"`
	BaseURL                             string `yaml:"base_url"`
	RequestTimeout                      int    `yaml:"request_timeout"`
	GetTokenRetries                     uint   `yaml:"get_token_retries"`
	ServerPublicKeyAndIDRefreshDuration int    `yaml:"server_public_key_and_id_refresh_duration"`
	KeySize                             int    `yaml:"key_size"`
	EnableKeyRotation                   bool   `yaml:"enable_key_rotation"`
	KeyRotationInterval                 int    `yaml:"key_rotation_interval"`
	SecureKeyStorage                    bool   `yaml:"secure_key_storage"`
}

// Config will wrap the high level of configs
type Config struct {
	App           App           `yaml:"APP"`
	PostgreSQL    PostgreSQL    `yaml:"POSTGRESQL"`
	MongoDB       MongoDB       `yaml:"MONGO_DB"`
	Redis         Redis         `yaml:"REDIS"`
	Logger        Logger        `yaml:"LOGGER"`
	Crypt         Crypt         `yaml:"CRYPT"`
	JWT           JWT           `yaml:"JWT"`
	Notifications Notifications `yaml:"NOTIFICATIONS"`
	ProxyServer   ProxyCMD      `yaml:"PROXY_SERVER"`
	Auth          Auth          `yaml:"AUTH"`
	Movadian      Movadian      `yaml:"movadian"`
}

// Load will read the yaml file and convert it into a config struct
func Load(configPath string) (Config, error) {
	viper.SetEnvPrefix("TB")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetConfigFile(configPath)
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	cfg := Config{}

	err := viper.MergeInConfig()
	if err != nil {
		return cfg, fmt.Errorf("Error in reading config: %w", err)
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return cfg, fmt.Errorf("error while unmarshalling the config: %w", err)
	}

	return cfg, nil
}

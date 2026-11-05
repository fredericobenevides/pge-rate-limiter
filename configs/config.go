package configs

import (
	"errors"
	"os"
	"time"

	"github.com/spf13/viper"
)

type RateLimiterConfig struct {
	Port                string        `mapstructure:"PORT"`
	RedisHost           string        `mapstructure:"REDIS_HOST"`
	RedisStorageEnabled bool          `mapstructure:"REDIS_STORAGE_ENABLED"`
	WindowDuration      time.Duration `mapstructure:"WINDOW_DURATION"`
	RequestIPLimit      int           `mapstructure:"REQUEST_IP_LIMIT"`
	RequestTokenLimit   int           `mapstructure:"REQUEST_TOKEN_LIMIT"`
	BlockIPDuration     time.Duration `mapstructure:"BLOCK_IP_DURATION"`
	BlockTokenDuration  time.Duration `mapstructure:"BLOCK_TOKEN_DURATION"`
}

func LoadConfig(path string) (*RateLimiterConfig, error) {
	viper.AddConfigPath(path)
	viper.SetConfigFile(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	var cfg RateLimiterConfig

	err := viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

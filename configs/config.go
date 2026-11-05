package configs

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

type RateLimiterConfig struct {
	Port string `mapstructure:"PORT"`
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

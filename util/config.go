package util

import (
	"time"

	"github.com/spf13/viper"
)

// Config is the configuration for the application.
// It is populated by the environment variables or config file.
type Config struct {
	DBSource             string        `mapstructure:"DB_SOURCE" validate:"required"`
	ServerAddress        string        `mapstructure:"SERVER_ADDRESS" validate:"required"`
	TokenSymmetricKey    string        `mapstructure:"TOKEN_SYMMETRIC_KEY" validate:"required"`
	AccessTokenDuration  time.Duration `mapstructure:"ACCESS_TOKEN_DURATION" validate:"required"`
	TokenType            string        `mapstructure:"TOKEN_TYPE" validate:"required,oneof=paseto jwt"`
	RefreshTokenDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION" validate:"required"`
}

// LoadConfig loads the configuration from the file specified by the path.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}

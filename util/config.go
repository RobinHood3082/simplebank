package util

import "github.com/spf13/viper"

// Config is the configuration for the application.
// It is populated by the environment variables or config file.
type Config struct {
	DBSource      string `mapstructure:"DB_SOURCE" validate:"required"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS" validate:"required"`
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

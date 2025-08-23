package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// NewViper loads config from .env file
func NewViper() *viper.Viper {
	config := viper.New()

	config.SetConfigFile(".env")
	config.AddConfigPath("./../../")
	config.AddConfigPath("./")
	err := config.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %w \n", err))
	}

	return config
}

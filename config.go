package main

import (
	"github.com/spf13/viper"
)

func InitConfig() error {
	viper.SetConfigName("config")         // name of config file (without extension)
	viper.SetConfigType("yaml")           // REQUIRED if the config file does not have the extension in the name
	viper.AddConfigPath(CONFIG_BASE_PATH) // optionally look for config in the working directory
	err := viper.ReadInConfig()           // Find and read the config file
	if err != nil {                       // Handle errors reading the config file
		return err
	}
	return nil
}

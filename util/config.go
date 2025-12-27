package util

import "github.com/spf13/viper"

//Config holds application configuration values
type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

//LoadConfig reads configuration from file and environment var
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	//Read from environment variables
	viper.AutomaticEnv()

	//Load config file
	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	//Map config to struct
	err = viper.Unmarshal(&config)
	return
}

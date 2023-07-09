package util

import "github.com/spf13/viper"

// Config stores all configuration of the application.
// The values are read by viper from a config file or environment variables
type Config struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

// LoadConfig reads configuration from file or environment vairables.
func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app") // app.env
	viper.SetConfigType("env") // can also be json, xml etc.

	viper.AutomaticEnv() // override if env variables exist

	err = viper.ReadInConfig() // err is pre-defined return value
	if err != nil {
		return
	}

	// unmarshal the config values into target config struct/object
	err = viper.Unmarshal(&config) // config is pre-defined return value
	return
}

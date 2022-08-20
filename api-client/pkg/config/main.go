package config

import (
	"os"

	"github.com/spf13/viper"
)

type HostAlias struct {
	Name string
	IpAddress string `mapstructure:"ip_address"`
}

type Configuration struct {
	Address string
	Username string
	Password string
	Site string

	FixedOnly bool `mapstructure:"fixed_only"`
	LowercaseHostnames bool `mapstructure:"lowercase_hostnames"`

	HostAliases []HostAlias `mapstructure:"host_aliases"`
}

func InitConfig () *Configuration {
	v := viper.New()

	v.SetConfigName("api-client")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.SetEnvPrefix("API_CLIENT")
	v.AutomaticEnv()
	v.SetDefault("site", "default")
	v.SetDefault("fixed_only", true)
	v.SetDefault("lowercase_hostnames", true)

	configFile := os.Getenv("API_CLIENT_CONFIG_FILE")
	if configFile != "" {
		v.SetConfigFile(configFile)
	}

	if viperError := v.ReadInConfig(); viperError != nil {
		_, isFileNotFoundError := viperError.(viper.ConfigFileNotFoundError)
		if isFileNotFoundError && configFile != "" {
			panic(viperError)
		} else if !isFileNotFoundError {
			panic(viperError)
		}
	}

	config := &Configuration{}
	unmarshalError := v.Unmarshal(config)
	if unmarshalError != nil {
		panic(unmarshalError)
	}

	return config
}

package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Server ServerConfig `json:"server" mapstructure:"server"`
	RTU    RtuConfig    `json:"rtu" mapstructure:"rtu"`
	CRM    CrmConfig    `json:"amo_crm" mapstructure:"amo_crm"`
	Debug  bool         `json:"debug" mapstructure:"debug"`
}

type ServerConfig struct {
	ServerPort     int           `json:"port" mapstructure:"port"`
	RequestTimeout time.Duration `json:"request_timeout" mapstructure:"request_timeout"`
}

type RtuConfig struct {
	Host string `json:"host" mapstructure:"host"`
}

type CrmConfig struct {
	ServiceCode string `json:"service_code" mapstructure:"service_code"`
}

func LoadConfig() *Config {
	viper.SetConfigName("config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("read error config file: %s", err))
	}

	var config Config
	err = viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Errorf("unmarshal error config file: %s", err))
	}
	return &config
}

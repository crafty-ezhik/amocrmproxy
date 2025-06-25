package main

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type Config struct {
	Server ServerConfig `json:"server"`
	RTU    RtuConfig    `json:"rtu"`
	CRM    CrmConfig    `json:"amo_crm"`
	Debug  bool         `json:"debug"`
}

type ServerConfig struct {
	ServerPort     int           `json:"port"`
	RequestTimeout time.Duration `json:"request_timeout"`
}

type RtuConfig struct {
	Host string `json:"host"`
}

type CrmConfig struct {
	ServiceCode string `json:"service_code"`
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

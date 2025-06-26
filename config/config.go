package config

import (
	"fmt"
	"github.com/spf13/viper"
	"time"
)

// Config - структура содержащая конфигурацию приложения
type Config struct {
	Server ServerConfig `json:"server" mapstructure:"server"`
	RTU    RtuConfig    `json:"rtu" mapstructure:"rtu"`
	CRM    CrmConfig    `json:"amo_crm" mapstructure:"amo_crm"`
	Email  EmailConfig  `json:"email" mapstructure:"email"`
	Debug  bool         `json:"debug" mapstructure:"debug"`
}

type ServerConfig struct {
	Host           string        `json:"host" mapstructure:"host"`
	ServerPort     int           `json:"port" mapstructure:"port"`
	RequestTimeout time.Duration `json:"request_timeout" mapstructure:"request_timeout"`
}

type RtuConfig struct {
	Host          string `json:"host" mapstructure:"host"`
	CrmApiPort    int    `json:"crm_api_port" mapstructure:"crm_api_port"`
	ClientApiPort int    `json:"client_api_port" mapstructure:"client_api_port"`
	AdminApiPort  int    `json:"admin_api_port" mapstructure:"admin_api_port"`
}

type CrmConfig struct {
	ServiceCode string `json:"service_code" mapstructure:"service_code"`
}

type EmailConfig struct {
	Host      string `json:"host" mapstructure:"host"`
	Port      int    `json:"port" mapstructure:"port"`
	Login     string `json:"login" mapstructure:"login"`
	Password  string `json:"password" mapstructure:"password"`
	Recipient string `json:"recipient" mapstructure:"recipient"`
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

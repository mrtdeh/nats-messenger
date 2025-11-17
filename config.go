package main

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	DEBUG       = "ncconfig/debug"
	LOG_LEVEL   = "ncconfig/log_level"
	HTTP_HOST   = "ncconfig/http_host"
	HTTP_PORT   = "ncconfig/http_port"
	DC_NAME     = "ncconfig/dc_name"
	SERVER_NAME = "ncconfig/server_name"
	NATS_URL    = "ncconfig/nats_url"
)

type Config struct {
	Debug    bool   `mapstructure:"ncconfig/debug"`
	DC       string `mapstructure:"ncconfig/dc_name"`
	Name     string `mapstructure:"ncconfig/server_name"`
	HttpHost string `mapstructure:"ncconfig/http_host"`
	HttpPort string `mapstructure:"ncconfig/http_port"`
	NatsURL  string `mapstructure:"ncconfig/nats_url"`
}

func LoadConfiguration() Config {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Bind environment variables
	viper.BindEnv(DEBUG, "DEBUG")
	viper.BindEnv(LOG_LEVEL, "LOG_LEVEL")
	viper.BindEnv(HTTP_HOST, "HTTP_HOST")
	viper.BindEnv(HTTP_PORT, "HTTP_PORT")
	viper.BindEnv(DC_NAME, "DC")
	viper.BindEnv(SERVER_NAME, "NAME")
	viper.BindEnv(NATS_URL, "NATS_URL")

	// Set the default configuration
	viper.SetDefault(DEBUG, false)
	viper.SetDefault(LOG_LEVEL, "info")
	viper.SetDefault(HTTP_HOST, "localhost")
	viper.SetDefault(HTTP_PORT, 3080)
	viper.SetDefault(DC_NAME, "farzan")
	viper.SetDefault(SERVER_NAME, "local")
	viper.SetDefault(NATS_URL, "nats://localhost:4222")

	var cnf Config
	viper.Unmarshal(&cnf)

	return cnf

}

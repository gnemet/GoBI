package config

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Database   DatabaseConfig   `mapstructure:"database"`
	CursorPool CursorPoolConfig `mapstructure:"cursor_pool"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
	Schema   string `mapstructure:"schema"`
}

func (cfg *DatabaseConfig) GetConnectStr() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable options='-c search_path=%s'",
		cfg.Host, cfg.User, cfg.Password, cfg.Database, cfg.Port, cfg.Schema,
	)
}

type CursorPoolConfig struct {
	MaxConnections     int    `mapstructure:"max_connections"`
	IdleTimeout        string `mapstructure:"idle_timeout"`
	AbsoluteTimeout    string `mapstructure:"absolute_timeout"`
	PageSize           int    `mapstructure:"page_size"`
	AvailablePageSizes []int  `mapstructure:"available_page_sizes"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Config file not found, using defaults")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Defaults
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.Database.Host == "" {
		cfg.Database.Host = "localhost"
	}
	if cfg.Database.Port == "" {
		cfg.Database.Port = "5432"
	}
	if cfg.CursorPool.PageSize == 0 {
		cfg.CursorPool.PageSize = 10
	}

	return &cfg, nil
}

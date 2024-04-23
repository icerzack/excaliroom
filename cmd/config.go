package cmd

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Apps struct {
		LogLevel string `yaml:"log_level"`
		Rest     struct {
			Port int `yaml:"port"`
			JWT  struct {
				ValidationURL string `yaml:"validation_url"`
				HeaderName    string `yaml:"header_name"`
			} `yaml:"jwt"`
		} `yaml:"rest"`
	} `yaml:"apps"`
	Storage struct {
		Users struct {
			Type          string `yaml:"type"`
			RedisAddress  string `yaml:"redis_address"`
			RedisPassword string `yaml:"redis_password"`
			RedisDB       int    `yaml:"redis_db"`
		} `yaml:"users"`
		Rooms struct {
			Type          string `yaml:"type"`
			RedisAddress  string `yaml:"redis_address"`
			RedisPassword string `yaml:"redis_password"`
			RedisDB       int    `yaml:"redis_db"`
		} `yaml:"rooms"`
	} `yaml:"storage"`
}

func ParseConfig(path string, logger *zap.Logger) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		logger.Error("Failed to open config file", zap.Error(err))
		return nil, fmt.Errorf("error opening file %w", err)
	}
	defer file.Close()

	var config Config
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		logger.Error("Failed to decode config file", zap.Error(err))
		return nil, fmt.Errorf("error decoding file %w", err)
	}

	return &config, nil
}

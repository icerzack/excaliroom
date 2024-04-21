package cmd

import (
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"os"
)

type Config struct {
	Apps struct {
		LogLevel string `yaml:"log-level"`
		Rest     struct {
			Port int `yaml:"port"`
			JWT  struct {
				ValidationURL string `yaml:"validation-url"`
				HeaderName    string `yaml:"header-name"`
			} `yaml:"jwt"`
		} `yaml:"rest"`
	} `yaml:"apps"`
	Storage struct {
		Users struct {
			Type          string `yaml:"type"`
			RedisAddress  string `yaml:"redis-address"`
			RedisPassword string `yaml:"redis-password"`
			RedisDB       int    `yaml:"redis-db"`
		} `yaml:"users"`
		Rooms struct {
			Type          string `yaml:"type"`
			RedisAddress  string `yaml:"redis-address"`
			RedisPassword string `yaml:"redis-password"`
			RedisDB       int    `yaml:"redis-db"`
		} `yaml:"rooms"`
	} `yaml:"storage"`
}

func ParseConfig(path string, logger *zap.Logger) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		logger.Error("Failed to open config file", zap.Error(err))
		return nil, err
	}
	defer file.Close()

	var config Config
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		logger.Error("Failed to decode config file", zap.Error(err))
		return nil, err
	}

	return &config, nil
}

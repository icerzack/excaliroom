package cmd

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Apps struct {
		Rest struct {
			Port       int `yaml:"port"`
			Validation struct {
				JWTHeaderName      string `yaml:"jwt_header_name"`
				JWTValidationURL   string `yaml:"jwt_validation_url"`
				BoardValidationURL string `yaml:"board_validation_url"`
			} `yaml:"validation"`
		} `yaml:"rest"`
	} `yaml:"apps"`
	Logging struct {
		Level       string `yaml:"level"`
		WriteToFile bool   `yaml:"write_to_file"`
	} `yaml:"logging"`
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
	Cache struct {
		Type          string `yaml:"type"`
		TTL           int64  `yaml:"ttl"`
		RedisAddress  string `yaml:"redis_address"`
		RedisPassword string `yaml:"redis_password"`
		RedisDB       int    `yaml:"redis_db"`
	} `yaml:"cache"`
}

func ParseConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println("error opening config file Path:", path, err)
		return nil, fmt.Errorf("error opening file %w", err)
	}
	defer file.Close()

	var config Config
	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		fmt.Println("error decoding config file", err)
		return nil, fmt.Errorf("error decoding file %w", err)
	}

	return &config, nil
}

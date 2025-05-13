package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`
	Backends []string `yaml:"backends"`
	Ratelimit struct {
		DefaultCapacity int `yaml:"default_capacity"`
		DefaultRate     time.Duration `yaml:"default_rate"`
	} `yaml:"rate_limit"`
}

func InitConfig() (*Config, error) {
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config 
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
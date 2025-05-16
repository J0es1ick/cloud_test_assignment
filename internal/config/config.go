package config

import (
	"os"
	"sync"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port string `yaml:"port"`
	} `yaml:"server"`

	Database struct {
		Host    	   string `yaml:"host"`
		Port   	       string `yaml:"port"`
		User           string `yaml:"user"`
		Password       string `yaml:"password"`
		Name           string `yaml:"name"`
		SSLMode        string `yaml:"sslmode"`
		ConnectTimeout time.Duration `yaml:"connect_timeout"`
	}

	Backends []string `yaml:"backends"`

	Ratelimit struct {
		DefaultCapacity int `yaml:"default_capacity"`
		DefaultRate     time.Duration `yaml:"default_rate"`
	} `yaml:"rate_limit"`
}

var (
    currentConfig atomic.Value
    configMutex   sync.RWMutex
)

func InitConfig() (*Config, error) {
	data, err := os.ReadFile(os.Getenv("CONFIG_PATH"))
	if err != nil {
		return nil, err
	}

	var cfg Config 
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func ReloadConfig() error {
    newCfg, err := InitConfig()
    if err != nil {
        return err
    }
    
    configMutex.Lock()
    currentConfig.Store(newCfg)
    configMutex.Unlock()
    return nil
}

func GetConfig() *Config {
    return currentConfig.Load().(*Config)
}

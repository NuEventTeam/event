package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env      string   `yaml:"env"`
	Database Database `yaml:"database"`
	Cache    Cache    `yaml:"cache"`
	JWT      JWT      `yaml:"jwt"`
	Http     Http     `yaml:"http"`
}

type JWT struct {
	Secret string        `yaml:"secret"`
	Expiry time.Duration `yaml:"expiry"`
}

type GRPC struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}
type Http struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type Database struct {
	Host     string `yaml:"host"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	Port     int    `yaml:"port"`
}

type Cache struct {
	Port     string `yaml:"port"`
	Host     string `yaml:"host"`
	Password string `yaml:"password"`
	Prefix   string `yaml:"prefix"`
	Index    int    `yaml:"index"`
}

var CDNBaseUrl string

func MustLoad(path string) *Config {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exists: " + path)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic(err)
	}
	if cfg.Env == "local" {
		CDNBaseUrl = "http://localhost:8003"
	} else if cfg.Env == "dev" {
		CDNBaseUrl = "http://64.23.188.226:8003"

	}
	return &cfg
}

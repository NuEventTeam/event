package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

type Config struct {
	Env      string   `yaml:"env"`
	GRPC     GRPC     `yaml:"grpc"`
	Database Database `yaml:"database"`
	Cache    Cache    `yaml:"cache"`
}

type GRPC struct {
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
	Port   string `yaml:"port"`
	Host   string `yaml:"host"`
	Prefix string `yaml:"prefix"`
	Index  int    `yaml:"index"`
}

func MustLoad() *Config {
	path := "./config/local.yaml"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		panic("config file does not exists: " + path)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}

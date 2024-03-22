package config

import (

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

type CDN struct {
	KeyID string `yaml:"key_id"`
	SecretAccessKey string `yaml:"secret_access_key"`
	Region string `yaml:"region"`
	BucketName string `yaml:"bucket_name"
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

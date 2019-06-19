package config

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	"os"
)

type SyntectConfig struct {
	Name    string  `json:"name"`
	Syntect Syntect `json:"syntect"`
}

type Syntect struct {
	Path string `json:"path"`
	Host string `json:"host"`
	Port string `json:"port"`
}

var DefaultConfig = SyntectConfig{
	Name: "SyntectService",
	Syntect: Syntect{
		Path: "/usr/local/bin/syntect_server",
		Host: "127.0.0.1",
		Port: "9999",
	},
}

func init() {
	configPath := os.Getenv("SYNTECT_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	conf := config.NewConfig()
	_ = conf.Load(
		file.NewSource(file.WithPath(configPath)),
		env.NewSource(env.WithStrippedPrefix("SYNTECT")),
	)
	err := conf.Scan(&DefaultConfig)
	if err != nil {
		panic(err)
	}
}

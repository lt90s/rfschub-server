package config

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	"os"
)

type AccountConfig struct {
	Name    string        `json:"name"`
	Store   string        `json:"store"`
	Mongodb MongodbConfig `json:"mongodb"`
}

type MongodbConfig struct {
	Uri      string `json:"uri"`
	Database string `json:"database"`
}

var DefaultConfig = AccountConfig{
	Name:  "AccountService",
	Store: "mongodb",
	Mongodb: MongodbConfig{
		Uri:      "mongodb://127.0.0.1:27017",
		Database: "rfschub",
	},
}

func init() {
	configPath := os.Getenv("ACCOUNT_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	conf := config.NewConfig()
	_ = conf.Load(
		file.NewSource(file.WithPath(configPath)),
		env.NewSource(env.WithStrippedPrefix("ACCOUNT")),
	)
	err := conf.Scan(&DefaultConfig)
	if err != nil {
		panic(err)
	}
}

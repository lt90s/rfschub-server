package config

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	"os"
)

type RepositoryConfig struct {
	Name     string        `json:"name"`
	Store    string        `json:"name"`
	Mongodb  MongodbConfig `json:"mongodb"`
	MongoUri string        `json:"mongoduri"`
	Syncer   SyncConfig    `json:"syncer"`
	Syntect  string        `json:"syntect"`
}

type SyncConfig struct {
	Gits        string `json:"gits"`
	Concurrency int    `json:"concurrency"`
	Timeout     int    `json:"timeout"`
}

type MongodbConfig struct {
	Uri      string `json:"uri"`
	Database string `json:"database"`
}

var DefaultConfig = RepositoryConfig{
	Name:  "RepositoryService",
	Store: "mongodb",
	Mongodb: MongodbConfig{
		Uri:      "mongodb://127.0.0.1:27017",
		Database: "rfschub",
	},
	Syncer: SyncConfig{
		Gits:        "GitService",
		Concurrency: 16,
		Timeout:     30,
	},
	Syntect: "SyntectService",
}

func init() {
	configPath := os.Getenv("REPOSITORY_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	conf := config.NewConfig()
	_ = conf.Load(
		file.NewSource(file.WithPath(configPath)),
		env.NewSource(env.WithStrippedPrefix("REPOSITORY")),
	)
	err := conf.Scan(&DefaultConfig)
	if err != nil {
		panic(err)
	}
}

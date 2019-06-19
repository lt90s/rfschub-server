package config

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	"os"
)

type GitService struct {
	Name string `json:"name"`
}

type IndexConfig struct {
	Name        string        `json:"name"`        // index service name
	Concurrency int           `json:"concurrency"` // how many indexers can run concurrently
	Path        string        `json:"path"`        // universal-ctags binary path
	Timeout     int           `json:"expire"`      // index task timeout (second)
	Size        int64         `json:"size"`        // max file size to index
	Gits        GitService    `json:"gits"`        // git service name
	Store       string        `json:"store"`
	Mongodb     MongodbConfig `json:"mongodb"`
}

type MongodbConfig struct {
	Uri      string `json:"uri"`
	Database string `json:"database"`
}

var DefaultConfig = IndexConfig{
	Name:        "IndexService",
	Concurrency: 2,
	Path:        "/usr/local/bin/universal-ctags",
	Timeout:     600,
	Gits: GitService{
		Name: "GitService",
	},
	Size:  256 * 1024, // 256KB
	Store: "mongodb",
	Mongodb: MongodbConfig{
		Uri:      "mongodb://127.0.0.1:27017",
		Database: "rfschub",
	},
}

func init() {
	configPath := os.Getenv("INDEX_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	conf := config.NewConfig()
	_ = conf.Load(
		file.NewSource(file.WithPath(configPath)),
		env.NewSource(env.WithStrippedPrefix("INDEX")),
	)
	err := conf.Scan(&DefaultConfig)
	if err != nil {
		panic(err)
	}
}

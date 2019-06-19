package config

import (
	accountClient "github.com/lt90s/rfschub-server/account/client"
	gitsClient "github.com/lt90s/rfschub-server/gits/client"
	indexClient "github.com/lt90s/rfschub-server/index/client"
	projectClient "github.com/lt90s/rfschub-server/project/client"
	repoClient "github.com/lt90s/rfschub-server/repository/client"
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	"os"
)

type ApiConfig struct {
	Client ClientConfig `json:"client"`
	Jwt    JwtConfig    `json:"jwt"`
}

type ClientConfig struct {
	Services ServiceConfig `json:"services"`
}

type JwtConfig struct {
	Realm string `json:"realm"`
	Key   []byte `json:"key"`
}

type ServiceConfig struct {
	Git        gitsClient.ServerConfig    `json:"git"`
	Repository repoClient.ServerConfig    `json:"repository"`
	Account    accountClient.ServerConfig `json:"account"`
	Project    projectClient.ServerConfig `json:"project"`
	Index      indexClient.ServerConfig   `json:"index"`
}

var (
	DefaultJwtKey = []byte("ba9e6e6fa65a1093e2daaa1ba20c416d7583041ccaaf6b274e6a89e5fca8f3c0")
)

var DefaultConfig = ApiConfig{
	Client: ClientConfig{
		Services: ServiceConfig{
			Git: gitsClient.ServerConfig{
				ServiceName: "GitService",
			},
			Repository: repoClient.ServerConfig{
				ServiceName: "RepositoryService",
			},
			Account: accountClient.ServerConfig{
				ServiceName: "AccountService",
			},
			Project: projectClient.ServerConfig{
				ServiceName: "ProjectService",
			},
			Index: indexClient.ServerConfig{
				ServiceName: "IndexService",
			},
		},
	},
	Jwt: JwtConfig{
		Realm: "rfschub.com",
		Key:   DefaultJwtKey,
	},
}

func init() {
	configPath := os.Getenv("API_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	conf := config.NewConfig()
	_ = conf.Load(
		file.NewSource(file.WithPath(configPath)),
		env.NewSource(env.WithStrippedPrefix("API")),
	)
	err := conf.Scan(&DefaultConfig)
	if err != nil {
		panic(err)
	}
}

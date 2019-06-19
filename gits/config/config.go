package config

import (
	"github.com/micro/go-config"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
	"os"
)

type GitConfer interface {
	GetServiceConf() ServiceConf
	GetCommandConf() CommandConf
}

type ServiceConf struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

type CommandConcurrency struct {
	Clone   int64 `json:"clone"`
	Archive int64 `json:"archive"`
	Other   int64 `json:"other"`
}

type CommandConf struct {
	Path           string             `json:"path"`
	Data           string             `json:"data"`
	Concurrency    CommandConcurrency `json:"concurrency"`
	CloneTimeout   int                `json:"clonetimeout"`
	ArchiveTimeout int                `json:"archivetimeout"`
	DefaultTimeout int                `json:"defaulttimeout"`
}

type configuration struct {
	Service ServiceConf `json:"service"`
	Command CommandConf `json:"command"`
}

func (c configuration) GetServiceConf() ServiceConf {
	return c.Service
}

func (c configuration) GetCommandConf() CommandConf {
	return c.Command
}

var DefaultGitConfer = configuration{
	Service: ServiceConf{
		Name: "GitService",
		Id:   "GitService_1",
	},
	Command: CommandConf{
		Path: "/usr/local/bin/git",
		Data: "/tmp",
		Concurrency: CommandConcurrency{
			Clone:   4,
			Archive: 12,
			Other:   1,
		},
		// TODO: what if the repository is so big that it cannot be cloned within `CloneTimeout`
		CloneTimeout:   1200, // 20 minutes
		DefaultTimeout: 60,   // 1 minutes
		ArchiveTimeout: 600,  // 10 minutes
	},
}

func init() {
	configPath := os.Getenv("GITS_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}
	conf := config.NewConfig()
	_ = conf.Load(
		file.NewSource(file.WithPath(configPath)),
		env.NewSource(env.WithStrippedPrefix("GITS")),
	)
	err := conf.Scan(&DefaultGitConfer)
	if err != nil {
		panic(err)
	}
}

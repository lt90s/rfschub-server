package client

import (
	accountClient "github.com/lt90s/rfschub-server/account/client"
	"github.com/lt90s/rfschub-server/account/proto"
	"github.com/lt90s/rfschub-server/api/config"
	gitsClient "github.com/lt90s/rfschub-server/gits/client"
	"github.com/lt90s/rfschub-server/gits/proto"
	indexClient "github.com/lt90s/rfschub-server/index/client"
	"github.com/lt90s/rfschub-server/index/proto"
	projectClient "github.com/lt90s/rfschub-server/project/client"
	"github.com/lt90s/rfschub-server/project/proto"
	repoClient "github.com/lt90s/rfschub-server/repository/client"
	"github.com/lt90s/rfschub-server/repository/proto"
)

type Client struct {
	GitClient     gits.GitsService
	RepoClient    repository.RepositoryService
	AccountClient account.AccountService
	ProjectClient project.ProjectService
	IndexClient   index.IndexService
}

var DefaultClient = NewClient(config.DefaultConfig.Client)

func NewClient(conf config.ClientConfig) *Client {
	return &Client{
		GitClient:     gitsClient.New(conf.Services.Git),
		RepoClient:    repoClient.New(conf.Services.Repository),
		AccountClient: accountClient.New(conf.Services.Account),
		ProjectClient: projectClient.New(conf.Services.Project),
		IndexClient:   indexClient.New(conf.Services.Index),
	}
}

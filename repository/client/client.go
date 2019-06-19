package client

import (
	proto "github.com/lt90s/rfschub-server/repository/proto"
	"github.com/micro/go-micro"
)

type ServerConfig struct {
	ServiceName string
}

func New(conf ServerConfig) proto.RepositoryService {
	service := micro.NewService()
	service.Init()

	client := proto.NewRepositoryService(conf.ServiceName, service.Client())
	return client
}

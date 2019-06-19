package client

import (
	proto "github.com/lt90s/rfschub-server/index/proto"
	"github.com/micro/go-micro"
)

type ServerConfig struct {
	ServiceName string
}

func New(conf ServerConfig) proto.IndexService {
	service := micro.NewService()
	service.Init()

	client := proto.NewIndexService(conf.ServiceName, service.Client())
	return client
}

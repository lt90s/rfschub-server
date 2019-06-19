package client

import (
	proto "github.com/lt90s/rfschub-server/gits/proto"
	"github.com/micro/go-micro"
)

type ServerConfig struct {
	ServiceName string
}

func New(conf ServerConfig) proto.GitsService {
	service := micro.NewService()
	service.Init()

	client := proto.NewGitsService(conf.ServiceName, service.Client())
	return client
}

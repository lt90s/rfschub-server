package client

import (
	proto "github.com/lt90s/rfschub-server/project/proto"
	"github.com/micro/go-micro"
)

type ServerConfig struct {
	ServiceName string `json:"name"`
}

func New(conf ServerConfig) proto.ProjectService {
	s := micro.NewService()
	s.Init()

	client := proto.NewProjectService(conf.ServiceName, s.Client())
	return client
}

package client

import (
	proto "github.com/lt90s/rfschub-server/syntect/proto"
	"github.com/micro/go-micro"
)

type ServerConfig struct {
	ServiceName string `json:"name"`
}

func New(conf ServerConfig) proto.SyntectService {
	s := micro.NewService()
	s.Init()

	client := proto.NewSyntectService(conf.ServiceName, s.Client())
	return client
}

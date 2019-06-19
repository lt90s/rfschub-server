package client

import (
	"github.com/lt90s/rfschub-server/account/proto"
	"github.com/micro/go-micro"
)

type ServerConfig struct {
	ServiceName string
}

func New(config ServerConfig) account.AccountService {
	s := micro.NewService()
	s.Init()

	client := account.NewAccountService(config.ServiceName, s.Client())
	return client
}

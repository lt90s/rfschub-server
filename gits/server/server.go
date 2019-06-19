package server

import (
	"github.com/lt90s/rfschub-server/gits/config"
	proto "github.com/lt90s/rfschub-server/gits/proto"
	"github.com/lt90s/rfschub-server/gits/service"
	"github.com/micro/go-micro"
	log "github.com/sirupsen/logrus"
)

type GitsServer struct {
	s micro.Service
}

func New(id string, name string) GitsServer {
	metadata := map[string]string{
		"id": id,
	}

	s := micro.NewService(micro.Name(name), micro.Metadata(metadata))

	s.Init()

	err := proto.RegisterGitsHandler(s.Server(), service.New(config.DefaultGitConfer))
	if err != nil {
		log.Panicf("register gits handler failed: err = %s\n", err.Error())
	}
	return GitsServer{
		s: s,
	}
}

func (g GitsServer) Start() error {
	return g.s.Run()
}

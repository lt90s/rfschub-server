package main

import (
	"github.com/lt90s/rfschub-server/syntect/config"
	proto "github.com/lt90s/rfschub-server/syntect/proto"
	"github.com/lt90s/rfschub-server/syntect/service"
	"github.com/micro/go-micro"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	conf := config.DefaultConfig
	s := micro.NewService(micro.Name(conf.Name))
	s.Init()

	servicer := service.NewSyntectService()
	defer servicer.Stop()

	err := proto.RegisterSyntectServiceHandler(s.Server(), servicer)
	if err != nil {
		log.Panicf("register syntect service handler error: %s", err.Error())
	}

	s.Run()
}

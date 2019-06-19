package main

import (
	"github.com/lt90s/rfschub-server/index/config"
	proto "github.com/lt90s/rfschub-server/index/proto"
	"github.com/lt90s/rfschub-server/index/service"
	"github.com/lt90s/rfschub-server/index/store/mongodb"
	"github.com/micro/go-micro"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	s := micro.NewService(micro.Name(config.DefaultConfig.Name))
	s.Init()

	store := mongodb.NewMongodbStore()
	err := proto.RegisterIndexHandler(s.Server(), service.NewIndexService(store))
	if err != nil {
		log.Panicf("register index service handler error: %s", err.Error())
	}

	s.Run()
}

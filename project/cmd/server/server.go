package main

import (
	"github.com/lt90s/rfschub-server/project/config"
	proto "github.com/lt90s/rfschub-server/project/proto"
	"github.com/lt90s/rfschub-server/project/service"
	"github.com/lt90s/rfschub-server/project/store"
	"github.com/lt90s/rfschub-server/project/store/mongodb"
	"github.com/micro/go-micro"
	"github.com/sirupsen/logrus"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	var store store.Store
	switch config.DefaultConfig.Store {
	case "mongodb":
		store = mongodb.NewMongodbStore()
	//case "mock":
	//	store = mock.NewMockStore()
	default:
		store = mongodb.NewMongodbStore()
	}
	s := micro.NewService(micro.Name(config.DefaultConfig.Name))
	s.Init()

	err := proto.RegisterProjectHandler(s.Server(), service.NewProjectService(store))
	if err != nil {
		logrus.Panicf("register project service handler error: %s", err.Error())
	}

	s.Run()
}

package main

import (
	"github.com/lt90s/rfschub-server/repository/config"
	"github.com/lt90s/rfschub-server/repository/service"
	"github.com/lt90s/rfschub-server/repository/store"
	"github.com/lt90s/rfschub-server/repository/store/mockdb"
	"github.com/lt90s/rfschub-server/repository/store/mongodb"
	"github.com/micro/go-micro"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	var store store.Store
	switch config.DefaultConfig.Store {
	//case "mock":
	//	store = mockdb.NewMockStore()
	case "mongodb":
		store = mongodb.NewMongodbStore()
	default:
		store = mockdb.NewMockStore()
	}
	s := micro.NewService(micro.Name(config.DefaultConfig.Name))
	s.Init()

	log.Info(config.DefaultConfig.Name)
	handler := service.NewRepositoryService(config.DefaultConfig, store)

	err := micro.RegisterHandler(s.Server(), handler)
	if err != nil {
		log.Panicf("register repository service handler failed: err = %s\n", err.Error())
	}

	log.Info("starts to run repository service...")
	err = s.Run()
	if err != nil {
		log.Warnf("repository service exit error: %s", err.Error())
	}
}

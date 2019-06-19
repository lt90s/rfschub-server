package main

import (
	"github.com/lt90s/rfschub-server/account/config"
	"github.com/lt90s/rfschub-server/account/proto"
	"github.com/lt90s/rfschub-server/account/service"
	"github.com/lt90s/rfschub-server/account/store"
	"github.com/lt90s/rfschub-server/account/store/mongodb"
	"github.com/micro/go-micro"
	log "github.com/sirupsen/logrus"
)

const name = "AccountService"

func main() {
	log.SetLevel(log.DebugLevel)

	var store store.Store

	switch config.DefaultConfig.Store {
	//case "mock":
	//	store = mock.NewMockStore()
	case "mongodb":
		store = mongodb.NewMongodbStore()
	default:
		store = mongodb.NewMongodbStore()
	}

	s := micro.NewService(micro.Name(name))
	s.Init()

	err := account.RegisterAccountServiceHandler(s.Server(), service.New(store))
	if err != nil {
		log.Panicf("register repository service handler failed: err = %s\n", err.Error())
	}

	log.Info("starts to run account service...")
	err = s.Run()
	if err != nil {
		log.Warnf("account service exit error: %s", err.Error())
	}
}

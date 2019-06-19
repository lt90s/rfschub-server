package main

import (
	"github.com/lt90s/rfschub-server/gits/config"
	"github.com/lt90s/rfschub-server/gits/server"
	"github.com/sirupsen/logrus"
)

func main() {

	logrus.SetLevel(logrus.DebugLevel)

	serviceConf := config.DefaultGitConfer.GetServiceConf()

	s := server.New(serviceConf.Id, serviceConf.Name)
	s.Start()
}

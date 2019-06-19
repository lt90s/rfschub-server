package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/rfschub-server/api/route"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	router := gin.Default()

	route.SetupRouter(router)

	router.Run(":8888")
}

package route

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/rfschub-server/api/client"
	"github.com/lt90s/rfschub-server/api/middlewares"
	"github.com/lt90s/rfschub-server/api/route/account"
	"github.com/lt90s/rfschub-server/api/route/project"
	"github.com/lt90s/rfschub-server/api/route/repository"
)

func SetupRouter(router *gin.Engine) {
	router.Use(middlewares.SetClientMiddleware(client.DefaultClient))
	router.Use(middlewares.ResponseMiddleware)

	account.SetupAccountRouter(router)
	repository.SetupRepositoryRoute(router)
	project.SetupProjectRouter(router)
}

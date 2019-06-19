package repository

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/rfschub-server/api/middlewares"
)

func SetupRepositoryRoute(router *gin.Engine) {
	auth := middlewares.JWTMiddleware.MiddlewareFunc()

	group := router.Group("/repository")

	group.GET("status", auth, GetRepositoryStatus)
	group.POST("clone", auth, CloneRepository)
	group.GET("namedCommits", auth, GetNamedCommits)
}

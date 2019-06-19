package account

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/rfschub-server/api/middlewares"
)

func SetupAccountRouter(router *gin.Engine) {
	router.POST("/account/login", middlewares.JWTMiddleware.LoginHandler)
	router.POST("/account/register", registerAccount)
	router.GET("/account/info", middlewares.JWTMiddleware.MiddlewareFunc(), getSelfInfo)
	router.GET("/account/info/:name", getUserInfo)
}

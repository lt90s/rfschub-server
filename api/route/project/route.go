package project

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/rfschub-server/api/middlewares"
)

func SetupProjectRouter(router *gin.Engine) {
	authFunc := middlewares.JWTMiddleware.MiddlewareFunc()

	router.POST("/project", authFunc, createProject)
	router.GET("/project", getProjectInfo)
	router.GET("/project/list", getUserProjects)
	router.GET("/project/symbol", searchSymbol)
	router.POST("/project/annotation", authFunc, addAnnotation)
	router.GET("/project/annotation/lines", getAnnotationLines)
	router.GET("/project/annotations", getAnnotations)
	router.GET("/project/annotation/latest", getLatestAnnotations)

	router.GET("/project/directory", getProjectDirectory)
	router.GET("/project/blob", getProjectBlob)
}

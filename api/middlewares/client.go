package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/rfschub-server/api/client"
)

const (
	clientKey = "__client_key__"
)

func SetClientMiddleware(client *client.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(clientKey, client)
	}
}

func GetClient(c *gin.Context) *client.Client {
	value, _ := c.Get(clientKey)
	return value.(*client.Client)
}

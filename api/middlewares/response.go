package middlewares

import (
	"github.com/gin-gonic/gin"
	"github.com/lt90s/rfschub-server/common/errors"
	"net/http"
)

const (
	errorKey = "__response_error_key__"
	dataKey  = "__response_data_key__"
)

func SetError(c *gin.Context, err error) {
	c.Set(errorKey, err)
}

func SetData(c *gin.Context, data interface{}) {
	c.Set(dataKey, data)
}

func ResponseMiddleware(c *gin.Context) {
	c.Next()
	if c.Writer.Written() {
		return
	}

	value, ok := c.Get(errorKey)
	if ok {
		if err, ok := value.(errors.Error); ok {
			c.JSON(err.HttpCode, gin.H{
				"code":    err.Code,
				"message": err.Message,
			})
		} else if err, ok := value.(error); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    -1,
				"message": err.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    -1,
				"message": "Unknown error",
			})
		}
		return
	}

	value, ok = c.Get(dataKey)
	if ok {
		c.JSON(http.StatusOK, value)
	}
}

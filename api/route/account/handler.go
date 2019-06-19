package account

import (
	"context"
	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/lt90s/rfschub-server/account/proto"
	"github.com/lt90s/rfschub-server/api/middlewares"
	"github.com/lt90s/rfschub-server/common/errors"
	"net/http"
	"regexp"
)

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

const (
	passwordMinSize = 6
)

var (
	nameRegexp  = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_]{0,16}$")
	emailRegexp = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

func registerAccount(c *gin.Context) {
	var data registerRequest
	err := c.ShouldBindJSON(&data)
	if err != nil || !nameRegexp.MatchString(data.Name) || !emailRegexp.MatchString(data.Email) || len(data.Password) < passwordMinSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid parameters",
		})
		return
	}
	client := middlewares.GetClient(c)

	req := &account.RegisterRequest{
		Name:     data.Name,
		Email:    data.Email,
		Password: data.Password,
	}

	rsp, err := client.AccountClient.Register(context.Background(), req)
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, rsp)
}

func getSelfInfo(c *gin.Context) {
	claim := jwt.ExtractClaims(c)
	middlewares.SetData(c, claim)
}

func getUserInfo(c *gin.Context) {
	user := c.Param("name")

	client := middlewares.GetClient(c)
	rsp, err := client.AccountClient.AccountInfoByName(context.Background(), &account.AccountName{Name: user})
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, rsp)

}

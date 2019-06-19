package middlewares

import (
	"context"
	"fmt"
	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/lt90s/rfschub-server/account/proto"
	"github.com/lt90s/rfschub-server/api/config"
	jwtGo "gopkg.in/dgrijalva/jwt-go.v3"
	"time"
)

const (
	identityKey = "ID"
	cookieKey   = "JWTToken"
)

var (
	JWTMiddleware = NewJwtMiddleware()
)

type loginRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AccountMatchResult struct {
	Id   string
	Role string
}

func NewJwtMiddleware() *jwt.GinJWTMiddleware {
	authenticator := func(c *gin.Context) (interface{}, error) {
		var request loginRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			return nil, jwt.ErrMissingLoginValues
		}

		if (request.Name == "" && request.Email == "") || request.Password == "" {
			return nil, jwt.ErrMissingLoginValues
		}

		req := account.LoginRequest{
			Name:     request.Name,
			Email:    request.Email,
			Password: request.Password,
		}

		client := GetClient(c)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		rsp, err := client.AccountClient.Login(ctx, &req)
		if err != nil {
			return nil, jwt.ErrFailedAuthentication
		}
		fmt.Println(rsp.Info)
		return rsp.Info, nil
	}

	middleware := &jwt.GinJWTMiddleware{
		Realm:             config.DefaultConfig.Jwt.Realm,
		SigningAlgorithm:  "HS256",
		Key:               config.DefaultConfig.Jwt.Key,
		Timeout:           7 * 24 * time.Hour,
		IdentityHandler:   identityHandler,
		Authenticator:     authenticator,
		PayloadFunc:       payloadFunc,
		SendCookie:        true,
		SecureCookie:      true,
		SendAuthorization: false,
		TokenLookup:       "cookie:" + cookieKey,
	}
	if err := middleware.MiddlewareInit(); err != nil {
		panic(err)
	}
	return middleware
}

func payloadFunc(data interface{}) jwt.MapClaims {
	if info, ok := data.(*account.AccountInfo); ok {
		return jwt.MapClaims{
			"id":        info.Id,
			"name":      info.Name,
			"createdAt": info.CreatedAt,
		}
	}
	return jwt.MapClaims{}
}

func identityHandler(claims jwtGo.MapClaims) interface{} {
	return claims["id"]
}

// must be logined to use this
func GetUserId(c *gin.Context) string {
	uid, ok := c.Get("userID")
	if !ok {
		panic("user id not exist")
	}
	return uid.(string)
}

func ExtractUserId(c *gin.Context) string {
	cookie, _ := c.Cookie(cookieKey)
	if cookie == "" {
		return ""
	}
	token, err := jwtGo.Parse(cookie, func(token *jwtGo.Token) (i interface{}, e error) {
		return config.DefaultConfig.Jwt.Key, nil
	})
	if err != nil {
		return ""
	}
	claims, ok := token.Claims.(jwtGo.MapClaims)
	if !ok {
		return ""
	}

	value, ok := claims["id"]
	if !ok {
		return ""
	}

	id, ok := value.(string)
	if !ok {
		return ""
	}
	return id
}

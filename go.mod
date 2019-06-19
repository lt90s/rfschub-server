module github.com/lt90s/rfschub-server

go 1.12

require (
	github.com/appleboy/gin-jwt v2.5.0+incompatible
	github.com/gin-gonic/gin v1.4.0
	github.com/golang/protobuf v1.3.1
	github.com/micro/go-config v1.1.0
	github.com/micro/go-micro v1.1.0
	github.com/sirupsen/logrus v1.4.2
	github.com/stretchr/testify v1.3.0
	github.com/tidwall/gjson v1.2.1 // indirect
	github.com/tidwall/match v1.0.1 // indirect
	github.com/tidwall/pretty v0.0.0-20190325153808-1166b9ac2b65 // indirect
	github.com/xdg/scram v0.0.0-20180814205039-7eeb5667e42c // indirect
	github.com/xdg/stringprep v1.0.0 // indirect
	go.etcd.io/bbolt v1.3.2 // indirect
	go.mongodb.org/mongo-driver v1.0.2
	golang.org/x/crypto v0.0.0-20190325154230-a5d413f7728c
	golang.org/x/net v0.0.0-20190503192946-f4e77d36d62c
	golang.org/x/sync v0.0.0-20190423024810-112230192c58
	google.golang.org/genproto v0.0.0-20190327125643-d831d65fe17d // indirect
	gopkg.in/dgrijalva/jwt-go.v3 v3.2.0
	k8s.io/klog v0.3.2 // indirect
)

replace (
	github.com/testcontainers/testcontainer-go => github.com/testcontainers/testcontainers-go v0.0.0-20190108154635-47c0da630f72
	github.com/ugorji/go v1.1.4 => github.com/ugorji/go/codec v0.0.0-20190204201341-e444a5086c43
)

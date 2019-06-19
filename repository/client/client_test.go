package client

import (
	"context"
	"github.com/lt90s/rfschub-server/repository/proto"
	"testing"
)

var (
	conf = ServerConfig{
		ServiceName: "RepositoryService",
	}
)

func TestClient(t *testing.T) {
	client := New(conf)

	rsp, err := client.IsRepositoryExist(context.Background(), &repository.RepositoryExistRequest{Url: "github.com/lt90s/goanalytics"})
	t.Log(err)
	t.Log(rsp)
}

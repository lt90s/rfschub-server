package client

import (
	"context"
	"github.com/lt90s/rfschub-server/account/proto"
	"github.com/lt90s/rfschub-server/common/errors"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	client account.AccountService
)

func init() {
	conf := ServerConfig{
		ServiceName: "AccountService",
	}
	client = New(conf)
}
func TestClient_Register(t *testing.T) {
	ctx := context.Background()
	rsp, err := client.Register(ctx, &account.RegisterRequest{Name: "foo", Email: "foo@bar.com", Password: "123456"})
	require.NoError(t, err)
	t.Log(rsp)

	_, err = client.Register(ctx, &account.RegisterRequest{Name: "foo", Email: "foo@bar.com", Password: "123456"})
	require.Error(t, err)

	rErr := errors.FromError(err)
	require.Equal(t, int(account.ErrorCode_ErrorNameUsed), rErr.Code)
}

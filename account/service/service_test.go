package service

import (
	"context"
	proto "github.com/lt90s/rfschub-server/account/proto"
	"github.com/lt90s/rfschub-server/account/store/mock"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAccountService_Register(t *testing.T) {
	store := mock.NewMockStore()
	s := New(store)

	req := proto.RegisterRequest{
		Name: "foo",
	}
	rsp := proto.RegisterResponse{}
	ctx := context.Background()

	err := s.Register(ctx, &req, &rsp)
	require.NoError(t, err)
	require.Equal(t, proto.ErrorCode_ErrorParameterInvalid, rsp.Code)

	req.Email = "abc.def"
	err = s.Register(ctx, &req, &rsp)
	require.NoError(t, err)
	require.Equal(t, proto.ErrorCode_ErrorParameterInvalid, rsp.Code)

	req.Email = "abc@def.com"
	err = s.Register(ctx, &req, &rsp)
	require.NoError(t, err)
	require.Equal(t, proto.ErrorCode_ErrorParameterInvalid, rsp.Code)

	req.Password = "123456"
	err = s.Register(ctx, &req, &rsp)
	require.NoError(t, err)
	require.Equal(t, proto.ErrorCode_Success, rsp.Code)

	err = s.Register(ctx, &req, &rsp)
	require.NoError(t, err)
	require.Equal(t, proto.ErrorCode_ErrorNameUsed, rsp.Code)

	req.Name = "bar"
	err = s.Register(ctx, &req, &rsp)
	require.NoError(t, err)
	require.Equal(t, proto.ErrorCode_ErrorEmailRegistered, rsp.Code)
}

func TestAccountService_Login(t *testing.T) {
	store := mock.NewMockStore()
	s := New(store)

	req := proto.RegisterRequest{
		Name:     "foo",
		Email:    "abc@def.com",
		Password: "123456",
	}
	rsp := proto.RegisterResponse{}
	ctx := context.Background()

	err := s.Register(ctx, &req, &rsp)
	require.NoError(t, err)
	require.Equal(t, proto.ErrorCode_Success, rsp.Code)

	loginRequest := proto.LoginRequest{
		Name:     "foo",
		Password: "123456",
	}
	loginResponse := proto.LoginResponse{}
	err = s.Login(ctx, &loginRequest, &loginResponse)
	require.NoError(t, err)
	require.Equal(t, proto.ErrorCode_Success, loginResponse.Code)
	t.Log(loginResponse.Info)

	loginRequest = proto.LoginRequest{
		Email:    "abc@def.com",
		Password: "123456",
	}
	err = s.Login(ctx, &loginRequest, &loginResponse)
	require.NoError(t, err)
	require.Equal(t, proto.ErrorCode_Success, loginResponse.Code)
	t.Log(loginResponse.Info)

	loginRequest = proto.LoginRequest{
		Name:     "foox",
		Password: "123456",
	}
	err = s.Login(ctx, &loginRequest, &loginResponse)
	require.NoError(t, err)
	require.Equal(t, proto.ErrorCode_ErrorParameterInvalid, loginResponse.Code)

	loginRequest = proto.LoginRequest{
		Email:    "abc@def.com",
		Password: "1234567",
	}
	err = s.Login(ctx, &loginRequest, &loginResponse)
	require.NoError(t, err)
	require.Equal(t, proto.ErrorCode_ErrorParameterInvalid, loginResponse.Code)

}

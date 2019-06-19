package store

import (
	"context"
	"errors"
)

type Store interface {
	CreateAccount(ctx context.Context, name, email, password string, code []byte) (string, error)
	LoginAccount(ctx context.Context, name, email, password string) (info AccountInfo, err error)
	GetAccountId(ctx context.Context, name string) (string, error)
	GetAccountInfoByName(ctx context.Context, name string) (info AccountInfo, err error)
	GetAccountsBasicInfo(ctx context.Context, uids []string) ([]BasicInfo, error)
}

var (
	ErrNameUsed        = errors.New("name already used")
	ErrEmailRegistered = errors.New("email already registered")
	ErrNoMatch         = errors.New("name or password not correct")
	ErrNoAccount       = errors.New("account not exist")
	ErrNotActivate     = errors.New("account not activated")
)

type AccountInfo struct {
	Id        string `json:"id"`
	Name      string `json:"name" bson:"name"`
	CreatedAt int64  `json:"createdAt" bson:"createdAt"`
}

type BasicInfo struct {
	Id   string
	Name string
}

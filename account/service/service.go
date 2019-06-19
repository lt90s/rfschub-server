package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	proto "github.com/lt90s/rfschub-server/account/proto"
	"github.com/lt90s/rfschub-server/account/store"
	"github.com/lt90s/rfschub-server/common/errors"
	log "github.com/sirupsen/logrus"
	"regexp"
	"time"
)

type accountService struct {
	store store.Store
}

func New(store store.Store) proto.AccountServiceHandler {
	return &accountService{
		store: store,
	}
}

const (
	passwordMinLen = 6
)

var (
	nameRegexp    = regexp.MustCompile("^[a-zA-Z][a-zA-Z0-9_]{0,16}$")
	emailRegexp   = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	reservedNames = map[string]struct{}{
		"create": {},
	}
)

func (a *accountService) Register(ctx context.Context, req *proto.RegisterRequest, rsp *proto.RegisterResponse) error {
	if !nameRegexp.MatchString(req.Name) || len(req.Password) < passwordMinLen || !emailRegexp.MatchString(req.Email) {
		return errors.NewBadRequestError(-1, "parameter invalid")
	}

	if _, ok := reservedNames[req.Name]; ok {
		return errors.NewBadRequestError(int(proto.ErrorCode_ErrorNameUsed), "name already used")
	}

	code := make([]byte, 16)
	_, err := rand.Read(code)
	if err != nil {
		return errors.NewInternalError(-1, err.Error())
	}
	id, err := a.store.CreateAccount(ctx, req.Name, req.Email, req.Password, code)

	activateCode := hex.EncodeToString(code) + id
	log.Infof("[Register] activate code: name=%s code=%s", req.Name, activateCode)

	if err != nil {
		switch err {
		case store.ErrNameUsed:
			return errors.NewBadRequestError(int(proto.ErrorCode_ErrorNameUsed), err.Error())
		case store.ErrEmailRegistered:
			return errors.NewBadRequestError(int(proto.ErrorCode_ErrorEmailRegistered), err.Error())
		case store.ErrNotActivate:
			return errors.NewBadRequestError(int(proto.ErrorCode_ErrorNotActivated), err.Error())
		default:
			return errors.NewInternalError(-1, err.Error())
		}
	}

	// TODO: send activation email
	return nil
}

func (a *accountService) Login(ctx context.Context, req *proto.LoginRequest, rsp *proto.LoginResponse) error {
	log.Debugf("login request: name=%s email=%s", req.Name, req.Email)
	info, err := a.store.LoginAccount(ctx, req.Name, req.Email, req.Password)
	if err != nil {
		if err == store.ErrNoMatch {
			return errors.NewUnauthorizedError(int(proto.ErrorCode_ErrorNamePasswordMisMatch), "name or password wrong")
		} else {
			return errors.NewInternalError(-1, err.Error())
		}
	}
	fmt.Println(info)
	rsp.Info = &proto.AccountInfo{
		Id:        info.Id,
		Name:      info.Name,
		CreatedAt: info.CreatedAt,
	}
	return nil
}

func (a *accountService) AccountId(ctx context.Context, req *proto.AccountIdRequest, rsp *proto.AccountIdResponse) error {
	now := time.Now()
	log.Debugf("[AccountId]: name=%s", req.Username)
	id, err := a.store.GetAccountId(ctx, req.Username)
	if err != nil {
		return err
	}

	log.Debugf("[AccountId]: name=%s uid=%s", req.Username, id)
	rsp.Uid = id
	log.Debugf("[AccountId] time consume: %v", time.Since(now))
	return nil
}

func (a *accountService) AccountsBasicInfo(ctx context.Context, req *proto.AccountsBasicInfoRequest, rsp *proto.AccountsBasicInfoResponse) error {
	log.Debugf("[AccountsBasicInfo]: uids=%v", req.Uids)
	infos, err := a.store.GetAccountsBasicInfo(ctx, req.Uids)
	if err != nil {
		return err
	}

	for _, info := range infos {
		rsp.Infos = append(rsp.Infos, &proto.BasicInfo{Id: info.Id, Name: info.Name})
	}
	return nil
}

func (a *accountService) AccountInfoByName(ctx context.Context, req *proto.AccountName, rsp *proto.AccountInfo) error {
	info, err := a.store.GetAccountInfoByName(ctx, req.Name)
	if err != nil {
		if err == store.ErrNoAccount {
			return errors.NewNotFoundError(-1, err.Error())
		} else {
			return errors.NewInternalError(-1, err.Error())
		}
	}
	rsp.Id = info.Id
	rsp.Name = info.Name
	rsp.CreatedAt = info.CreatedAt
	return nil
}

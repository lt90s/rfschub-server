package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/lt90s/rfschub-server/common/errors"
	"github.com/lt90s/rfschub-server/syntect/config"
	proto "github.com/lt90s/rfschub-server/syntect/proto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context/ctxhttp"
	"net/http"
	"os/exec"
	"path"
	"time"
)

type syntectService struct {
	cmd *exec.Cmd
	url string
}

const (
	testCode = `#include<stdio.h>
			 	int main() {
				    printf("hello world\n");
				    return 0;
			 	}`
)

func NewSyntectService() *syntectService {
	cmd := exec.Command(config.DefaultConfig.Syntect.Path)
	host, port := config.DefaultConfig.Syntect.Host, config.DefaultConfig.Syntect.Port
	cmd.Env = []string{
		"ROCKET_ADDRESS=" + host,
		"ROCKET_PORT=" + port,
	}

	err := cmd.Start()
	if err != nil {
		log.Panic(err)
	}

	// give syntect server some time to boot
	time.Sleep(1 * time.Second)

	s := &syntectService{
		cmd: cmd,
		url: fmt.Sprintf("http://%s:%s/", host, port),
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var rsp proto.RenderCodeResponse
	req := &proto.RenderCodeRequest{
		File:  "hello.c",
		Theme: proto.CodeTheme_SolarizedLight,
		Code:  testCode,
	}
	err = s.RenderCode(ctx, req, &rsp)
	if err != nil {
		_ = cmd.Process.Kill()
		log.Panicf("syntect server may not work correct: %s", err.Error())
	}
	log.Debugf("rendered Code: %s", rsp.RenderedCode)
	return s
}

type syntectRequest struct {
	Extension string `json:"extension"`
	Theme     string `json:"theme"`
	Code      string `json:"code"`
}

type syntectResponse struct {
	Data string `json:"data"`
}

func (service *syntectService) RenderCode(ctx context.Context, req *proto.RenderCodeRequest, rsp *proto.RenderCodeResponse) error {
	theme := "Solarized (light)"
	switch req.Theme {
	case proto.CodeTheme_SolarizedDark:
		theme = "Solarized (dark)"
	case proto.CodeTheme_SolarizedLight:
		theme = "Solarized (light)"
	default:
		theme = "Solarized (light)"
	}
	log.Debugf("[RenderCode]: file=%s theme=%s", req.File, req.File)

	ext := path.Ext(req.File)
	if ext == "" {
		return errors.NewBadRequestError(-1, "no file extension")
	}
	data, _ := json.Marshal(syntectRequest{
		Extension: path.Ext(req.File)[1:],
		Code:      req.Code,
		Theme:     theme,
	})

	response, err := ctxhttp.Post(ctx, http.DefaultClient, service.url, "application/json", bytes.NewReader(data))

	if err != nil {
		return errors.NewInternalError(-1, err.Error())
	}

	if response.StatusCode != http.StatusOK {
		return errors.NewInternalError(-1, "syntect server error")
	}
	defer response.Body.Close()
	var tmp syntectResponse
	if err := json.NewDecoder(response.Body).Decode(&tmp); err != nil {
		return errors.NewInternalError(-1, err.Error())
	}

	rsp.RenderedCode = tmp.Data
	return nil
}

func (service *syntectService) Stop() {
	_ = service.cmd.Process.Kill()
}

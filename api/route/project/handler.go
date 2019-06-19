package project

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lt90s/rfschub-server/account/proto"
	"github.com/lt90s/rfschub-server/api/client"
	"github.com/lt90s/rfschub-server/api/middlewares"
	"github.com/lt90s/rfschub-server/common/errors"
	"github.com/lt90s/rfschub-server/common/url"
	"github.com/lt90s/rfschub-server/index/proto"
	"github.com/lt90s/rfschub-server/project/proto"
	"github.com/lt90s/rfschub-server/repository/proto"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
)

func createProject(c *gin.Context) {
	var data struct {
		Url    string `json:"url"`
		Hash   string `json:"hash"`
		Name   string `json:"name"`
		Branch bool   `json:"branch"`
	}

	err := c.ShouldBindJSON(&data)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	client := middlewares.GetClient(c)

	ctx := context.Background()
	req := &repository.RepositoryExistRequest{
		Url:  data.Url,
		Hash: data.Hash,
	}
	// check if repository exists
	// should check all given parameters
	rsp, err := client.RepoClient.IsRepositoryExist(ctx, req)
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	if !rsp.Exist {
		middlewares.SetError(c, errors.NewNotFoundError(-1, "repository not exist"))
		return
	}

	newRsp, err := client.ProjectClient.NewProject(ctx, &project.NewProjectRequest{
		Uid:    middlewares.GetUserId(c),
		Url:    data.Url,
		Hash:   data.Hash,
		Name:   data.Name,
		Branch: data.Branch,
	})

	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, newRsp)
}

func getProjectInfo(c *gin.Context) {
	username := c.Query("user")
	repo := c.Query("repo")
	name := c.Query("name")

	client := middlewares.GetClient(c)
	ctx := context.Background()

	uid := middlewares.ExtractUserId(c)

	idRsp, err := client.AccountClient.AccountId(ctx, &account.AccountIdRequest{Username: username})
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	infoRsp, err := client.ProjectClient.ProjectInfo(ctx, &project.ProjectInfoRequest{Uid: uid, OwnerUid: idRsp.Uid, Url: repo, Name: name})
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, infoRsp)
}

func getUserProjects(c *gin.Context) {
	user := c.Query("user")

	client := middlewares.GetClient(c)
	ctx := context.Background()

	idRsp, err := client.AccountClient.AccountId(ctx, &account.AccountIdRequest{Username: user})
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	rsp, err := client.ProjectClient.ListProjects(ctx, &project.ListProjectsRequest{Uid: idRsp.Uid})
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, rsp)
}

func searchSymbol(c *gin.Context) {
	client := middlewares.GetClient(c)
	ctx := context.Background()

	repo := c.Query("repo")
	hash := c.Query("hash")
	name := c.Query("name")

	rsp, err := client.IndexClient.SearchSymbol(ctx, &index.SearchSymbolRequest{
		Url:    repo,
		Hash:   hash,
		Symbol: name,
	})

	if err != nil {
		log.Warnf("search symbol error: %v", err)
		middlewares.SetError(c, errors.FromError(err))
		return
	}
	middlewares.SetData(c, rsp)
}

func addAnnotation(c *gin.Context) {
	var req project.AddAnnotationRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		middlewares.SetError(c, errors.NewBadRequestError(-1, err.Error()))
		return
	}

	req.File = strings.TrimPrefix(req.File, "/")
	req.Uid = middlewares.GetUserId(c)
	client := middlewares.GetClient(c)
	ctx := context.Background()

	_, err = client.ProjectClient.AddAnnotation(ctx, &req)
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
	} else {
		middlewares.SetData(c, gin.H{})
	}
}

func getAnnotationLines(c *gin.Context) {
	pid := c.Query("pid")
	file := c.Query("file")

	client := middlewares.GetClient(c)
	ctx := context.Background()

	rsp, err := client.ProjectClient.GetAnnotationLines(ctx, &project.GetAnnotationLinesRequest{
		Pid:  pid,
		File: file,
	})
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}
	middlewares.SetData(c, rsp)
}

func getAnnotations(c *gin.Context) {
	pid := c.Query("pid")
	file := c.Query("file")
	lineNumber := c.Query("line")

	line, err := strconv.Atoi(lineNumber)
	if err != nil {
		middlewares.SetError(c, errors.NewBadRequestError(-1, "line must be number"))
		return
	}

	client := middlewares.GetClient(c)
	ctx := context.Background()

	rsp, err := client.ProjectClient.GetAnnotations(ctx, &project.GetAnnotationsRequest{
		Pid:        pid,
		File:       file,
		LineNumber: int32(line),
	})
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, rsp)
}

func getLatestAnnotations(c *gin.Context) {
	pid := c.Query("pid")
	parent := c.Query("parent")

	client := middlewares.GetClient(c)
	ctx := context.Background()

	rsp, err := client.ProjectClient.GetLatestAnnotations(ctx, &project.GetLatestAnnotationsRequest{
		Pid:    pid,
		Parent: parent,
	})

	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, rsp)
}

func doGetProjectInfo(ctx context.Context, client *client.Client, uid, username, repo, name string) (rsp *project.ProjectInfoResponse, err error) {
	idRsp, err := client.AccountClient.AccountId(ctx, &account.AccountIdRequest{Username: username})
	if err != nil {
		log.Warnf("AccountId error: username=%s error=%v", username, err)
		return
	}

	rsp, err = client.ProjectClient.ProjectInfo(ctx, &project.ProjectInfoRequest{Uid: uid, OwnerUid: idRsp.Uid, Url: repo, Name: name})
	if err != nil {
		log.Warnf("ProjectInfo error: uid=%s ownerUid=%s url=%s name=%s error=%v", uid, idRsp.Uid, repo, name, err)
		return
	}

	return
}

func getProjectDirectory(c *gin.Context) {
	user := c.Query("user")
	repo := c.Query("repo")
	name := c.Query("name")
	path := c.Query("path")

	repo, ok := url.NormalizeRepoUrl(repo)
	if !ok || name == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	client := middlewares.GetClient(c)
	ctx := context.Background()

	info, err := doGetProjectInfo(ctx, client, "", user, repo, name)
	if err != nil {
		log.Warnf("doGetProjectInfo error: user=%s repo=%s name=%s path=%s error=%v", user, repo, name, path, err)
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	// TODO: this two rpc can run parallel
	aRsp, err := client.ProjectClient.GetLatestAnnotations(ctx, &project.GetLatestAnnotationsRequest{
		Pid:    info.Id,
		Parent: path,
	})

	if err != nil {
		log.Warnf("GetLatestAnnotations error: user=%s repo=%s name=%s path=%s error=%v", user, repo, name, path, err)
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	dRsp, err := client.RepoClient.Directory(ctx, &repository.DirectoryRequest{Url: repo, Hash: info.Hash, Path: path})
	if err != nil {
		log.Warnf("Directory error: user=%s repo=%s name=%s path=%s error=%v", user, repo, name, path, err)
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, gin.H{
		"info":        info,
		"directory":   dRsp.Entries,
		"annotations": aRsp.Annotations,
	})
}

func getProjectBlob(c *gin.Context) {
	user := c.Query("user")
	repo := c.Query("repo")
	name := c.Query("name")
	file := c.Query("file")

	repo, ok := url.NormalizeRepoUrl(repo)
	if !ok || name == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	client := middlewares.GetClient(c)
	ctx := context.Background()

	uid := middlewares.ExtractUserId(c)

	info, err := doGetProjectInfo(ctx, client, uid, user, repo, name)
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	lRsp, err := client.ProjectClient.GetAnnotationLines(ctx, &project.GetAnnotationLinesRequest{
		Pid:  info.Id,
		File: file,
	})
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	bRsp, err := client.RepoClient.Blob(ctx, &repository.BlobRequest{
		Url:  repo,
		Hash: info.Hash,
		Path: file,
	})
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, gin.H{
		"info":  info,
		"blob":  bRsp,
		"lines": lRsp.Lines,
	})
}

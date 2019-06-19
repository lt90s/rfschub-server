package repository

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/lt90s/rfschub-server/api/middlewares"
	"github.com/lt90s/rfschub-server/common/errors"
	"github.com/lt90s/rfschub-server/common/url"
	"github.com/lt90s/rfschub-server/gits/proto"
	"github.com/lt90s/rfschub-server/repository/proto"
	"net/http"
)

func GetRepositoryStatus(c *gin.Context) {
	repo := c.Query("repo")

	repo, ok := url.NormalizeRepoUrl(repo)
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	client := middlewares.GetClient(c)
	ctx := context.Background()
	req := &gits.GetCloneStatusRequest{Url: repo}
	rsp, err := client.GitClient.GetCloneStatus(ctx, req)
	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, rsp)
}

func CloneRepository(c *gin.Context) {
	var cloneRequest struct {
		Repo string `json:"repo"`
	}
	err := c.ShouldBindJSON(&cloneRequest)
	if err != nil {
		c.AbortWithStatus(400)
		return
	}

	repo, ok := url.NormalizeRepoUrl(cloneRequest.Repo)
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	client := middlewares.GetClient(c)
	req := &gits.CloneRequest{Url: repo}
	rsp, err := client.GitClient.Clone(ctx, req)

	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, rsp)
}

func GetNamedCommits(c *gin.Context) {
	repo := c.Query("repo")

	repo, ok := url.NormalizeRepoUrl(repo)
	if !ok {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	client := middlewares.GetClient(c)
	ctx := context.Background()
	req := &repository.NamedCommitsRequest{Url: repo}
	rsp, err := client.RepoClient.NamedCommits(ctx, req)

	if err != nil {
		middlewares.SetError(c, errors.FromError(err))
		return
	}

	middlewares.SetData(c, rsp)
}

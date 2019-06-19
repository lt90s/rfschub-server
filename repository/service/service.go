package service

import (
	"context"
	"github.com/lt90s/rfschub-server/common/errors"
	"github.com/lt90s/rfschub-server/common/url"
	"github.com/lt90s/rfschub-server/repository/config"
	proto "github.com/lt90s/rfschub-server/repository/proto"
	"github.com/lt90s/rfschub-server/repository/store"
	"github.com/lt90s/rfschub-server/syntect/client"
	"github.com/lt90s/rfschub-server/syntect/proto"
	log "github.com/sirupsen/logrus"
	"strings"
)

type RepositoryService struct {
	store         store.Store
	syncer        *syncer
	syntectClient syntect.SyntectService
}

var (
	errorUrlInvalid   = errors.NewBadRequestError(-1, "repository url invalid")
	errorRepoNotFound = errors.NewNotFoundError(-1, "repository not exist")
	errorInSync       = errors.NewServiceUnavailable(int(proto.RepositoryErrorCode_InSync), "in sync")
)

func NewRepositoryService(config config.RepositoryConfig, s store.Store) proto.RepositoryServiceHandler {
	return &RepositoryService{
		store:         s,
		syncer:        newSyncer(config, s),
		syntectClient: client.New(client.ServerConfig{ServiceName: config.Syntect}),
	}
}

// get repository's branches and tags
func (r *RepositoryService) NamedCommits(ctx context.Context, req *proto.NamedCommitsRequest, rsp *proto.NamedCommitsResponse) error {
	repoUrl, ok := url.NormalizeRepoUrl(req.Url)
	if !ok {
		return errorUrlInvalid
	}

	commits, err := r.store.GetRepository(ctx, repoUrl)
	if err == store.ErrorRepositoryNotFound {
		err = r.syncer.syncRepository(ctx, repoUrl)
		if err != nil {
			if err == ErrRepositoryNotFound {
				return errorRepoNotFound
			} else if err == ErrInSync {
				return errorInSync
			} else {
				return errors.NewInternalError(-1, err.Error())
			}
		} else {
			return errorInSync
		}
	}

	if err != nil {
		return errors.NewInternalError(-1, err.Error())
	}

	rsp.Commits = make([]*proto.NamedCommit, 0, len(commits))
	for _, commit := range commits {
		rsp.Commits = append(rsp.Commits, &proto.NamedCommit{
			Name:   commit.Name,
			Hash:   commit.Hash,
			Branch: commit.Branch,
		})
	}
	return nil
}

func (r *RepositoryService) IsRepositoryExist(ctx context.Context, req *proto.RepositoryExistRequest, rsp *proto.RepositoryExistResponse) error {
	exist, err := r.store.RepositoryExist(ctx, req.Url, req.Hash)
	if err != nil {
		return errors.NewInternalError(-1, err.Error())
	}

	rsp.Exist = exist
	return nil
}

//Get file list under the specified directory
//in case the directories may not be synced from git service
//`RepositoryErrorCode_InSync` is returned
//client should be retry later as syncing is in progress
func (r *RepositoryService) Directory(ctx context.Context, req *proto.DirectoryRequest, rsp *proto.DirectoryResponse) error {
	log.Debugf("[Directory]: url=%s hash=%s", req.Url, req.Hash)

	repoUrl, ok := url.NormalizeRepoUrl(req.Url)
	if !ok {
		return errorUrlInvalid
	}

	synced, entries, err := r.store.GetDirectoryEntries(ctx, repoUrl, req.Hash, req.Path)

	if err != nil {
		if err == store.ErrorDirectoryNotFound {
			return errorRepoNotFound
		}
		return errors.NewInternalError(-1, err.Error())
	}

	if !synced {
		// it's ok to ignore the error
		// the client should retry later
		_ = r.syncer.syncDirectories(ctx, repoUrl, req.Hash)
		return errorInSync
	}

	rsp.Entries = make([]*proto.DirectoryEntry, len(entries))
	for i, entry := range entries {
		rsp.Entries[i] = &proto.DirectoryEntry{
			File: entry.File,
			Dir:  entry.Dir,
		}
	}
	return nil
}

func (r *RepositoryService) Blob(ctx context.Context, req *proto.BlobRequest, rsp *proto.BlobResponse) error {
	log.Debugf("[Blob]: url=%s hash=%s file=%s", req.Url, req.Hash, req.Path)
	repoUrl, ok := url.NormalizeRepoUrl(req.Url)
	if !ok {
		return errorUrlInvalid
	}

	blob, err := r.store.GetBlob(ctx, repoUrl, req.Hash, req.Path)
	if err != nil {
		if err == store.ErrorBlobNotFound {
			return errors.NewNotFoundError(-1, "blob not found")
		}
		return errors.NewInternalError(-1, err.Error())
	}

	if !blob.Synced {
		_ = r.syncer.syncBlob(ctx, repoUrl, req.Hash, req.Path)
		return errorInSync
	}

	// TODO: save renderedCode ?
	if blob.Plain {
		// TODO: extend this list
		if strings.HasSuffix(req.Path, ".md") {
			rsp.Content = blob.Content
		} else {
			result, err := r.syntectClient.RenderCode(ctx, &syntect.RenderCodeRequest{
				File:  req.Path,
				Theme: syntect.CodeTheme_SolarizedLight,
				Code:  blob.Content,
			})
			if err != nil || result.RenderedCode == "" {
				rsp.Content = blob.Content
			} else {
				rsp.Content = result.RenderedCode
			}
		}
	}
	rsp.Plain = blob.Plain
	return nil
}

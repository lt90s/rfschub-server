package service

import (
	"context"
	"github.com/lt90s/rfschub-server/common/errors"
	"github.com/lt90s/rfschub-server/common/url"
	"github.com/lt90s/rfschub-server/gits/config"
	proto "github.com/lt90s/rfschub-server/gits/proto"
	log "github.com/sirupsen/logrus"
)

type GitService struct {
	commander *gitCommander
}

func New(confer config.GitConfer) *GitService {
	return &GitService{
		commander: newGitCommander(confer.GetCommandConf()),
	}
}

var (
	errRepositoryUrlInvalid = errors.NewBadRequestError(int(proto.ErrorCode_RepoUrlInvalid), "repository url invalid")
)

func (g *GitService) Clone(ctx context.Context, req *proto.CloneRequest, rsp *proto.CloneResponse) error {
	log.Debugf("Prepare to clone: url=%s", req.Url)
	repoUrl, ok := url.NormalizeRepoUrl(req.Url)
	if !ok {
		return errRepositoryUrlInvalid
	}
	err := g.commander.clone(ctx, repoUrl)
	if err != nil {
		if err == errorGitBusy {
			return errors.NewServiceUnavailable(int(proto.ErrorCode_GitsBusy), err.Error())
		} else if err == errorRepositoryCloning {
			return errors.NewServiceUnavailable(int(proto.ErrorCode_RepoCloning), err.Error())
		} else {
			return errors.NewInternalError(-1, err.Error())
		}
	}
	return nil
}

// get clone status
func (g GitService) GetCloneStatus(ctx context.Context, req *proto.GetCloneStatusRequest, rsp *proto.GetCloneStatusResponse) error {
	log.Debugf("query clone status: url=%s", req.Url)
	repoUrl, ok := url.NormalizeRepoUrl(req.Url)
	if !ok {
		return errRepositoryUrlInvalid
	}

	status, progress := g.commander.cloneStatus(ctx, repoUrl)
	log.Debugf("query clone status: url=%s status=%d progress=%s", req.Url, status, progress)
	rsp.Status = status
	rsp.Progress = progress
	return nil
}

func (g GitService) GetNamedCommits(ctx context.Context, req *proto.GetNamedCommitsRequest, rsp *proto.GetNamedCommitsResponse) error {
	log.Debugf("query branches and tags: url=%s", req.Url)
	repoUrl, ok := url.NormalizeRepoUrl(req.Url)
	if !ok {
		return errRepositoryUrlInvalid
	}
	namedCommits, err := g.commander.getNamedCommits(ctx, repoUrl)
	if err != nil {
		log.Warnf("query branches and tags: url=%s error=%s", req.Url, err.Error())
		if err == errorRepositoryNotExist {
			return errors.NewNotFoundError(int(proto.ErrorCode_RepoNotExist), "repository not exist")
		} else {
			return errors.NewInternalError(-1, err.Error())
		}
	}
	log.Debugf("query branches and tags: url=%s commits=%s", req.Url, namedCommits)
	rsp.Commits = namedCommits
	return nil
}

func (g GitService) GetRepositoryFiles(ctx context.Context, req *proto.GetRepositoryFilesRequest, rsp *proto.GetRepositoryFilesResponse) error {
	log.Debugf("get repository files: url=%s commit=%s", req.Url, req.Commit)
	repoUrl, ok := url.NormalizeRepoUrl(req.Url)
	if !ok {
		return errRepositoryUrlInvalid
	}
	entries, err := g.commander.getRepositoryFiles(ctx, repoUrl, req.Commit)
	if err != nil {
		log.Warnf("get repository files: url=%s commit=%s err=%s", req.Url, req.Commit, err.Error())
		if err == errorRepositoryNotExist {
			return errors.NewNotFoundError(int(proto.ErrorCode_RepoNotExist), "repository not exist")
		} else {
			return errors.NewInternalError(-1, err.Error())
		}
	}
	log.Debugf("get repository files: url=%s commit=%s entries=%v", req.Url, req.Commit, entries)
	rsp.Entries = entries
	return nil
}

func (g GitService) GetRepositoryBlob(ctx context.Context, req *proto.GetRepositoryBlobRequest, rsp *proto.GetRepositoryBlobResponse) error {
	log.Debugf("get repository blob: url=%s commit=%s file=%s", req.Url, req.Commit, req.File)
	repoUrl, ok := url.NormalizeRepoUrl(req.Url)
	if !ok {
		return errRepositoryUrlInvalid
	}
	plain, content, err := g.commander.getRepositoryBlob(ctx, repoUrl, req.Commit, req.File)
	if err != nil {
		log.Warnf("get repository blob: url=%s commit=%s file=%s err=%s", req.Url, req.Commit, req.File, err.Error())
		if err == errorRepositoryNotExist {
			return errors.NewNotFoundError(int(proto.ErrorCode_RepoNotExist), "repository not exist")
		} else {
			return errors.NewInternalError(-1, err.Error())
		}
		return nil
	}
	log.Debugf("get repository blob: url=%s commit=%s file=%s content=%s plain=%v", req.Url, req.Commit, req.File, content, plain)
	rsp.Content = content
	rsp.Plain = plain
	return nil
}

type archiveWriter struct {
	stream proto.Gits_ArchiveStream
}

func (w archiveWriter) Write(p []byte) (int, error) {
	err := w.stream.Send(&proto.ArchiveResponse{Data: p})
	return len(p), err
}

func (g GitService) Archive(ctx context.Context, req *proto.ArchiveRequest, stream proto.Gits_ArchiveStream) error {
	log.Debugf("archive repository: url=%s commit=%s", req.Url, req.Commit)

	repoUrl, ok := url.NormalizeRepoUrl(req.Url)
	if !ok {
		//return errors.New("repository url invalid")
	}

	err := g.commander.archive(ctx, repoUrl, req.Commit, archiveWriter{stream: stream})
	if err != nil {
		log.Warnf("get repository archive: url=%s commit=%s err=%s", req.Url, req.Commit, err.Error())

		if err == errorRepositoryNotExist {
			return errors.NewNotFoundError(int(proto.ErrorCode_RepoNotExist), "repository not exist")
		} else {
			return errors.NewInternalError(-1, err.Error())
		}
	}

	return nil
}

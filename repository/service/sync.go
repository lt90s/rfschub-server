package service

import (
	"context"
	"errors"
	"github.com/lt90s/rfschub-server/gits/client"
	"github.com/lt90s/rfschub-server/gits/proto"
	"github.com/lt90s/rfschub-server/repository/config"
	"github.com/lt90s/rfschub-server/repository/store"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"strings"
	"sync"
	"time"
)

var (
	ErrInSync             = errors.New("already in sync")
	ErrSyncerBusy         = errors.New("syncer busy")
	ErrRepositoryNotFound = errors.New("repository not found")
)

type syncer struct {
	store     store.Store
	gitClient gits.GitsService
	sema      *semaphore.Weighted
	mutext    sync.RWMutex
	tasks     map[string]context.CancelFunc
	wg        *sync.WaitGroup
	timeout   time.Duration
}

func newSyncer(config config.RepositoryConfig, store store.Store) *syncer {
	gitConf := client.ServerConfig{
		ServiceName: config.Syncer.Gits,
	}
	return &syncer{
		store:     store,
		gitClient: client.New(gitConf),
		sema:      semaphore.NewWeighted(int64(config.Syncer.Concurrency)),
		tasks:     make(map[string]context.CancelFunc),
		wg:        &sync.WaitGroup{},
		timeout:   time.Duration(config.Syncer.Timeout) * time.Second,
	}
}

// 1. check if already in sync
// 2. acquire semaphore
// 3. acquire task mutex
// 4. check again if already in sync
// 5. add task
// 6. add WaitGroup
func (s *syncer) prepareSync(url, commit, file string) (context.Context, error) {
	key := url + "@" + commit
	s.mutext.RLock()
	if _, ok := s.tasks[key]; ok {
		s.mutext.RUnlock()
		return nil, ErrInSync
	}
	s.mutext.RUnlock()

	if !s.sema.TryAcquire(1) {
		return nil, ErrSyncerBusy
	}

	s.mutext.Lock()
	if _, ok := s.tasks[key]; ok {
		s.mutext.Unlock()
		return nil, ErrInSync
	}
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)

	s.tasks[key] = cancel
	s.mutext.Unlock()

	s.wg.Add(1)
	return ctx, nil
}

// 1. s.wg.Done
// 2. cancel to release resource
// 3. release semaphore
// 4. delete task
func (s *syncer) finishSync(ctx context.Context, url, commit, file string) {
	key := url + "@" + commit
	cancel := s.tasks[key]
	s.wg.Done()
	cancel()
	s.sema.Release(1)
	s.mutext.Lock()
	delete(s.tasks, key)
	s.mutext.Unlock()
}

// synchronize repository's branches and tags
func (s *syncer) syncRepository(ctx context.Context, url string) error {
	req := gits.GetCloneStatusRequest{
		Url: url,
	}
	rsp, err := s.gitClient.GetCloneStatus(ctx, &req)
	if err != nil {
		return errorRepoNotFound
	}

	if rsp.Status == gits.CloneStatus_Cloning {
		return errorInSync
	}

	if rsp.Status != gits.CloneStatus_Cloned {
		return errorRepoNotFound
	}

	ctx, err = s.prepareSync(url, "", "")
	if err != nil {
		return err
	}
	go s.doSyncRepository(ctx, url)
	return nil
}

func (s *syncer) doSyncRepository(ctx context.Context, url string) {
	log.Debugf("start sync repository: url=%s", url)
	defer s.finishSync(ctx, url, "", "")
	req := &gits.GetNamedCommitsRequest{
		Url: url,
	}
	rsp, err := s.gitClient.GetNamedCommits(ctx, req)
	if err != nil {
		log.Warnf("doSyncRepository: git service error, error=%s", err.Error())
		return
	}

	log.Debugf("finish sync repository: url=%s", url)
	commits := make([]store.NamedCommit, 0, len(rsp.Commits))
	for _, commit := range rsp.Commits {
		commits = append(commits, store.NamedCommit{
			Name:   commit.Name,
			Hash:   commit.Hash,
			Branch: commit.Branch,
		})
	}
	s.store.AddRepository(ctx, url, commits)
}

// synchronize all files
func (s *syncer) syncDirectories(ctx context.Context, url, commit string) error {
	ctx, err := s.prepareSync(url, commit, "")
	if err != nil {
		return err
	}
	go s.doSyncDirectories(ctx, url, commit)
	return nil
}

// get all files from GitService and save them
func (s *syncer) doSyncDirectories(ctx context.Context, url, commit string) {
	log.Debugf("start sync directories: url=%s commit=%s", url, commit)
	defer s.finishSync(ctx, url, commit, "")

	req := gits.GetRepositoryFilesRequest{
		Url:    url,
		Commit: commit,
	}
	rsp, err := s.gitClient.GetRepositoryFiles(ctx, &req)
	if err != nil {
		log.Warnf("sync directory, git service returns error: error=%s", err.Error())
		return
	}

	log.Debugf("finish sync directories: url=%s commit=%s files=%v", url, commit, rsp.Entries)

	err = s.store.SetDirectories(ctx, url, commit, rsp.Entries)
	if err != nil {
		log.Warnf("sync directory, save directory entries error: url=%s commit=%s err=%s", url, commit, err.Error())
	}
}

// synchronize regular file
func (s *syncer) syncBlob(ctx context.Context, url, commit, file string) error {
	ctx, err := s.prepareSync(url, commit, file)
	if err != nil {
		return err
	}
	go s.doSyncBlob(ctx, url, commit, file)
	return nil
}

func (s *syncer) doSyncBlob(ctx context.Context, url, commit, file string) {
	defer s.finishSync(ctx, url, commit, file)
	req := gits.GetRepositoryBlobRequest{
		Url:    url,
		Commit: commit,
		File:   strings.TrimPrefix(file, "/"),
	}
	rsp, err := s.gitClient.GetRepositoryBlob(ctx, &req)
	if err != nil {
		log.Warnf("sync blob, git service returns error: error=%s", err.Error())
		return
	}

	err = s.store.SetBlob(ctx, url, commit, file, rsp.Content, rsp.Plain)
	if err != nil {
		log.Warnf("sync blob, save blob error: url=%s commit=%s file=%s err=%s", url, commit, file, err.Error())
	}
}

func (s *syncer) shutdown() {
	s.mutext.Lock()
	defer s.mutext.Unlock()
	for _, cancel := range s.tasks {
		cancel()
	}
}

func (s *syncer) wait(timeout time.Duration) {
	go func() {
		<-time.NewTimer(timeout).C
		s.shutdown()
	}()
	s.wg.Wait()
}

package service

import (
	"bytes"
	"context"
	"errors"
	"github.com/lt90s/rfschub-server/gits/config"
	proto "github.com/lt90s/rfschub-server/gits/proto"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/semaphore"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	plainFileMaxSize = 256 * 1024
)

var (
	errorRepositoryCloning  = errors.New("repository cloning")
	errorGitBusy            = errors.New("git commander busy")
	errorRepositoryCloned   = errors.New("repository cloned")
	errorRepositoryNotExist = errors.New("repository not exist")
	errorFileNotFound       = errors.New("file not found")
)

type gitCommander struct {
	conf       config.CommandConf
	cloneSem   *semaphore.Weighted
	archiveSem *semaphore.Weighted
	otherSem   *semaphore.Weighted

	wg *sync.WaitGroup

	statusMutex sync.RWMutex
	status      map[string]cloneProgress
}

type cloneProgress struct {
	progress string
	err      error
}

type progressUpdater func(progress string)

type progressWriter struct {
	count   int
	updater progressUpdater
}

// import Writer interface
func (w *progressWriter) Write(p []byte) (int, error) {
	var lastLine []byte
	for i := 0; i < len(p); {
		var nextI, j int
		for j = i; j < len(p); j++ {
			if p[j] == '\r' {
				nextI = j + 1
				if j+1 < len(p) && p[j] == '\n' {
					nextI = j + 2
				}
				break
			} else if p[j] == '\n' {
				nextI = j + 1
				break
			}
		}
		if j >= len(p) {
			nextI = j
		}
		lastLine = p[i:j]
		w.count += 1
		i = nextI
	}
	// scanner do not handle \r
	//scanner := bufio.NewScanner(bytes.NewBuffer(p))
	//var lastLine string
	//for scanner.Scan() {
	//	lastLine = scanner.Text()
	//	w.count += 1
	//}

	// send progress until at least we got more than 10 lines
	if w.count >= 20 {
		w.count = 0
		w.updater(string(lastLine))
	}

	return len(p), nil
}

type lineWriter struct {
	buf   []byte
	lines []string
	r, w  int
}

// implement io.Writer
func (lw *lineWriter) Write(p []byte) (int, error) {
	n := len(p)
	lw.buf = append(lw.buf, p...)
	lw.w += n
	for {
		i := bytes.IndexByte(lw.buf[lw.r:lw.w], '\r')
		if i == -1 {
			i = bytes.IndexByte(lw.buf[lw.r:lw.w], '\n')
			if i == -1 {
				break
			}
		}

		line := lw.buf[lw.r : lw.r+i+1]
		lineLength := len(line)

		if lineLength > 1 && line[lineLength-2] == '\r' {
			line = line[:lineLength-2]
		} else {
			line = line[:lineLength-1]
		}
		lw.lines = append(lw.lines, string(line))
		lw.r += i + 1
	}
	if lw.r != 0 {
		lw.buf = lw.buf[lw.r:lw.w]
		lw.w = len(lw.buf)
		lw.r = 0
	}
	return n, nil
}

func newGitCommander(conf config.CommandConf) *gitCommander {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, conf.Path, "--version")
	err := cmd.Run()
	if err != nil {
		log.Panicf("git bin path not exist: path=%s", conf.Path)
	}

	cmd = exec.CommandContext(ctx, conf.Path, "config", "--global", "core.pager", "")
	err = cmd.Run()
	if err != nil {
		log.Panicf("git config --global core.pager error: %s\n", err.Error())
	}

	err = os.MkdirAll(conf.Data, 0755)
	if err != nil {
		log.Panicf("create directory error: dir=%s error=%s", conf.Data, err.Error())
	}

	return &gitCommander{
		conf:       conf,
		cloneSem:   semaphore.NewWeighted(conf.Concurrency.Clone),
		archiveSem: semaphore.NewWeighted(conf.Concurrency.Archive),
		otherSem:   semaphore.NewWeighted(conf.Concurrency.Other),
		status:     make(map[string]cloneProgress, conf.Concurrency.Clone),
		wg:         &sync.WaitGroup{},
	}
}

func (g *gitCommander) urlToLocal(repoUrl string) (string, error) {
	u, err := url.Parse(repoUrl)
	if err != nil {
		return "", err
	}
	p := strings.TrimPrefix(u.Path, "/")
	if strings.Count(p, "/") > 1 {
		return "", errors.New("invalid repository url")
	}

	return path.Join(g.conf.Data, p), nil
}

func (g *gitCommander) prepareClone(ctx context.Context, url string, dst string) error {
	// first check if already cloned
	_, err := os.Stat(dst)
	if err == nil {
		return errorRepositoryCloned
	}
	// then check if url is been cloning
	g.statusMutex.Lock()
	defer g.statusMutex.Unlock()

	if _, ok := g.status[url]; ok {
		log.Debugf("prepareClone: already cloning: url=%s", url)
		return errorRepositoryCloning
	}

	g.status[url] = cloneProgress{progress: "prepare cloning"}

	log.Debug(g.status)

	return nil
}

func (g *gitCommander) isRepositoryCloned(url string) bool {
	dir, err := g.urlToLocal(url)
	if err != nil {
		return false
	}

	_, err = os.Stat(dir)
	if err == nil {
		return true
	}
	return false
}

func (g *gitCommander) clone(ctx context.Context, url string) error {
	dstDir, err := g.urlToLocal(url)
	if err != nil {
		return err
	}

	err = g.prepareClone(ctx, url, dstDir)
	if err != nil {
		return err
	}

	// try acquire sema
	if !g.cloneSem.TryAcquire(1) {
		return errorGitBusy
	}

	log.Debugf("clone timeout setting: %ds", g.conf.CloneTimeout)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(g.conf.CloneTimeout)*time.Second)

	updater := func(progress string) {
		log.Debugf("update progress: url=%s, progress=%s", url, progress)
		// need lock here?
		g.status[url] = cloneProgress{progress: progress}
	}

	// run git command
	g.wg.Add(1)
	go func() {
		defer func() {
			g.wg.Done()
			cancel()
			g.cloneSem.Release(1)
			g.statusMutex.Lock()
			delete(g.status, url)
			g.statusMutex.Unlock()
			os.Remove(dstDir + "_tmp")
		}()
		now := time.Now()
		parentDir := path.Dir(dstDir)
		err := os.MkdirAll(parentDir, 0755)
		if err != nil {
			return
		}
		tmpDir := dstDir + "_tmp"
		cmd := exec.CommandContext(ctx, g.conf.Path, "clone", "--mirror", "--progress", url, tmpDir)
		pw := &progressWriter{updater: updater}
		cmd.Stderr = pw
		cmd.Stdout = pw

		err = cmd.Run()
		if err == nil {
			err = os.Rename(tmpDir, dstDir)
		}
		if err != nil {
			// need lock here?
			log.Warnf("clone repository error: url=%s error=%s", url, err.Error())
			g.status[url] = cloneProgress{err: err}
		}
		log.Debugf("clone repository success: url=%s dir=%s time=%v", url, dstDir, time.Since(now))
	}()

	return nil
}

func (g *gitCommander) cloneStatus(ctx context.Context, url string) (status proto.CloneStatus, progress string) {
	status = proto.CloneStatus_Unknown
	if g.isRepositoryCloned(url) {
		status = proto.CloneStatus_Cloned
		return
	}

	g.statusMutex.RLock()
	defer g.statusMutex.RUnlock()
	if p, ok := g.status[url]; ok {
		progress = p.progress
		status = proto.CloneStatus_Cloning
		return
	}

	dstDir, err := g.urlToLocal(url)
	if err != nil {
		status = proto.CloneStatus_Unknown
		return
	}

	_, err = os.Stat(dstDir)
	if err == nil {
		status = proto.CloneStatus_Cloned
	} else {
		status = proto.CloneStatus_Unknown
	}

	return
}

func (g *gitCommander) getNamedCommits(ctx context.Context, url string) (commits []*proto.NamedCommit, err error) {
	// try acquire sema
	if !g.otherSem.TryAcquire(1) {
		err = errorGitBusy
		return
	}
	defer g.otherSem.Release(1)

	if !g.isRepositoryCloned(url) {
		err = errorRepositoryNotExist
		return
	}
	dir, err := g.urlToLocal(url)
	lw := &lineWriter{}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(g.conf.DefaultTimeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, g.conf.Path, "show-ref", "--heads", "--tags")
	cmd.Dir = dir
	cmd.Stdout = lw

	err = cmd.Run()

	if err == nil {
		for _, line := range lw.lines {
			parts := strings.Fields(line)
			if len(parts) != 2 {
				continue
			}

			var prefix string
			var isBranch bool
			if strings.HasPrefix(parts[1], "refs/heads/") {
				prefix = "refs/heads/"
				isBranch = true
			} else if strings.HasPrefix(parts[1], "refs/tags/") {
				prefix = "refs/tags/"
			} else {
				continue
			}

			commits = append(commits, &proto.NamedCommit{
				Hash:   parts[0],
				Name:   strings.TrimPrefix(parts[1], prefix),
				Branch: isBranch,
			})
		}
	}
	return
}

type fileEntryWriter struct {
	lw      lineWriter
	entries []*proto.FileEntry
}

var spaceRegex = regexp.MustCompile("[ \t]")

func (f *fileEntryWriter) Write(p []byte) (int, error) {
	n, err := f.lw.Write(p)
	for _, line := range f.lw.lines {
		//log.Logger.Debug("line", "line", line)
		parts := spaceRegex.Split(line, -1)
		if len(parts) < 4 {
			continue
		}

		var isDir bool
		if parts[1] == "tree" {
			isDir = true
		}
		f.entries = append(f.entries, &proto.FileEntry{
			File: strings.Trim(parts[3], ""),
			Dir:  isDir,
		})
	}
	f.lw.lines = f.lw.lines[:0]
	return n, err
}

func (g *gitCommander) getRepositoryFiles(ctx context.Context, url string, commit string) (entries []*proto.FileEntry, err error) {
	// try acquire sema
	if !g.otherSem.TryAcquire(1) {
		err = errorGitBusy
		return
	}
	defer g.otherSem.Release(1)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(g.conf.DefaultTimeout)*time.Second)
	defer cancel()

	if !g.isRepositoryCloned(url) {
		err = errorRepositoryNotExist
		return
	}
	dir, _ := g.urlToLocal(url)

	args := []string{"ls-tree", "-r", "-t", commit}
	cmd := exec.CommandContext(ctx, g.conf.Path, args...)
	few := &fileEntryWriter{}
	cmd.Stdout = few
	cmd.Dir = dir

	err = cmd.Run()
	if err == nil {
		entries = few.entries
	}
	log.Debugf("RepositoryFiles: dir=%s commit=%s error=%v entries=%v", dir, commit, err, entries)
	return
}

func (g *gitCommander) wait() {
	g.wg.Wait()
}

type contentWriter struct {
	buffer  bytes.Buffer
	cancel  context.CancelFunc
	binary  bool
	checked bool
}

// check if content is binary
// if content length is bigger than plainFileMaxSize(default 16KB)
// treat it as binary
func (cw *contentWriter) Write(p []byte) (int, error) {
	if !cw.checked {
		for i := 0; i < 1024 && i < len(p); i++ {
			if p[i] == 0 {
				cw.binary = true
				cw.cancel()
				return len(p), nil
			}
		}
		cw.checked = true
	}

	if cw.buffer.Len()+len(p) > plainFileMaxSize {
		// treat as binary
		cw.binary = true
		cw.cancel()
		return len(p), nil
	}
	return cw.buffer.Write(p)
}

func (g *gitCommander) getRepositoryBlob(ctx context.Context, url, commit, file string) (plain bool, content string, err error) {
	// try acquire sema
	if !g.otherSem.TryAcquire(1) {
		err = errorGitBusy
		return
	}
	defer g.otherSem.Release(1)

	ctx, cancel := context.WithTimeout(ctx, time.Duration(g.conf.DefaultTimeout)*time.Second)
	defer cancel()

	if !g.isRepositoryCloned(url) {
		err = errorRepositoryNotExist
		return
	}
	dir, _ := g.urlToLocal(url)

	cw := &contentWriter{
		cancel: cancel,
	}
	args := []string{"show", commit + ":" + file}
	cmd := exec.CommandContext(ctx, g.conf.Path, args...)

	cmd.Dir = dir
	cmd.Stdout = cw

	err = cmd.Run()
	if err != nil {
		if cw.binary {
			err = nil
			return
		}
		err = errorFileNotFound
	} else if !cw.binary {
		plain = true
		content = cw.buffer.String()
		if strings.HasPrefix(content, "tree "+commit+":"+file) {
			err = errorFileNotFound
			content = ""
		}
	}
	return
}

func (g *gitCommander) archive(ctx context.Context, url, commit string, writer io.Writer) error {
	// try acquire archive sema
	if !g.archiveSem.TryAcquire(1) {
		return errorGitBusy
	}
	defer g.archiveSem.Release(1)

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(g.conf.ArchiveTimeout)*time.Second)
	defer cancel()

	if !g.isRepositoryCloned(url) {
		return errorRepositoryNotExist
	}
	dir, _ := g.urlToLocal(url)

	cmd := exec.CommandContext(ctx, g.conf.Path,
		"archive",
		"--worktree-attributes",
		"--format=tar",
		//"-0",
		commit,
		"--")
	cmd.Dir = dir
	cmd.Stdout = writer
	//cmd.Stderr = writer

	return cmd.Run()
}

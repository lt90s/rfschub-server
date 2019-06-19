package service

import (
	"archive/tar"
	"bytes"
	"context"
	"github.com/lt90s/rfschub-server/gits/client"
	"github.com/lt90s/rfschub-server/gits/proto"
	"github.com/lt90s/rfschub-server/index/config"
	"github.com/lt90s/rfschub-server/index/store"
	log "github.com/sirupsen/logrus"
	"io"
	"time"
)

type indexRequest struct {
	url  string
	hash string
}

type indexResult struct {
	url     string
	hash    string
	success bool
}

type indexer struct {
	reqChan   <-chan indexRequest
	resChan   chan<- indexResult
	timeout   time.Duration
	store     store.Store
	stop      chan struct{}
	gitClient gits.GitsService
	maxSize   int64
	buffer    []byte
	cmds      []*commander
}

func newIndexer(config config.IndexConfig, reqChan <-chan indexRequest, resChan chan<- indexResult, store store.Store) *indexer {
	gitClient := client.New(client.ServerConfig{ServiceName: config.Gits.Name})
	cmds := make([]*commander, config.Concurrency)
	for i := 0; i < config.Concurrency; i++ {
		cmds[i] = newCommander(config.Path)
	}
	indexer := &indexer{
		reqChan:   reqChan,
		resChan:   resChan,
		store:     store,
		timeout:   time.Duration(config.Timeout) * time.Second,
		gitClient: gitClient,
		stop:      make(chan struct{}, config.Concurrency),
		maxSize:   config.Size,
		buffer:    make([]byte, config.Size),
		cmds:      cmds,
	}

	for i := 0; i < config.Concurrency; i++ {
		go indexer.start(i)
	}

	return indexer
}

func (indexer *indexer) start(i int) {
	for {
		select {
		case <-indexer.stop:
			indexer.cmds[i].stop()
			return
		case task, ok := <-indexer.reqChan:
			if !ok {
				return
			}
			log.Debugf("new index task: url=%s hash=%s", task.url, task.hash)
			ctx, cancel := context.WithTimeout(context.Background(), indexer.timeout)
			err := indexer.indexRepository(ctx, task, i)
			cancel()
			log.Debugf("finish index task: url=%s hash=%s err=%v", task.url, task.hash, err)
			success := true
			if err != nil {
				success = false
			}
			_ = indexer.store.SetTaskState(context.Background(), task.url, task.hash, success)
			indexer.resChan <- indexResult{
				url:     task.url,
				hash:    task.hash,
				success: success,
			}
		}
	}
}

type gitsArchiveReader struct {
	buffer bytes.Buffer
	as     gits.Gits_ArchiveService
}

func (gas *gitsArchiveReader) Read(p []byte) (int, error) {
	if gas.buffer.Len() > 0 {
		n, err := gas.buffer.Read(p)
		return n, err
	}
	rsp, err := gas.as.Recv()
	if err != nil {
		log.Warnf("recv archive data error: error=%s", err.Error())
		return 0, err
	}

	_, err = gas.buffer.Write(rsp.Data)
	if err != nil {
		return 0, err
	}

	n, err := gas.buffer.Read(p)
	return n, err
}

func (indexer *indexer) indexRepository(ctx context.Context, task indexRequest, index int) error {
	now := time.Now()
	req := &gits.ArchiveRequest{Url: task.url, Commit: task.hash}
	as, err := indexer.gitClient.Archive(ctx, req)
	if err != nil {
		log.Warnf("[indexRepository] git client archive returns error: %s", err.Error())
		return err
	}
	reader := &gitsArchiveReader{as: as}

	buffer := indexer.buffer
	tarReader := tar.NewReader(reader)
	for {
		var hdr *tar.Header
		hdr, err = tarReader.Next()
		if err == io.EOF {
			err = nil
			break
		}
		if err != nil {
			log.Warnf("[indexRepository] tarReader.Next() returns error: %s", err.Error())
			return err
		}

		// check if not regular
		if hdr.Typeflag != tar.TypeReg {
			continue
		}

		if hdr.Size > indexer.maxSize {
			continue
		}

		// check if already indexed
		// as we index a file at a time, some files may be indexed already
		// TODO: ask store if this file is indexed

		n, err := tarReader.Read(buffer[:1024])
		if err != nil && err != io.EOF {
			log.Warnf("[indexRepository] tarReader.Read returns error: %s", err.Error())
			return err
		}

		// check if binary
		if bytes.IndexByte(buffer[:n], 0) != -1 {
			log.Debugf("[indexRepository] ignore binary file, name=%s", hdr.Name)
			continue
		}

		ok, err := indexer.store.RepositoryFileIndexed(ctx, task.url, task.hash, hdr.Name)
		if err != nil {
			log.Warnf("[indexRepository] check RepositoryFileIndexed failed: url=%s hash=%s name=%s error=%v", task.url, task.hash, hdr.Name, err)
		}

		// already indexed
		if ok {
			continue
		}

		log.Debugf("[indexRepository] start to index file: name=%s size=%d", hdr.Name, hdr.Size)

		// read it all
		_, err = tarReader.Read(buffer[n:])
		if err != nil && err != io.EOF {
			return err
		}

		entries, err := indexer.cmds[index].indexFile(hdr.Name, buffer[:hdr.Size])
		if err != nil {
			log.Warnf("[indexRepository] index file error: %s", err.Error())
		}

		err = indexer.saveResponseEntries(ctx, task.url, task.hash, entries, buffer)
		if err != nil {
			log.Warnf("[indexRepository] save repository file index entries error: %s", err.Error())
			break
		}
	}
	log.Debugf("[indexRepository] finish indexing: url=%s commit=%s time=%v", req.Url, req.Commit, time.Since(now))

	return nil
}

func (indexer *indexer) saveResponseEntries(ctx context.Context, url, hash string, entries []ResponseEntry, content []byte) error {
	liner := newLiner(content)
	indexEntries := make([]store.IndexEntry, 0, len(entries))
	for _, entry := range entries {
		var lineBefore, lineAfter string
		line, err := liner.getLine(entry.Line)
		if err != nil {
			return err
		}
		lineAfter, err = liner.getLine(entry.Line + 1)
		if err == errLineNotExist {
			lineBefore, _ = liner.getLine(entry.Line - 2)
		}

		lb, _ := liner.getLine(entry.Line - 1)
		if lineBefore != "" {
			lineBefore += "\n" + lb
		} else {
			lineBefore = lb
		}

		indexEntry := store.IndexEntry{
			Url:        url,
			Hash:       hash,
			File:       entry.Path,
			Name:       entry.Name,
			Pattern:    entry.Pattern,
			Language:   entry.Language,
			LineNumber: entry.Line,
			Line:       line,
			LineBefore: lineBefore,
			LineAfter:  lineAfter,
			Kind:       entry.Kind,
			Scope:      entry.Scope,
			ScopeKind:  entry.ScopeKind,
		}

		indexEntries = append(indexEntries, indexEntry)
	}
	if len(indexEntries) == 0 {
		return nil
	}
	return indexer.store.AddFileIndexEntries(ctx, indexEntries)
}

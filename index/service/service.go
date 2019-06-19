package service

import (
	"context"
	"github.com/lt90s/rfschub-server/common/errors"
	"github.com/lt90s/rfschub-server/index/config"
	proto "github.com/lt90s/rfschub-server/index/proto"
	"github.com/lt90s/rfschub-server/index/store"
	log "github.com/sirupsen/logrus"
	"sync"
)

type indexService struct {
	concurrency int
	mu          sync.RWMutex
	tasks       map[string]string
	indexers    []*indexer
	stop        chan struct{}
	reqChan     chan indexRequest
	resChan     chan indexResult
	store       store.Store
}

func NewIndexService(store store.Store) proto.IndexHandler {

	concurrency := config.DefaultConfig.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}
	reqChan := make(chan indexRequest, concurrency)
	resChan := make(chan indexResult, concurrency)

	indexers := make([]*indexer, 0, concurrency)
	for i := 0; i < concurrency; i++ {
		indexers = append(indexers, newIndexer(config.DefaultConfig, reqChan, resChan, store))
	}

	service := &indexService{
		concurrency: concurrency,
		tasks:       make(map[string]string, config.DefaultConfig.Concurrency),
		indexers:    indexers,
		stop:        make(chan struct{}, 1),
		reqChan:     reqChan,
		resChan:     resChan,
		store:       store,
	}

	go service.checkResult()
	return service
}

func (service *indexService) checkResult() {
	for {
		select {
		case <-service.stop:
			return
		case res := <-service.resChan:
			service.mu.Lock()
			delete(service.tasks, res.hash)
			service.mu.Unlock()
		}
	}
}

func (service *indexService) IndexRepository(ctx context.Context, req *proto.IndexRepositoryRequest, rsp *proto.IndexRepositoryResponse) error {
	log.Debugf("[IndexRepository]: url=%s hash=%s", req.Url, req.Hash)
	// check if already indexed
	indexed, err := service.store.RepositoryIndexed(ctx, req.Url, req.Hash)
	if err != nil {
		log.Warnf("[IndexRepository] get repository indexed error: url=%s hash=%s err=%s", req.Url, req.Hash, err.Error())
		return err
	}

	if indexed {
		log.Debugf("[IndexRepository] already indexed: url=%s hash=%s", req.Url, req.Hash)
		rsp.Indexed = true
		return nil
	}

	service.mu.Lock()
	if _, ok := service.tasks[req.Hash]; ok {
		log.Debugf("[IndexRepository] indexing: url=%s hash=%s", req.Url, req.Hash)
		service.mu.Unlock()
		return nil
	}

	if len(service.tasks) >= service.concurrency {
		log.Debugf("[IndexRepository] indexer busy: url=%s hash=%s", req.Url, req.Hash)
		service.mu.Unlock()
		return errors.NewServiceUnavailable(-1, "indexer busy")
	}

	service.tasks[req.Hash] = req.Url
	service.mu.Unlock()

	err = service.store.NewIndexTask(ctx, req.Url, req.Hash)
	if err != nil {
		log.Warnf("[IndexRepository] new index task error: url=%s hash=%s error=%v", req.Url, req.Hash, err)
		service.mu.Lock()
		delete(service.tasks, req.Hash)
		service.mu.Unlock()
		return err
	}

	request := indexRequest{
		url:  req.Url,
		hash: req.Hash,
	}
	// should not block here
	service.reqChan <- request
	return nil
}

func (service *indexService) IndexStatus(ctx context.Context, req *proto.IndexStatusRequest, rsp *proto.IndexStatusResponse) error {
	service.mu.RLock()
	if url, ok := service.tasks[req.Hash]; ok {
		if url == req.Url {
			service.mu.RUnlock()
			rsp.Status = proto.StatusCode_StatusIndexing
			return nil
		}
	}

	indexed, err := service.store.RepositoryIndexed(ctx, req.Url, req.Hash)
	if err != nil {
		return err
	}

	if indexed {
		rsp.Status = proto.StatusCode_StatusIndexed
	} else {
		rsp.Status = proto.StatusCode_StatusUnIndexed
	}
	return nil
}

func (service *indexService) SearchSymbol(ctx context.Context, req *proto.SearchSymbolRequest, rsp *proto.SearchSymbolResponse) error {
	log.Debugf("[SearchSymbol] url=%s hash=%s symbol=%s", req.Url, req.Hash, req.Symbol)
	symbols, err := service.store.FindSymbols(ctx, req.Url, req.Hash, req.Symbol)
	if err != nil {
		return err
	}

	rsp.Symbols = make([]*proto.SymbolResult, 0, len(symbols))
	for _, symbol := range symbols {
		rsp.Symbols = append(rsp.Symbols, &proto.SymbolResult{
			File:       symbol.File,
			LineNumber: int32(symbol.LineNumber),
			Line:       symbol.Line,
			LineBefore: symbol.LineBefore,
			LineAfter:  symbol.LineAfter,
			Kind:       symbol.Kind,
		})
	}
	return nil
}

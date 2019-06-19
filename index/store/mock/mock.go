package mock

import (
	"context"
	"fmt"
	"github.com/lt90s/rfschub-server/index/config"
	"github.com/lt90s/rfschub-server/index/store"
	"sync"
	"time"
)

type mockStore struct {
	mutex   sync.RWMutex
	tasks   []indexTask
	iMutex  sync.RWMutex
	indexes []store.IndexEntry
}

type indexTask struct {
	url       string
	hash      string
	success   bool
	createdAt int64
}

func NewMockStore() store.Store {
	return &mockStore{
		tasks:   make([]indexTask, 0),
		indexes: make([]store.IndexEntry, 0),
	}
}

func (m *mockStore) NewIndexTask(ctx context.Context, url, hash string) error {
	m.mutex.Lock()
	timestamp := time.Now().Unix()
	defer m.mutex.Unlock()
	for _, task := range m.tasks {
		if task.url == url && task.hash == hash {
			if timestamp-task.createdAt > int64(config.DefaultConfig.Timeout) {
				task.createdAt = timestamp
				return nil
			} else {
				return store.ErrIndexTaskExist
			}
		}
	}
	m.tasks = append(m.tasks, indexTask{
		url:       url,
		hash:      hash,
		createdAt: timestamp,
	})
	return nil
}

func (m *mockStore) SetTaskState(ctx context.Context, url, hash string, success bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for idx, task := range m.tasks {
		if task.url == url && task.hash == hash {
			if success {
				m.tasks[idx].success = true
			} else {
				m.tasks[idx].createdAt = 0
			}
		}
	}
	return nil
}

func (m *mockStore) RepositoryIndexed(ctx context.Context, url, hash string) (bool, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	for _, task := range m.tasks {
		if task.url == url && task.hash == hash {
			return task.success, nil
		}
	}
	return false, nil
}

func (m *mockStore) RepositoryFileIndexed(ctx context.Context, url, hash, file string) (bool, error) {
	m.iMutex.RLock()
	defer m.iMutex.RUnlock()
	for _, index := range m.indexes {
		if index.Url == url && index.Hash == hash && index.File == file {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockStore) AddFileIndexEntries(ctx context.Context, entries []store.IndexEntry) error {
	m.iMutex.Lock()
	defer m.iMutex.Unlock()
	m.indexes = append(m.indexes, entries...)
	fmt.Println(entries)
	return nil
}

func (m *mockStore) FindSymbols(ctx context.Context, url, hash, name string) (symbols []store.Symbol, err error) {
	m.iMutex.RLock()
	defer m.iMutex.RUnlock()
	for _, index := range m.indexes {
		if index.Url == url && index.Hash == hash && index.Name == name {
			symbols = append(symbols, store.Symbol{
				File:       index.File,
				LineNumber: index.LineNumber,
				Line:       index.Line,
				LineBefore: index.LineBefore,
				LineAfter:  index.LineAfter,
				Kind:       index.Kind,
			})
		}
	}
	return
}

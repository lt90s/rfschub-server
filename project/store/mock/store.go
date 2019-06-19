package mock

import (
	"context"
	"github.com/lt90s/rfschub-server/project/store"
	"strconv"
	"sync"
)

type mockStore struct {
	pMutex   sync.RWMutex
	projects []project
	pid      int
}

type project struct {
	id      int
	uid     string
	url     string
	hash    string
	name    string
	branch  bool
	indexed bool
	members []int
}

func NewMockStore() *mockStore {
	return &mockStore{
		projects: make([]project, 0),
		pid:      10000,
	}
}

func (m *mockStore) NewProject(ctx context.Context, uid, url, hash, name string, branch bool) error {
	m.pMutex.Lock()
	defer m.pMutex.Unlock()

	exist := false
	for _, project := range m.projects {
		if project.uid == uid && project.url == url && project.hash == hash && project.name == name {
			exist = true
			break
		}
	}
	if exist {
		return store.ErrProjectExist
	}

	m.projects = append(m.projects, project{
		id:     m.pid,
		uid:    uid,
		url:    url,
		hash:   hash,
		name:   name,
		branch: branch,
	})
	m.pid += 1
	return nil
}

func (m *mockStore) GetProjectInfo(ctx context.Context, uid, url, name string) (info store.ProjectInfo, err error) {
	m.pMutex.RLock()
	defer m.pMutex.RUnlock()

	for _, project := range m.projects {
		if project.uid == uid && project.url == url && project.name == name {
			info.Hash = project.hash
			info.Id = strconv.Itoa(project.id)
			info.Branch = project.branch
			info.Indexed = project.indexed
			return
		}
	}
	err = store.ErrProjectNotExist
	return
}

func (m *mockStore) SetProjectIndexed(ctx context.Context, uid, url, hash string) error {
	m.pMutex.Lock()
	defer m.pMutex.Unlock()

	for idx, project := range m.projects {
		if project.uid == uid && project.url == url && project.hash == hash {
			m.projects[idx].indexed = true
			break
		}
	}
	return nil
}

package mockdb

import (
	"context"
	"github.com/lt90s/rfschub-server/gits/proto"
	"github.com/lt90s/rfschub-server/repository/store"
	"path"
)

type mockStore struct {
	details  []repositoryDetail
	repoInfo map[string]repositoryInfo
}

type repositoryInfo struct {
	commits []store.NamedCommit
}

type repositoryDetail struct {
	url    string
	commit string
	synced bool
	files  []repositoryFile
}

type repositoryFile struct {
	selfPath   string
	parentPath string
	dir        bool
	content    string
	synced     bool
	plain      bool
}

func NewMockStore() store.Store {
	return &mockStore{
		repoInfo: make(map[string]repositoryInfo),
		details:  make([]repositoryDetail, 0),
	}
}

func (m *mockStore) AddRepository(ctx context.Context, url string, commits []store.NamedCommit) error {
	if _, ok := m.repoInfo[url]; ok {
		return nil
	}

	m.repoInfo[url] = repositoryInfo{
		commits: commits,
	}
	return nil
}

func (m *mockStore) GetRepository(ctx context.Context, url string) ([]store.NamedCommit, error) {
	info, ok := m.repoInfo[url]
	if !ok {
		return nil, store.ErrorRepositoryNotFound
	}
	return info.commits, nil
}

func (m *mockStore) GetCommitByName(ctx context.Context, url string, name string) (string, error) {
	info, ok := m.repoInfo[url]
	if !ok {
		return "", store.ErrorRepositoryNotFound
	}

	for _, commit := range info.commits {
		if commit.Name == name {
			return commit.Hash, nil
		}
	}
	return "", store.ErrorRepositoryNotFound
}

func (m *mockStore) RepositoryExist(ctx context.Context, url string, hash string) (bool, error) {
	info, ok := m.repoInfo[url]
	if !ok {
		return false, nil
	}
	for _, commit := range info.commits {
		if commit.Hash == hash {
			return true, nil
		}
	}
	return false, nil
}

func (m *mockStore) GetDirectoryEntries(ctx context.Context, url, commit, path string) (synced bool, entries []store.DirectoryEntry, err error) {
	synced = false
	for _, detail := range m.details {
		if detail.url != url || detail.commit != commit {
			continue
		}
		synced = true
		for _, file := range detail.files {
			if file.parentPath == path {
				entries = append(entries, store.DirectoryEntry{
					File: file.selfPath,
					Dir:  file.dir,
				})
			}
		}
		break
	}
	return
}

func (m *mockStore) SetDirectories(ctx context.Context, url, commit string, entries []*gits.FileEntry) error {
	detail := repositoryDetail{
		url:    url,
		commit: commit,
		synced: true,
	}

	files := make([]repositoryFile, 0, len(entries))
	for _, entry := range entries {
		parentPath := path.Dir(entry.File)
		files = append(files, repositoryFile{
			selfPath:   entry.File,
			parentPath: parentPath,
			dir:        entry.Dir,
		})
	}

	detail.files = files
	m.details = append(m.details, detail)
	return nil
}

func (m *mockStore) SetBlob(ctx context.Context, url, commit, path, content string, plain bool) error {
	for i := range m.details {
		if m.details[i].url != url || m.details[i].commit != commit {
			continue
		}
		ok := false
		for j := range m.details[i].files {
			if m.details[i].files[j].selfPath == path {
				ok = true
				m.details[i].files[j].synced = true
				m.details[i].files[j].content = content
				m.details[i].files[j].plain = plain
			}
		}
		if ok {
			break
		}
	}
	return nil
}

func (m *mockStore) GetBlob(ctx context.Context, url, commit, path string) (blob store.Blob, err error) {
	for _, detail := range m.details {
		if detail.url != url || detail.commit != commit {
			continue
		}
		for _, file := range detail.files {
			if file.selfPath == path {
				blob.Synced = file.synced
				blob.Content = file.content
				blob.Plain = file.plain
				return
			}
		}
	}
	err = store.ErrorBlobNotFound
	return
}

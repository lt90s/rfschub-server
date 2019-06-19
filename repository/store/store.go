package store

import (
	"context"
	"errors"
	"github.com/lt90s/rfschub-server/gits/proto"
)

type Store interface {
	GetCommitByName(ctx context.Context, url string, name string) (string, error)
	AddRepository(ctx context.Context, url string, commits []NamedCommit) error
	GetRepository(ctx context.Context, url string) ([]NamedCommit, error)
	RepositoryExist(ctx context.Context, url string, hash string) (bool, error)
	SetDirectories(ctx context.Context, url, commit string, entries []*gits.FileEntry) error
	GetDirectoryEntries(ctx context.Context, url, name, path string) (bool, []DirectoryEntry, error)
	SetBlob(ctx context.Context, url, commit, path, content string, plain bool) error
	GetBlob(ctx context.Context, url, commit, path string) (Blob, error)
}

type DirectoryEntry struct {
	File string `bson:"file"`
	Dir  bool   `bson:"dir"`
}

type NamedCommit struct {
	Name   string `bson:"name"`
	Hash   string `bson:"hash"`
	Branch bool   `bson:"branch"`
}

type Blob struct {
	Content string `bson:"content"`
	Plain   bool   `bson:"plain"`
	Synced  bool   `bson:"synced"`
}

var (
	ErrorRepositoryNotFound = errors.New("repository not found")
	ErrorDirectoryNotFound  = errors.New("directory not found")
	ErrorBlobNotFound       = errors.New("blob not found")
)

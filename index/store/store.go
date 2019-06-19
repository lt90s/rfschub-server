package store

import (
	"context"
	"errors"
)

type Store interface {
	// create new index task. if task not expire, ErrIndexTaskExist should be returned
	NewIndexTask(ctx context.Context, url, hash string) error
	// set task state success or failure
	SetTaskState(ctx context.Context, url, hash string, success bool) error
	RepositoryIndexed(ctx context.Context, url, hash string) (bool, error)
	RepositoryFileIndexed(ctx context.Context, url, hash, file string) (bool, error)
	// add all index symbols of a file
	AddFileIndexEntries(ctx context.Context, entries []IndexEntry) error
	FindSymbols(ctx context.Context, url, hash, name string) (symbols []Symbol, err error)
}

var (
	ErrIndexTaskExist = errors.New("index task exists")
)

type IndexEntry struct {
	Url        string `json:"url" bson:"url"`
	Hash       string `json:"hash" bson:"hash"`
	File       string `json:"file" bson:"file"`
	Name       string `json:"name" bson:"name"`
	Pattern    string `json:"pattern" bson:"pattern"`
	Language   string `json:"language" bson:"language"`
	LineNumber int    `json:"lineNumber" bson:"lineNumber"`
	Line       string `json:"line" bson:"line"`
	LineBefore string `json:"lineBefore" bson:"lineBefore"`
	LineAfter  string `json:"lineAfter" bson:"lineAfter"`
	Kind       string `json:"kind" bson:"kind"`
	Scope      string `json:"scope" bson:"scope"`
	ScopeKind  string `json:"scopeKind" bson:"scopeKind"`
}

type Symbol struct {
	File       string `json:"file" bson:"file"`
	LineNumber int    `json:"lineNumber" bson:"lineNumber"`
	Line       string `json:"line" bson:"line"`
	LineBefore string `json:"lineBefore" bson:"lineBefore"`
	LineAfter  string `json:"lineAfter" bson:"lineAfter"`
	Kind       string `json:"kind" bson:"kind"`
}

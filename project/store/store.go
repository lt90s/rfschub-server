package store

import (
	"context"
	"errors"
)

type Store interface {
	NewProject(ctx context.Context, uid, url, hash, name string, branch bool) error
	GetProjectInfo(ctx context.Context, uid, url, name string) (ProjectInfo, error)
	SetProjectIndexed(ctx context.Context, uid, url, hash string) error
	ProjectExists(ctx context.Context, pid string) bool
	GetUserProjects(ctx context.Context, uid string) (projects []ProjectInfo, err error)
	AddAnnotation(ctx context.Context, pid, uid, file, annotation string, lineNumber int) error
	GetAnnotationLines(ctx context.Context, pid, file string) (lines []int32, err error)
	GetAnnotations(ctx context.Context, pid, file string, lineNumber int) (records []AnnotationRecord, err error)
	UpdateLatestAnnotation(ctx context.Context, pid, parent, sub, file, brief string, lineNumber int) error
	GetLatestAnnotations(ctx context.Context, pid, parent string) (annotations []LatestAnnotation, err error)
}

var (
	ErrProjectExist    = errors.New("project already created")
	ErrProjectNotExist = errors.New("project not exist")
)

type ProjectInfo struct {
	Id        string `bson:"-"`
	Url       string `bson:"url"`
	Name      string `bson:"name"`
	Hash      string `bson:"hash"`
	Branch    bool   `bson:"branch"`
	Indexed   bool   `bson:"indexed"`
	CreatedAt int64  `bson:"createdAt"`
}

type AnnotationRecord struct {
	Uid        string `bson:"uid"`
	Annotation string `bson:"annotation"`
	CreatedAt  int64  `bson:"createdAt"`
}

type LatestAnnotation struct {
	Sub        string `bson:"sub"`
	File       string `bson:"file"`
	LineNumber int    `bson:"lineNumber"`
	Brief      string `bson:"brief"`
	Timestamp  int64  `bson:"timestamp"`
}

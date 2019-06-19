// Code generated by protoc-gen-micro. DO NOT EDIT.
// source: repository.proto

package repository

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
	math "math"
)

import (
	context "context"
	client "github.com/micro/go-micro/client"
	server "github.com/micro/go-micro/server"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ client.Option
var _ server.Option

// Client API for RepositoryService service

type RepositoryService interface {
	IsRepositoryExist(ctx context.Context, in *RepositoryExistRequest, opts ...client.CallOption) (*RepositoryExistResponse, error)
	NamedCommits(ctx context.Context, in *NamedCommitsRequest, opts ...client.CallOption) (*NamedCommitsResponse, error)
	Directory(ctx context.Context, in *DirectoryRequest, opts ...client.CallOption) (*DirectoryResponse, error)
	Blob(ctx context.Context, in *BlobRequest, opts ...client.CallOption) (*BlobResponse, error)
}

type repositoryService struct {
	c    client.Client
	name string
}

func NewRepositoryService(name string, c client.Client) RepositoryService {
	if c == nil {
		c = client.NewClient()
	}
	if len(name) == 0 {
		name = "repository"
	}
	return &repositoryService{
		c:    c,
		name: name,
	}
}

func (c *repositoryService) IsRepositoryExist(ctx context.Context, in *RepositoryExistRequest, opts ...client.CallOption) (*RepositoryExistResponse, error) {
	req := c.c.NewRequest(c.name, "RepositoryService.IsRepositoryExist", in)
	out := new(RepositoryExistResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repositoryService) NamedCommits(ctx context.Context, in *NamedCommitsRequest, opts ...client.CallOption) (*NamedCommitsResponse, error) {
	req := c.c.NewRequest(c.name, "RepositoryService.NamedCommits", in)
	out := new(NamedCommitsResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repositoryService) Directory(ctx context.Context, in *DirectoryRequest, opts ...client.CallOption) (*DirectoryResponse, error) {
	req := c.c.NewRequest(c.name, "RepositoryService.Directory", in)
	out := new(DirectoryResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *repositoryService) Blob(ctx context.Context, in *BlobRequest, opts ...client.CallOption) (*BlobResponse, error) {
	req := c.c.NewRequest(c.name, "RepositoryService.Blob", in)
	out := new(BlobResponse)
	err := c.c.Call(ctx, req, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for RepositoryService service

type RepositoryServiceHandler interface {
	IsRepositoryExist(context.Context, *RepositoryExistRequest, *RepositoryExistResponse) error
	NamedCommits(context.Context, *NamedCommitsRequest, *NamedCommitsResponse) error
	Directory(context.Context, *DirectoryRequest, *DirectoryResponse) error
	Blob(context.Context, *BlobRequest, *BlobResponse) error
}

func RegisterRepositoryServiceHandler(s server.Server, hdlr RepositoryServiceHandler, opts ...server.HandlerOption) error {
	type repositoryService interface {
		IsRepositoryExist(ctx context.Context, in *RepositoryExistRequest, out *RepositoryExistResponse) error
		NamedCommits(ctx context.Context, in *NamedCommitsRequest, out *NamedCommitsResponse) error
		Directory(ctx context.Context, in *DirectoryRequest, out *DirectoryResponse) error
		Blob(ctx context.Context, in *BlobRequest, out *BlobResponse) error
	}
	type RepositoryService struct {
		repositoryService
	}
	h := &repositoryServiceHandler{hdlr}
	return s.Handle(s.NewHandler(&RepositoryService{h}, opts...))
}

type repositoryServiceHandler struct {
	RepositoryServiceHandler
}

func (h *repositoryServiceHandler) IsRepositoryExist(ctx context.Context, in *RepositoryExistRequest, out *RepositoryExistResponse) error {
	return h.RepositoryServiceHandler.IsRepositoryExist(ctx, in, out)
}

func (h *repositoryServiceHandler) NamedCommits(ctx context.Context, in *NamedCommitsRequest, out *NamedCommitsResponse) error {
	return h.RepositoryServiceHandler.NamedCommits(ctx, in, out)
}

func (h *repositoryServiceHandler) Directory(ctx context.Context, in *DirectoryRequest, out *DirectoryResponse) error {
	return h.RepositoryServiceHandler.Directory(ctx, in, out)
}

func (h *repositoryServiceHandler) Blob(ctx context.Context, in *BlobRequest, out *BlobResponse) error {
	return h.RepositoryServiceHandler.Blob(ctx, in, out)
}

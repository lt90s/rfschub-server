package service

import (
	"context"
	accountClient "github.com/lt90s/rfschub-server/account/client"
	"github.com/lt90s/rfschub-server/account/proto"
	"github.com/lt90s/rfschub-server/common/errors"
	"github.com/lt90s/rfschub-server/common/url"
	indexClient "github.com/lt90s/rfschub-server/index/client"
	"github.com/lt90s/rfschub-server/index/proto"
	"github.com/lt90s/rfschub-server/project/config"
	proto "github.com/lt90s/rfschub-server/project/proto"
	"github.com/lt90s/rfschub-server/project/store"
	log "github.com/sirupsen/logrus"
	"path"
	"strings"
)

type projectService struct {
	store         store.Store
	indexClient   index.IndexService
	accountClient account.AccountService
}

func NewProjectService(store store.Store) proto.ProjectHandler {
	conf := config.DefaultConfig
	return &projectService{
		store:         store,
		indexClient:   indexClient.New(indexClient.ServerConfig{ServiceName: conf.Index}),
		accountClient: accountClient.New(accountClient.ServerConfig{ServiceName: conf.Account}),
	}
}

// TODO: like project & join project

// create a new annotation project
func (service *projectService) NewProject(ctx context.Context, req *proto.NewProjectRequest, rsp *proto.NewProjectResponse) error {
	repoUrl, ok := url.NormalizeRepoUrl(req.Url)
	if !ok {
		return errors.NewBadRequestError(-1, "repository url invalid")
	}
	log.Debugf("[NewProject]: data=%v", req)
	// TODO: check if #project exceed quota
	err := service.store.NewProject(ctx, req.Uid, repoUrl, req.Hash, req.Name, req.Branch)
	if err != nil {
		if err == store.ErrProjectExist {
			return errors.NewBadRequestError(int(proto.ErrorCode_ProjectExist), "project exists")
		} else {
			return errors.NewInternalError(-1, err.Error())
		}
	}
	service.requestForIndexing(ctx, req.Url, req.Url, req.Hash)
	return nil
}

func (service *projectService) requestForIndexing(ctx context.Context, uid, url, hash string) {
	iRsp, err := service.indexClient.IndexRepository(ctx, &index.IndexRepositoryRequest{
		Url:  url,
		Hash: hash,
	})
	log.Debugf("[requestForIndexing]: uid=%s url=%s hash=%s err=%v index=%v", uid, url, hash, err, iRsp.Indexed)
	if err == nil && iRsp.Indexed {
		err = service.store.SetProjectIndexed(ctx, uid, url, hash)
		if err != nil {
			log.Warnf("[requestForIndexing] set project indexed error: uid=%s url=%s hash=%s error=%v", uid, url, hash, err)
		}
	}
}

func (service *projectService) ProjectInfo(ctx context.Context, req *proto.ProjectInfoRequest, rsp *proto.ProjectInfoResponse) error {
	log.Debugf("[ProjectInfo]: uid=%s ownerUid=%s url=%s name=%s", req.Uid, req.OwnerUid, req.Url, req.Name)

	repoUrl, ok := url.NormalizeRepoUrl(req.Url)
	if !ok {
		return errors.NewBadRequestError(-1, "repository url invalid")
	}

	info, err := service.store.GetProjectInfo(ctx, req.OwnerUid, repoUrl, req.Name)
	if err != nil {
		if err == store.ErrProjectNotExist {
			return errors.NewNotFoundError(-1, err.Error())
		}
		return err
	}

	// side effect: if repository is not indexed, request for indexing
	if !info.Indexed {
		log.Debugf("[projectInfo] project not indexed, request for indexing: uid=%s ownerUid=%s url=%s name=%s", req.Uid, req.OwnerUid, req.Url, req.Name)
		service.requestForIndexing(ctx, req.OwnerUid, repoUrl, info.Hash)
	}
	// TODO: membership
	if req.Uid == req.OwnerUid {
		rsp.CanAnnotate = true
	}
	rsp.Id = info.Id
	rsp.Hash = info.Hash
	rsp.Branch = info.Branch
	rsp.Indexed = info.Indexed
	return nil
}

func (service *projectService) ListProjects(ctx context.Context, req *proto.ListProjectsRequest, rsp *proto.ListProjectsResponse) error {
	infos, err := service.store.GetUserProjects(ctx, req.Uid)
	if err != nil {
		return err
	}

	rsp.Projects = make([]*proto.ProjectInfo, 0, len(infos))
	for _, info := range infos {
		rsp.Projects = append(rsp.Projects, &proto.ProjectInfo{
			Url:       info.Url,
			Hash:      info.Hash,
			Name:      info.Name,
			Branch:    info.Branch,
			CreatedAt: info.CreatedAt,
		})
	}
	return nil
}

func (service *projectService) AddAnnotation(ctx context.Context, req *proto.AddAnnotationRequest, rsp *proto.AddAnnotationResponse) error {
	if !service.store.ProjectExists(ctx, req.Pid) {
		return errors.NewNotFoundError(-1, "project not exists")
	}
	// TODO: check if req.Uid has right to add annotation to this project req.Pid
	// TODO: check if req.Url and req.File are valid

	err := service.store.AddAnnotation(ctx, req.Pid, req.Uid, req.File, req.Annotation, int(req.LineNumber))
	if err != nil {
		log.Warnf("[AddAnnotation] add annotation error: %v", err)
		return err
	}

	currentFile := req.File
	for currentFile != "." && currentFile != "/" {
		parent := path.Dir(currentFile)
		sub := path.Base(currentFile)
		var brief string
		if len(req.Annotation) > 64 {
			brief = req.Annotation[:64]
		} else {
			brief = req.Annotation
		}
		log.Debugf("[AddAnnotation] update latest annotation: pid=%s parent=%s file=%s", req.Pid, parent, currentFile)
		err = service.store.UpdateLatestAnnotation(ctx, req.Pid, parent, sub, req.File, brief, int(req.LineNumber))
		if err != nil {
			log.Warnf("[AddAnnotation] update latest annotation error: %v", err)
			break
		}
		currentFile = path.Dir(currentFile)
	}
	return nil
}

func (service *projectService) GetAnnotationLines(ctx context.Context, req *proto.GetAnnotationLinesRequest, rsp *proto.GetAnnotationLinesResponse) error {
	lines, err := service.store.GetAnnotationLines(ctx, req.Pid, req.File)
	if err != nil {
		log.Warnf("[GetFileAnnotationLines] get annotation lines error: pid=%s file=%s error=%v", req.Pid, req.File, err)
	}

	rsp.Lines = lines
	return nil
}

// TODO: pagination support
func (service *projectService) GetAnnotations(ctx context.Context, req *proto.GetAnnotationsRequest, rsp *proto.GetAnnotationsResponse) error {
	req.File = strings.TrimPrefix(req.File, "/")

	log.Debugf("[GetAnnotations]: pid=%s file=%s line=%d", req.Pid, req.File, req.LineNumber)
	records, err := service.store.GetAnnotations(ctx, req.Pid, req.File, int(req.LineNumber))
	if err != nil {
		log.Warnf("[GetAnnotations] store get annotations error: %v", err)
		return err
	}
	log.Debugf("[GetAnnotations]: records=%v", records)

	rsp.Records = make([]*proto.AnnotationRecord, len(records))

	uidsMap := make(map[string]struct{})
	for i := range records {
		rsp.Records[i] = &proto.AnnotationRecord{
			Uid:        records[i].Uid,
			Annotation: records[i].Annotation,
			CreatedAt:  records[i].CreatedAt,
		}
		uidsMap[records[i].Uid] = struct{}{}
	}
	uids := make([]string, 0, len(uidsMap))
	for uid := range uidsMap {
		uids = append(uids, uid)
	}

	// TODO: it's a mess
	infoRsp, err := service.accountClient.AccountsBasicInfo(ctx, &account.AccountsBasicInfoRequest{Uids: uids})
	if err != nil {
		log.Warnf("[GetAnnotations] get account basic info error: %v", err)
		return err
	}
	uidInfoMap := make(map[string]*account.BasicInfo, len(infoRsp.Infos))
	for _, info := range infoRsp.Infos {
		uidInfoMap[info.Id] = info
	}

	for i := range records {
		uid := rsp.Records[i].Uid
		info, ok := uidInfoMap[uid]
		if !ok {
			continue
		}
		rsp.Records[i].Name = info.Name
	}
	return nil
}

func (service *projectService) GetLatestAnnotations(ctx context.Context, req *proto.GetLatestAnnotationsRequest, rsp *proto.GetLatestAnnotationsResponse) error {
	log.Debugf("[GetLatestAnnotations]: pid=%s parent=%s", req.Pid, req.Parent)
	annotations, err := service.store.GetLatestAnnotations(ctx, req.Pid, req.Parent)
	if err != nil {
		return err
	}

	rsp.Annotations = make([]*proto.LatestAnnotation, 0, len(annotations))
	for _, annotation := range annotations {
		rsp.Annotations = append(rsp.Annotations, &proto.LatestAnnotation{
			Sub:        annotation.Sub,
			File:       annotation.File,
			Brief:      annotation.Brief,
			LineNumber: int32(annotation.LineNumber),
			Timestamp:  annotation.Timestamp,
		})
	}
	return nil
}

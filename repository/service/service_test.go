package service

import (
	"context"
	"github.com/lt90s/rfschub-server/repository/config"
	proto "github.com/lt90s/rfschub-server/repository/proto"
	"github.com/lt90s/rfschub-server/repository/store/mockdb"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

// these tests depend on the `repoUrl` repository cloned by the GitService
// TODO: test independently

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestRepositoryService_Repository_Directory_Blob(t *testing.T) {
	store := mockdb.NewMockStore()
	rs := NewRepositoryService(config.DefaultConfig, store)

	ctx := context.Background()

	var rsp proto.RepositoryResponse
	for i := 0; i < 10; i++ {
		err := rs.Repository(ctx, &proto.RepositoryRequest{Url: repoUrl}, &rsp)
		require.NoError(t, err)
		if rsp.Code == proto.RepositoryErrorCode_Success {
			break
		}
		require.Equal(t, proto.RepositoryErrorCode_InSync, rsp.Code)
		time.Sleep(100 * time.Millisecond)
	}
	t.Log(rsp.Commits)

	var dRsp proto.DirectoryResponse
	for i := 0; i < 10; i++ {
		err := rs.Directory(ctx, &proto.DirectoryRequest{Url: repoUrl, Commit: commit, Path: "api"}, &dRsp)
		require.NoError(t, err)
		if dRsp.Code == proto.RepositoryErrorCode_Success {
			break
		}
		require.Equal(t, proto.RepositoryErrorCode_InSync, dRsp.Code)
		time.Sleep(100 * time.Millisecond)
	}
	t.Log(dRsp.Entries)

	var bRsp proto.BlobResponse
	for i := 0; i < 10; i++ {
		err := rs.Blob(ctx, &proto.BlobRequest{Url: repoUrl, Commit: commit, Path: ".gitignore"}, &bRsp)
		require.NoError(t, err)
		if bRsp.Code == proto.RepositoryErrorCode_Success {
			break
		}
		require.Equal(t, proto.RepositoryErrorCode_InSync, bRsp.Code)
		time.Sleep(100 * time.Millisecond)
	}
	require.True(t, bRsp.Plain)
	t.Log(bRsp.Content)
}

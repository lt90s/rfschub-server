package service

import (
	"context"
	"github.com/lt90s/rfschub-server/repository/config"
	"github.com/lt90s/rfschub-server/repository/store/mockdb"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

const (
	repoUrl = "https://github.com/lt90s/goanalytics"
	commit  = "0da408de63c77b1766c5cde56478d32fdc75ad1e"
)

func TestSyncRepository(t *testing.T) {
	store := mockdb.NewMockStore()
	syncer := newSyncer(config.DefaultConfig, store)
	ctx := context.Background()

	err := syncer.syncRepository(ctx, repoUrl)
	require.NoError(t, err)

	err = syncer.syncRepository(ctx, repoUrl)
	require.Equal(t, ErrInSync, err)

	syncer.wait(10 * time.Second)

	commits, err := store.GetRepository(ctx, repoUrl)
	require.NoError(t, err)
	t.Log(commits)
}

func TestSyncDirectories_SyncBlob(t *testing.T) {
	store := mockdb.NewMockStore()

	syncer := newSyncer(config.DefaultConfig, store)

	err := syncer.syncDirectories(context.Background(), repoUrl, commit)
	require.NoError(t, err)

	syncer.wait(10 * time.Second)

	synced, entries, err := store.GetDirectoryEntries(context.Background(), repoUrl, commit, "api")
	require.NoError(t, err)
	require.True(t, synced)
	require.Len(t, entries, 3)

	dirs := []string{"api/authentication", "api/middlewares", "api/router"}
	for _, dir := range dirs {
		found := false
		for _, entry := range entries {
			if entry.File == dir {
				found = true
				require.True(t, entry.Dir)
			}
		}
		require.True(t, found)
	}
	t.Log(entries)

	err = syncer.syncBlob(context.Background(), repoUrl, commit, ".gitignore")
	require.NoError(t, err)

	syncer.wait(10 * time.Second)
	blob, err := store.GetBlob(context.Background(), repoUrl, commit, ".gitignore")
	require.NoError(t, err)
	require.True(t, blob.Synced)
	require.True(t, blob.Plain)
	require.Equal(t, ".idea\n*.exe", blob.Content)
}

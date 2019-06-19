package test

import (
	"context"
	"github.com/lt90s/rfschub-server/project/store"
	"github.com/lt90s/rfschub-server/project/store/mock"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

const (
	uid    = "uid"
	url    = "url"
	hash   = "hash"
	name   = "name"
	branch = true
)

var (
	testedStore store.Store
)

func init() {
	target := os.Getenv("ANNOTATION_STORE_TARGET")
	switch target {
	default:
		testedStore = mock.NewMockStore()
	}
}

func TestNewProject(t *testing.T) {
	ctx := context.Background()
	err := testedStore.NewProject(ctx, uid, url, hash, name, branch)
	require.NoError(t, err)

	err = testedStore.NewProject(ctx, uid, url, hash, name, branch)
	require.Equal(t, store.ErrProjectExist, err)
}

package service

import (
	"context"
	"github.com/lt90s/rfschub-server/gits/config"
	proto "github.com/lt90s/rfschub-server/gits/proto"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"os"
	"strconv"
	"testing"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestCommand_newGitCommander(t *testing.T) {
	conf := config.CommandConf{
		Path: "/path/not/exist",
	}
	require.Panics(t, func() { _ = newGitCommander(conf) })
	conf.Path = "/usr/local/bin/git"
	require.NotPanics(t, func() { _ = newGitCommander(conf) })
}

func TestProgressWriter(t *testing.T) {
	progresses := make([]string, 0)

	updater := func(progress string) {
		progresses = append(progresses, progress)
	}
	pw := &progressWriter{updater: updater}

	for i := 0; i < 40; i++ {
		_, _ = pw.Write([]byte(strconv.Itoa(i)))
	}

	require.Len(t, progresses, 2)
	require.Equal(t, "19", progresses[0])
	require.Equal(t, "39", progresses[1])
}

var testConf = config.CommandConf{
	Path: "/usr/local/bin/git",
	Data: "/tmp/git",
	Concurrency: config.CommandConcurrency{
		Clone: 1,
		Other: 1,
	},
	CloneTimeout:   600,
	DefaultTimeout: 60,
}

func TestCommand_Clone_CloneStatus_NameCommits(t *testing.T) {
	commander := newGitCommander(testConf)
	url := "https://github.com/lt90s/goanalytics"
	dir, err := commander.urlToLocal(url)
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	ctx := context.Background()
	status, _ := commander.cloneStatus(ctx, url)
	require.Equal(t, proto.CloneStatus_Unknown, status)

	err = commander.clone(ctx, url)
	require.NoError(t, err)

	status, _ = commander.cloneStatus(ctx, url)
	require.Equal(t, proto.CloneStatus_Cloning, status)

	err = commander.clone(context.Background(), url)
	require.Equal(t, errorRepositoryCloning, err, commander.status)

	err = commander.clone(context.Background(), "https://github.com/lt90s/goanalytics-web")
	require.Equal(t, errorGitBusy, err)

	commander.wait()
	err = commander.clone(context.Background(), url)
	require.Equal(t, errorRepositoryCloned, err)

	status, _ = commander.cloneStatus(ctx, url)
	require.Equal(t, proto.CloneStatus_Cloned, status)

	commits, err := commander.getNamedCommits(ctx, url)
	require.NoError(t, err)
	require.True(t, len(commits) >= 1)
	t.Log(commits)

	entries, err := commander.getRepositoryFiles(ctx, url, commits[0].Hash)
	require.NoError(t, err)
	t.Log(entries)

	plain, content, err := commander.getRepositoryBlob(ctx, url, commits[0].Hash, "README.md")
	require.NoError(t, err)
	require.True(t, plain)
	t.Log(string(content))

	plain, _, err = commander.getRepositoryBlob(ctx, url, commits[0].Hash, ".github/image/summary.png")
	require.NoError(t, err)
	require.False(t, plain)
}

package url

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNormalizeRepoUrl(t *testing.T) {
	url := "https://github.com/lt90s/goanalytics"
	_, ok := NormalizeRepoUrl(url)
	require.True(t, ok)
}

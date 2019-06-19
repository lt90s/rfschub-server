package service

import (
	"context"
	"github.com/lt90s/rfschub-server/index/config"
	"github.com/lt90s/rfschub-server/index/store/mock"
	"github.com/sirupsen/logrus"
	"testing"
)

const (
	url  = "https://github.com/lt90s/goanalytics"
	hash = "0da408de63c77b1766c5cde56478d32fdc75ad1e"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestIndexer_indexRepository(t *testing.T) {
	indexer := newIndexer(config.DefaultConfig, nil, nil, mock.NewMockStore())
	err := indexer.indexRepository(context.Background(), indexRequest{url, hash}, 0)
	t.Log(err)

}

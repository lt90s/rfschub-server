package service

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestNewSyntectService(t *testing.T) {
	s := NewSyntectService()
	s.Stop()
}

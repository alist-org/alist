package aria2

import (
	"context"
	"time"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/aria2/rpc"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var DownTaskManager = task.NewTaskManager[string](3)
var notify = NewNotify()
var client rpc.Client

func InitClient(timeout int) (string, error) {
	client = nil
	uri := setting.GetStr(conf.Aria2Uri)
	secret := setting.GetStr(conf.Aria2Secret)
	return InitAria2Client(uri, secret, timeout)
}

func InitAria2Client(uri string, secret string, timeout int) (string, error) {
	c, err := rpc.New(context.Background(), uri, secret, time.Duration(timeout)*time.Second, notify)
	if err != nil {
		return "", errors.Wrap(err, "failed to init aria2 client")
	}
	version, err := c.GetVersion()
	if err != nil {
		return "", errors.Wrapf(err, "failed get aria2 version")
	}
	client = c
	log.Infof("using aria2 version: %s", version.Version)
	return version.Version, nil
}

func IsAria2Ready() bool {
	return client != nil
}

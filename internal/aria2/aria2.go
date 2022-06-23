package aria2

import (
	"context"
	"github.com/alist-org/alist/v3/pkg/aria2/rpc"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"time"
)

var DownTaskManager = task.NewTaskManager[string](3)
var notify = NewNotify()
var client rpc.Client

func InitAria2Client(uri string, secret string, timeout int) error {
	c, err := rpc.New(context.Background(), uri, secret, time.Duration(timeout)*time.Second, notify)
	if err != nil {
		return errors.Wrap(err, "failed to init aria2 client")
	}
	client = c
	version, err := client.GetVersion()
	if err != nil {
		return errors.Wrapf(err, "failed get aria2 version")
	}
	log.Infof("using aria2 version: %s", version.Version)
	return nil
}

func IsAria2Ready() bool {
	return client != nil
}

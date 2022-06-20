package aria2

import (
	"context"
	"github.com/alist-org/alist/v3/pkg/aria2/rpc"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/pkg/errors"
	"time"
)

var Aria2TaskManager = task.NewTaskManager()
var client rpc.Client

func InitAria2Client(uri string, secret string, timeout int) error {
	c, err := rpc.New(context.Background(), uri, secret, time.Duration(timeout)*time.Second, &Notify{})
	if err != nil {
		return errors.Wrap(err, "failed to init aria2 client")
	}
	client = c
	return nil
}

func IsAria2Ready() bool {
	return client != nil
}

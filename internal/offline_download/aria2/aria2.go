package aria2

import (
	"context"
	"fmt"
	"time"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/offline_download"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/aria2/rpc"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var notify = NewNotify()

type Aria2 struct {
	client rpc.Client
}

func (a *Aria2) Items() []model.SettingItem {
	// aria2 settings
	return []model.SettingItem{
		{Key: conf.Aria2Uri, Value: "http://localhost:6800/jsonrpc", Type: conf.TypeString, Group: model.OFFLINE_DOWNLOAD, Flag: model.PRIVATE},
		{Key: conf.Aria2Secret, Value: "", Type: conf.TypeString, Group: model.OFFLINE_DOWNLOAD, Flag: model.PRIVATE},
	}
}

func (a *Aria2) Init() (string, error) {
	a.client = nil
	uri := setting.GetStr(conf.Aria2Uri)
	secret := setting.GetStr(conf.Aria2Secret)
	c, err := rpc.New(context.Background(), uri, secret, 4*time.Second, notify)
	if err != nil {
		return "", errors.Wrap(err, "failed to init aria2 client")
	}
	version, err := c.GetVersion()
	if err != nil {
		return "", errors.Wrapf(err, "failed get aria2 version")
	}
	a.client = c
	log.Infof("using aria2 version: %s", version.Version)
	return fmt.Sprintf("aria2 version: %s", version.Version), nil
}

func (a *Aria2) IsReady() bool {
	//TODO implement me
	panic("implement me")
}

func (a *Aria2) AddURI(args *offline_download.AddUriArgs) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (a *Aria2) Remove(tid string) error {
	//TODO implement me
	panic("implement me")
}

func (a *Aria2) Status(tid string) (*offline_download.Status, error) {
	//TODO implement me
	panic("implement me")
}

func (a *Aria2) GetFile(tid string) *offline_download.File {
	//TODO implement me
	panic("implement me")
}

var _ offline_download.Tool = (*Aria2)(nil)

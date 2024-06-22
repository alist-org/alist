package pikpak

import (
	"fmt"
	// "github.com/alist-org/alist/v3/internal/model"
	"context"

	"github.com/alist-org/alist/v3/drivers/pikpak"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/offline_download/tool"
	"github.com/alist-org/alist/v3/internal/op"
	// "github.com/alist-org/alist/v3/pkg/utils"
	// "net/http"
	// "net/url"
	// "os"
	// "path"
	// "path/filepath"
)

type PikPak struct {
}

func (p PikPak) Name() string {
	return "pikpak"
}

func (p PikPak) Items() []model.SettingItem {
	return nil
}

func (p PikPak) Run(task *tool.DownloadTask) error {
	return errs.NotSupport
}

func (p PikPak) Init() (string, error) {
	return "ok", nil
}

func (p PikPak) IsReady() bool {
	return true
}

func (p PikPak) AddURL(args *tool.AddUrlArgs) (string, error) {
	// args.TempDir 已经被修改为了 DstDirPath
	storage, actualPath, err := op.GetStorageAndActualPath(args.TempDir)
	if err != nil {
		return "", err
	}
	pikpakDriver, ok := storage.(*pikpak.PikPak)
	if !ok {
		return "", fmt.Errorf("unsupported storage driver for offline download, only Pikpak is supported")
	}

	ctx := context.Background()
	parentDir, err := op.Get(ctx, storage, actualPath)
	if err != nil {
		return "", err
	}

	t, err := pikpakDriver.OfflineDownload(ctx, args.Url, parentDir, "")
	if err != nil {
		return "", fmt.Errorf("failed to add offline download task: %w", err)
	}

	return t.ID, nil
}

func (p PikPak) Remove(task *tool.DownloadTask) error {
	panic("should not be called")
}

func (p PikPak) Status(task *tool.DownloadTask) (*tool.Status, error) {
	s := &tool.Status{}
	s.Completed = true
	s.Progress = 100
	s.Err = nil
	return s, nil
}

func init() {
	tool.Tools.Add(&PikPak{})
}

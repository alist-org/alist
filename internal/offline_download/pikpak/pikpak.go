package pikpak

import (
	"context"
	"fmt"

	"github.com/alist-org/alist/v3/drivers/pikpak"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/offline_download/tool"
	"github.com/alist-org/alist/v3/internal/op"
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
	storage, _, err := op.GetStorageAndActualPath(task.DstDirPath)
	if err != nil {
		return err
	}
	pikpakDriver, ok := storage.(*pikpak.PikPak)
	if !ok {
		return fmt.Errorf("unsupported storage driver for offline download, only Pikpak is supported")
	}
	ctx := context.Background()
	err = pikpakDriver.DeleteOfflineTasks(ctx, []string{task.GID}, false)
	if err != nil {
		return err
	}
	return nil
}

func (p PikPak) Status(task *tool.DownloadTask) (*tool.Status, error) {
	storage, _, err := op.GetStorageAndActualPath(task.DstDirPath)
	if err != nil {
		return nil, err
	}
	pikpakDriver, ok := storage.(*pikpak.PikPak)
	if !ok {
		return nil, fmt.Errorf("unsupported storage driver for offline download, only Pikpak is supported")
	}
	tasks, err := GetTasks(pikpakDriver)
	if err != nil {
		return nil, err
	}
	s := &tool.Status{
		Progress:  0,
		NewGID:    "",
		Completed: false,
		Status:    "",
		Err:       nil,
	}
	for _, t := range tasks {
		if t.ID == task.GID {
			s.Progress = float64(t.Progress)
			s.Status = t.Message
			s.Completed = (t.Phase == "PHASE_TYPE_COMPLETE")
			if t.Phase == "PHASE_TYPE_ERROR" {
				s.Err = fmt.Errorf(t.Message)
			}
		} else {
			s.Progress = 0
			s.Status = "the task has been deleted"
			s.Completed = false
			s.Err = fmt.Errorf("the task has been deleted")
		}
	}
	return s, nil
}

func init() {
	tool.Tools.Add(&PikPak{})
}

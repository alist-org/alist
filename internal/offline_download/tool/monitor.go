package tool

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/stream"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Monitor struct {
	tool       Tool
	tsk        *task.Task[string]
	tempDir    string
	retried    int
	dstDirPath string
	finish     chan struct{}
	signal     chan int
}

func (m *Monitor) Loop() error {
	m.finish = make(chan struct{})
	var (
		err error
		ok  bool
	)
outer:
	for {
		select {
		case <-m.tsk.Ctx.Done():
			err := m.tool.Remove(m.tsk.ID)
			return err
		case <-m.signal:
			ok, err = m.Update()
			if ok {
				break outer
			}
		case <-time.After(time.Second * 2):
			ok, err = m.Update()
			if ok {
				break outer
			}
		}
	}
	if err != nil {
		return err
	}
	m.tsk.SetStatus("aria2 download completed, transferring")
	<-m.finish
	m.tsk.SetStatus("completed")
	return nil
}

// Update download status, return true if download completed
func (m *Monitor) Update() (bool, error) {
	info, err := m.tool.Status(m.tsk.ID)
	if err != nil {
		m.retried++
		log.Errorf("failed to get status of %s, retried %d times", m.tsk.ID, m.retried)
		return false, nil
	}
	if m.retried > 5 {
		return true, errors.Errorf("failed to get status of %s, retried %d times", m.tsk.ID, m.retried)
	}
	m.retried = 0
	m.tsk.SetProgress(info.Progress)
	m.tsk.SetStatus("tool: " + info.Status)
	if info.NewTID != "" {
		log.Debugf("followen by: %+v", info.NewTID)
		DownTaskManager.RawTasks().Delete(m.tsk.ID)
		m.tsk.ID = info.NewTID
		DownTaskManager.RawTasks().Store(m.tsk.ID, m.tsk)
		return false, nil
	}
	// if download completed
	if info.Completed {
		err := m.Complete()
		return true, errors.WithMessage(err, "failed to transfer file")
	}
	// if download failed
	if info.Err != nil {
		return true, errors.Errorf("failed to download %s, error: %s", m.tsk.ID, info.Err.Error())
	}
	return false, nil
}

var TransferTaskManager = task.NewTaskManager(3, func(k *uint64) {
	atomic.AddUint64(k, 1)
})

func (m *Monitor) Complete() error {
	// check dstDir again
	storage, dstDirActualPath, err := op.GetStorageAndActualPath(m.dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	var files []File
	if f := m.tool.GetFiles(m.tsk.ID); f != nil {
		files = f
	} else {
		files, err = GetFiles(m.tempDir)
		if err != nil {
			return errors.Wrapf(err, "failed to get files")
		}
	}
	// upload files
	var wg sync.WaitGroup
	wg.Add(len(files))
	go func() {
		wg.Wait()
		err := os.RemoveAll(m.tempDir)
		m.finish <- struct{}{}
		if err != nil {
			log.Errorf("failed to remove aria2 temp dir: %+v", err.Error())
		}
	}()
	for i, _ := range files {
		file := files[i]
		TransferTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
			Name: fmt.Sprintf("transfer %s to [%s](%s)", file.Path, storage.GetStorage().MountPath, dstDirActualPath),
			Func: func(tsk *task.Task[uint64]) error {
				defer wg.Done()
				mimetype := utils.GetMimeType(file.Path)
				rc, err := file.GetReadCloser()
				if err != nil {
					return errors.Wrapf(err, "failed to open file %s", file.Path)
				}
				s := &stream.FileStream{
					Ctx: nil,
					Obj: &model.Object{
						Name:     filepath.Base(file.Path),
						Size:     file.Size,
						Modified: file.Modified,
						IsFolder: false,
					},
					Reader:   rc,
					Mimetype: mimetype,
					Closers:  utils.NewClosers(rc),
				}
				relDir, err := filepath.Rel(m.tempDir, filepath.Dir(file.Path))
				if err != nil {
					log.Errorf("find relation directory error: %v", err)
				}
				newDistDir := filepath.Join(dstDirActualPath, relDir)
				return op.Put(tsk.Ctx, storage, newDistDir, s, tsk.SetProgress)
			},
		}))
	}
	return nil
}

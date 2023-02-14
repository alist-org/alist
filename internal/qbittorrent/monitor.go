package qbittorrent

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

type Monitor struct {
	tsk        *task.Task[string]
	tempDir    string
	dstDirPath string
	finish     chan struct{}
}

func (m *Monitor) Loop() error {
	var (
		err       error
		completed bool
	)
	m.finish = make(chan struct{})
outer:
	for {
		select {
		case <-m.tsk.Ctx.Done():
			err = qbclient.Delete(m.tsk.ID)
			return err
		case <-time.After(time.Second * 2):
			completed, err = m.update()
			if completed {
				break outer
			}
		}
	}
	if err != nil {
		return err
	}
	m.tsk.SetStatus("qbittorrent download completed, transferring")
	<-m.finish
	m.tsk.SetStatus("completed")
	return nil
}

func (m *Monitor) update() (bool, error) {
	info, err := qbclient.GetInfo(m.tsk.ID)
	if err != nil {
		m.tsk.SetStatus("qbittorrent " + string(info.State))
		return true, err
	}

	progress := float64(info.Completed) / float64(info.Size) * 100
	m.tsk.SetProgress(int(progress))
	switch info.State {
	case UPLOADING:
	case PAUSEDUP:
	case QUEUEDUP:
	case STALLEDUP:
	case FORCEDUP:
	case CHECKINGUP:
		err = m.complete()
		return true, errors.WithMessage(err, "failed to transfer file")
	case ALLOCATING:
	case DOWNLOADING:
	case METADL:
	case PAUSEDDL:
	case QUEUEDDL:
	case STALLEDDL:
	case CHECKINGDL:
	case FORCEDDL:
	case CHECKINGRESUMEDATA:
	case MOVING:
	case UNKNOWN: // or maybe should return an error for UNKNOWN?
		m.tsk.SetStatus("qbittorrent downloading")
		return false, nil
	case ERROR:
	case MISSINGFILES:
		return true, errors.Errorf("failed to download %s, error: %s", m.tsk.ID, info.State)
	}
	return true, errors.New("unknown error occurred downloading qbittorrent") // should never happen
}

var TransferTaskManager = task.NewTaskManager(3, func(k *uint64) {
	atomic.AddUint64(k, 1)
})

func (m *Monitor) complete() error {
	// check dstDir again
	storage, dstDirActualPath, err := op.GetStorageAndActualPath(m.dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	// get files
	files, err := qbclient.GetFiles(m.tsk.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to get files of %s", m.tsk.ID)
	}
	log.Debugf("files len: %d", len(files))
	// upload files
	var wg sync.WaitGroup
	wg.Add(len(files))
	go func() {
		wg.Wait()
		err := os.RemoveAll(m.tempDir)
		m.finish <- struct{}{}
		if err != nil {
			log.Errorf("failed to remove qbittorrent temp dir: %+v", err.Error())
		}
	}()
	for _, file := range files {
		filePath := filepath.Join(m.tempDir, file.Name)
		TransferTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
			Name: fmt.Sprintf("transfer %s to [%s](%s)", filePath, storage.GetStorage().MountPath, dstDirActualPath),
			Func: func(tsk *task.Task[uint64]) error {
				defer wg.Done()
				size := file.Size
				mimetype := utils.GetMimeType(filePath)
				f, err := os.Open(filePath)
				if err != nil {
					return errors.Wrapf(err, "failed to open file %s", filePath)
				}
				stream := &model.FileStream{
					Obj: &model.Object{
						Name:     path.Base(filePath),
						Size:     size,
						Modified: time.Now(),
						IsFolder: false,
					},
					ReadCloser: f,
					Mimetype:   mimetype,
				}
				newDistDir := filepath.Join(dstDirActualPath, file.Name)
				return op.Put(tsk.Ctx, storage, newDistDir, stream, tsk.SetProgress)
			},
		}))
	}
	return nil
}

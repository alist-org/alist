package qbittorrent

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/stream"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Monitor struct {
	tsk        *task.Task[string]
	tempDir    string
	dstDirPath string
	seedtime   int
	finish     chan struct{}
}

func (m *Monitor) Loop() error {
	var (
		err       error
		completed bool
	)
	m.finish = make(chan struct{})

	// wait for qbittorrent to parse torrent and create task
	m.tsk.SetStatus("waiting for qbittorrent to parse torrent and create task")
	waitCount := 0
	for {
		_, err := qbclient.GetInfo(m.tsk.ID)
		if err == nil {
			break
		}
		switch err.(type) {
		case InfoNotFoundError:
			break
		default:
			return err
		}

		waitCount += 1
		if waitCount >= 60 {
			return errors.New("torrent parse timeout")
		}
		timer := time.NewTimer(time.Second)
		<-timer.C
	}

outer:
	for {
		select {
		case <-m.tsk.Ctx.Done():
			// delete qbittorrent task and downloaded files when the task exits with error
			return qbclient.Delete(m.tsk.ID, true)
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
	case UPLOADING, PAUSEDUP, QUEUEDUP, STALLEDUP, FORCEDUP, CHECKINGUP:
		err = m.complete()
		return true, errors.WithMessage(err, "failed to transfer file")
	case ALLOCATING, DOWNLOADING, METADL, PAUSEDDL, QUEUEDDL, STALLEDDL, CHECKINGDL, FORCEDDL, CHECKINGRESUMEDATA, MOVING:
		m.tsk.SetStatus("qbittorrent downloading")
		return false, nil
	case ERROR, MISSINGFILES, UNKNOWN:
		return true, errors.Errorf("failed to download %s, error: %s", m.tsk.ID, info.State)
	}
	return true, errors.New("unknown error occurred downloading qbittorrent") // should never happen
}

var TransferTaskManager = task.NewTaskManager(3, func(k *uint64) {
	atomic.AddUint64(k, 1)
})

func (m *Monitor) complete() error {
	// check dstDir again
	storage, dstBaseDir, err := op.GetStorageAndActualPath(m.dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get storage")
	}
	// get files
	files, err := qbclient.GetFiles(m.tsk.ID)
	if err != nil {
		return errors.Wrapf(err, "failed to get files of %s", m.tsk.ID)
	}
	log.Debugf("files len: %d", len(files))
	// delete qbittorrent task but do not delete the files before transferring to avoid qbittorrent
	// accessing downloaded files and throw `cannot access the file because it is being used by another process` error
	// err = qbclient.Delete(m.tsk.ID, false)
	// if err != nil {
	// 	return err
	// }
	// upload files
	var wg sync.WaitGroup
	wg.Add(len(files))
	go func() {
		wg.Wait()
		m.finish <- struct{}{}
		if m.seedtime < 0 {
			log.Debugf("do not delete qb task %s", m.tsk.ID)
			return
		}
		log.Debugf("delete qb task %s after %d minutes", m.tsk.ID, m.seedtime)
		<-time.After(time.Duration(m.seedtime) * time.Minute)
		err := qbclient.Delete(m.tsk.ID, true)
		if err != nil {
			log.Errorln(err.Error())
		}
		err = os.RemoveAll(m.tempDir)
		if err != nil {
			log.Errorf("failed to remove qbittorrent temp dir: %+v", err.Error())
		}
	}()
	for _, file := range files {
		tempPath := filepath.Join(m.tempDir, file.Name)
		dstPath := filepath.Join(dstBaseDir, file.Name)
		dstDir := filepath.Dir(dstPath)
		fileName := filepath.Base(dstPath)
		TransferTaskManager.Submit(task.WithCancelCtx(&task.Task[uint64]{
			Name: fmt.Sprintf("transfer %s to [%s](%s)", tempPath, storage.GetStorage().MountPath, dstPath),
			Func: func(tsk *task.Task[uint64]) error {
				defer wg.Done()
				size := file.Size
				mimetype := utils.GetMimeType(tempPath)
				f, err := os.Open(tempPath)
				if err != nil {
					return errors.Wrapf(err, "failed to open file %s", tempPath)
				}
				s := stream.FileStream{
					Obj: &model.Object{
						Name:     fileName,
						Size:     size,
						Modified: time.Now(),
						IsFolder: false,
					},
					Reader:   f,
					Closers:  utils.NewClosers(f),
					Mimetype: mimetype,
				}
				ss, err := stream.NewSeekableStream(s, nil)
				if err != nil {
					return err
				}
				return op.Put(tsk.Ctx, storage, dstDir, ss, tsk.SetProgress)
			},
		}))
	}
	return nil
}

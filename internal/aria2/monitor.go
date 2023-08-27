package aria2

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/stream"
	"os"
	"path"
	"path/filepath"
	"strconv"
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
	retried    int
	c          chan int
	dstDirPath string
	finish     chan struct{}
}

func (m *Monitor) Loop() error {
	defer func() {
		notify.Signals.Delete(m.tsk.ID)
		// clear temp dir, should do while complete
		//_ = os.RemoveAll(m.tempDir)
	}()
	m.c = make(chan int)
	m.finish = make(chan struct{})
	notify.Signals.Store(m.tsk.ID, m.c)
	var (
		err error
		ok  bool
	)
outer:
	for {
		select {
		case <-m.tsk.Ctx.Done():
			_, err := client.Remove(m.tsk.ID)
			return err
		case <-m.c:
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

func (m *Monitor) Update() (bool, error) {
	info, err := client.TellStatus(m.tsk.ID)
	if err != nil {
		m.retried++
		log.Errorf("failed to get status of %s, retried %d times", m.tsk.ID, m.retried)
		return false, nil
	}
	if m.retried > 5 {
		return true, errors.Errorf("failed to get status of %s, retried %d times", m.tsk.ID, m.retried)
	}
	m.retried = 0
	if len(info.FollowedBy) != 0 {
		log.Debugf("followen by: %+v", info.FollowedBy)
		gid := info.FollowedBy[0]
		notify.Signals.Delete(m.tsk.ID)
		oldId := m.tsk.ID
		m.tsk.ID = gid
		DownTaskManager.RawTasks().Delete(oldId)
		DownTaskManager.RawTasks().Store(m.tsk.ID, m.tsk)
		notify.Signals.Store(gid, m.c)
		return false, nil
	}
	// update download status
	total, err := strconv.ParseUint(info.TotalLength, 10, 64)
	if err != nil {
		total = 0
	}
	downloaded, err := strconv.ParseUint(info.CompletedLength, 10, 64)
	if err != nil {
		downloaded = 0
	}
	progress := float64(downloaded) / float64(total) * 100
	m.tsk.SetProgress(int(progress))
	switch info.Status {
	case "complete":
		err := m.Complete()
		return true, errors.WithMessage(err, "failed to transfer file")
	case "error":
		return true, errors.Errorf("failed to download %s, error: %s", m.tsk.ID, info.ErrorMessage)
	case "active":
		m.tsk.SetStatus("aria2: " + info.Status)
		if info.Seeder == "true" {
			err := m.Complete()
			return true, errors.WithMessage(err, "failed to transfer file")
		}
		return false, nil
	case "waiting", "paused":
		m.tsk.SetStatus("aria2: " + info.Status)
		return false, nil
	case "removed":
		return true, errors.Errorf("failed to download %s, removed", m.tsk.ID)
	default:
		return true, errors.Errorf("failed to download %s, unknown status %s", m.tsk.ID, info.Status)
	}
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
	// get files
	files, err := client.GetFiles(m.tsk.ID)
	log.Debugf("files len: %d", len(files))
	if err != nil {
		return errors.Wrapf(err, "failed to get files of %s", m.tsk.ID)
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
				size, _ := strconv.ParseInt(file.Length, 10, 64)
				mimetype := utils.GetMimeType(file.Path)
				f, err := os.Open(file.Path)
				if err != nil {
					return errors.Wrapf(err, "failed to open file %s", file.Path)
				}
				s := stream.FileStream{
					Obj: &model.Object{
						Name:     path.Base(file.Path),
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
				relDir, err := filepath.Rel(m.tempDir, filepath.Dir(file.Path))
				if err != nil {
					log.Errorf("find relation directory error: %v", err)
				}
				newDistDir := filepath.Join(dstDirActualPath, relDir)
				return op.Put(tsk.Ctx, storage, newDistDir, ss, tsk.SetProgress)
			},
		}))
	}
	return nil
}

package aria2

import (
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"mime"
	"os"
	"path"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
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
	}
	if m.retried > 5 {
		return true, errors.Errorf("failed to get status of %s, retried %d times", m.tsk.ID, m.retried)
	}
	m.retried = 0
	if len(info.FollowedBy) != 0 {
		gid := info.FollowedBy[0]
		notify.Signals.Delete(m.tsk.ID)
		m.tsk.ID = gid
		notify.Signals.Store(gid, m.c)
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
	case "active", "waiting", "paused":
		m.tsk.SetStatus("aria2: " + info.Status)
		return false, nil
	case "removed":
		return true, errors.Errorf("failed to download %s, removed", m.tsk.ID)
	default:
		return true, errors.Errorf("failed to download %s, unknown status %s", m.tsk.ID, info.Status)
	}
}

var TransferTaskManager = task.NewTaskManager[uint64](3, func(k *uint64) {
	atomic.AddUint64(k, 1)
})

func (m *Monitor) Complete() error {
	// check dstDir again
	account, dstDirActualPath, err := operations.GetAccountAndActualPath(m.dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	// get files
	files, err := client.GetFiles(m.tsk.ID)
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
	for _, file := range files {
		TransferTaskManager.Submit(task.WithCancelCtx[uint64](&task.Task[uint64]{
			Name: fmt.Sprintf("transfer %s to [%s](%s)", file.Path, account.GetAccount().VirtualPath, dstDirActualPath),
			Func: func(tsk *task.Task[uint64]) error {
				defer wg.Done()
				size, _ := strconv.ParseInt(file.Length, 10, 64)
				mimetype := mime.TypeByExtension(path.Ext(file.Path))
				if mimetype == "" {
					mimetype = "application/octet-stream"
				}
				f, err := os.Open(file.Path)
				if err != nil {
					return errors.Wrapf(err, "failed to open file %s", file.Path)
				}
				stream := model.FileStream{
					Obj: model.Object{
						Name:     path.Base(file.Path),
						Size:     size,
						Modified: time.Now(),
						IsFolder: false,
					},
					ReadCloser: f,
					Mimetype:   mimetype,
				}
				return operations.Put(tsk.Ctx, account, dstDirActualPath, stream, tsk.SetProgress)
			},
		}))
	}
	return nil
}

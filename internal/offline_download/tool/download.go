package tool

import (
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/xhofe/tache"
	"sync"
	"time"
)

type DownloadTask struct {
	tache.Base
	Url          string       `json:"url"`
	DstDirPath   string       `json:"dst_dir_path"`
	TempDir      string       `json:"temp_dir"`
	DeletePolicy DeletePolicy `json:"delete_policy"`

	Status            string   `json:"status"`
	Signal            chan int `json:"-"`
	GID               string   `json:"-"`
	finish            chan struct{}
	tool              Tool
	callStatusRetried int
}

func (t *DownloadTask) Run() error {
	t.Signal = make(chan int)
	t.finish = make(chan struct{})
	defer func() {
		t.Signal = nil
		t.finish = nil
	}()
	gid, err := t.tool.AddURL(&AddUrlArgs{
		Url:     t.Url,
		UID:     t.ID,
		TempDir: t.TempDir,
		Signal:  t.Signal,
	})
	if err != nil {
		return err
	}
	t.GID = gid
	var (
		ok bool
	)
outer:
	for {
		select {
		case <-t.CtxDone():
			err := t.tool.Remove(t)
			return err
		case <-t.Signal:
			ok, err = t.Update()
			if ok {
				break outer
			}
		case <-time.After(time.Second * 3):
			ok, err = t.Update()
			if ok {
				break outer
			}
		}
	}
	if err != nil {
		return err
	}
	t.Status = "aria2 download completed, maybe transferring"
	t.finish <- struct{}{}
	t.Status = "offline download completed"
	return nil
}

// Update download status, return true if download completed
func (t *DownloadTask) Update() (bool, error) {
	info, err := t.tool.Status(t)
	if err != nil {
		t.callStatusRetried++
		log.Errorf("failed to get status of %s, retried %d times", t.ID, t.callStatusRetried)
		return false, nil
	}
	if t.callStatusRetried > 5 {
		return true, errors.Errorf("failed to get status of %s, retried %d times", t.ID, t.callStatusRetried)
	}
	t.callStatusRetried = 0
	t.SetProgress(info.Progress)
	t.Status = fmt.Sprintf("[%s]: %s", t.tool.Name(), info.Status)
	if info.NewGID != "" {
		log.Debugf("followen by: %+v", info.NewGID)
		t.GID = info.NewGID
		return false, nil
	}
	// if download completed
	if info.Completed {
		err := t.Complete()
		return true, errors.WithMessage(err, "failed to transfer file")
	}
	// if download failed
	if info.Err != nil {
		return true, errors.Errorf("failed to download %s, error: %s", t.ID, info.Err.Error())
	}
	return false, nil
}

func (t *DownloadTask) Complete() error {
	var (
		files []File
		err   error
	)
	if getFileser, ok := t.tool.(GetFileser); ok {
		files = getFileser.GetFiles(t)
	} else {
		files, err = GetFiles(t.TempDir)
		if err != nil {
			return errors.Wrapf(err, "failed to get files")
		}
	}
	// upload files
	var wg sync.WaitGroup
	wg.Add(len(files))
	go func() {
		wg.Wait()
		t.finish <- struct{}{}
	}()
	for i, _ := range files {
		file := files[i]
		TransferTaskManager.Add(&TransferTask{
			file:       file,
			dstDirPath: t.DstDirPath,
			wg:         &wg,
			tempDir:    t.TempDir,
		})
	}
	return nil
}

func (t *DownloadTask) GetName() string {
	return fmt.Sprintf("download %s to (%s)", t.Url, t.DstDirPath)
}

func (t *DownloadTask) GetStatus() string {
	return t.Status
}

var (
	DownloadTaskManager *tache.Manager[*DownloadTask]
)

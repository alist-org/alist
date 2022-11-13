package aria2

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/alist-org/alist/v3/drivers"
	conf2 "github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/task"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	conf2.Conf = conf2.DefaultConfig()
	absPath, err := filepath.Abs("../../data/temp")
	if err != nil {
		panic(err)
	}
	conf2.Conf.TempDir = absPath
	dB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.Init(dB)
}

func TestConnect(t *testing.T) {
	_, err := InitAria2Client("http://localhost:16800/jsonrpc", "secret", 3)
	if err != nil {
		t.Errorf("failed to init aria2: %+v", err)
	}
}

func TestDown(t *testing.T) {
	TestConnect(t)
	_, err := op.CreateStorage(context.Background(), model.Storage{
		ID:        0,
		MountPath: "/",
		Order:     0,
		Driver:    "Local",
		Status:    "",
		Addition:  `{"root_folder":"../../data"}`,
		Remark:    "",
	})
	if err != nil {
		t.Fatalf("failed to create storage: %+v", err)
	}
	err = AddURI(context.Background(), "https://nodejs.org/dist/index.json", "/test")
	if err != nil {
		t.Errorf("failed to add uri: %+v", err)
	}
	tasks := DownTaskManager.GetAll()
	if len(tasks) != 1 {
		t.Errorf("failed to get tasks: %+v", tasks)
	}
	for {
		tsk := tasks[0]
		t.Logf("task: %+v", tsk)
		if tsk.GetState() == task.SUCCEEDED {
			break
		}
		if tsk.GetState() == task.ERRORED {
			t.Fatalf("failed to download: %+v", tsk)
		}
		time.Sleep(time.Second)
	}
	for {
		if len(TransferTaskManager.GetAll()) == 0 {
			continue
		}
		tsk := TransferTaskManager.GetAll()[0]
		t.Logf("task: %+v", tsk)
		if tsk.GetState() == task.SUCCEEDED {
			break
		}
		if tsk.GetState() == task.ERRORED {
			t.Fatalf("failed to download: %+v", tsk)
		}
		time.Sleep(time.Second)
	}
}

package aria2

import (
	"context"
	"github.com/alist-org/alist/v3/conf"
	_ "github.com/alist-org/alist/v3/drivers"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/internal/store"
	"github.com/alist-org/alist/v3/pkg/task"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"path/filepath"
	"testing"
	"time"
)

func init() {
	conf.Conf = conf.DefaultConfig()
	absPath, err := filepath.Abs("../../data/temp")
	if err != nil {
		panic(err)
	}
	conf.Conf.TempDir = absPath
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	store.Init(db)
}

func TestConnect(t *testing.T) {
	err := InitAria2Client("http://localhost:16800/jsonrpc", "secret", 3)
	if err != nil {
		t.Errorf("failed to init aria2: %+v", err)
	}
}

func TestDown(t *testing.T) {
	TestConnect(t)
	err := operations.CreateAccount(context.Background(), model.Account{
		ID:          0,
		VirtualPath: "/",
		Index:       0,
		Driver:      "Local",
		Status:      "",
		Addition:    `{"root_folder":"../../data"}`,
		Remark:      "",
	})
	if err != nil {
		t.Fatalf("failed to create account: %+v", err)
	}
	err = AddURI(context.Background(), "https://nodejs.org/dist/index.json", "/test")
	if err != nil {
		t.Errorf("failed to add uri: %+v", err)
	}
	tasks := downTaskManager.GetAll()
	if len(tasks) != 1 {
		t.Errorf("failed to get tasks: %+v", tasks)
	}
	for {
		tsk := tasks[0]
		t.Logf("task: %+v", tsk)
		if tsk.GetState() == task.Succeeded {
			break
		}
		if tsk.GetState() == task.ERRORED {
			t.Fatalf("failed to download: %+v", tsk)
		}
		time.Sleep(time.Second)
	}
	for {
		if len(transferTaskManager.GetAll()) == 0 {
			continue
		}
		tsk := transferTaskManager.GetAll()[0]
		t.Logf("task: %+v", tsk)
		if tsk.GetState() == task.Succeeded {
			break
		}
		if tsk.GetState() == task.ERRORED {
			t.Fatalf("failed to download: %+v", tsk)
		}
		time.Sleep(time.Second)
	}
}

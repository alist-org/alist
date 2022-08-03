package operations_test

import (
	"context"
	"testing"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	dB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.Init(dB)
}

func TestCreateStorage(t *testing.T) {
	var storages = []struct {
		storage model.Storage
		isErr   bool
	}{
		{storage: model.Storage{Driver: "Local", MountPath: "/local", Addition: `{"root_folder":"."}`}, isErr: false},
		{storage: model.Storage{Driver: "Local", MountPath: "/local", Addition: `{"root_folder":"."}`}, isErr: true},
		{storage: model.Storage{Driver: "None", MountPath: "/none", Addition: `{"root_folder":"."}`}, isErr: true},
	}
	for _, storage := range storages {
		err := operations.CreateStorage(context.Background(), storage.storage)
		if err != nil {
			if !storage.isErr {
				t.Errorf("failed to create storage: %+v", err)
			} else {
				t.Logf("expect failed to create storage: %+v", err)
			}
		}
	}
}

func TestGetStorageVirtualFilesByPath(t *testing.T) {
	setupStorages(t)
	virtualFiles := operations.GetStorageVirtualFilesByPath("/a")
	var names []string
	for _, virtualFile := range virtualFiles {
		names = append(names, virtualFile.GetName())
	}
	var expectedNames = []string{"b", "c", "d"}
	if utils.SliceEqual(names, expectedNames) {
		t.Logf("passed")
	} else {
		t.Errorf("expected: %+v, got: %+v", expectedNames, names)
	}
}

func TestGetBalancedStorage(t *testing.T) {
	setupStorages(t)
	storage := operations.GetBalancedStorage("/a/d/e")
	if storage.GetStorage().MountPath != "/a/d/e" {
		t.Errorf("expected: /a/d/e, got: %+v", storage.GetStorage().MountPath)
	}
	storage = operations.GetBalancedStorage("/a/d/e")
	if storage.GetStorage().MountPath != "/a/d/e.balance" {
		t.Errorf("expected: /a/d/e.balance, got: %+v", storage.GetStorage().MountPath)
	}
}

func setupStorages(t *testing.T) {
	var storages = []model.Storage{
		{Driver: "Local", MountPath: "/a/b", Index: 0, Addition: `{"root_folder":"."}`},
		{Driver: "Local", MountPath: "/a/c", Index: 1, Addition: `{"root_folder":"."}`},
		{Driver: "Local", MountPath: "/a/d", Index: 2, Addition: `{"root_folder":"."}`},
		{Driver: "Local", MountPath: "/a/d/e", Index: 3, Addition: `{"root_folder":"."}`},
		{Driver: "Local", MountPath: "/a/d/e.balance", Index: 4, Addition: `{"root_folder":"."}`},
	}
	for _, storage := range storages {
		err := operations.CreateStorage(context.Background(), storage)
		if err != nil {
			t.Fatalf("failed to create storage: %+v", err)
		}
	}
}

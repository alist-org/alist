package operations_test

import (
	"context"
	"testing"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/internal/store"
	"github.com/alist-org/alist/v3/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	store.Init(db)
}

func TestCreateAccount(t *testing.T) {
	var accounts = []struct {
		account model.Account
		iserr   bool
	}{
		{account: model.Account{Driver: "Local", VirtualPath: "/local", Addition: "{}"}, iserr: false},
		{account: model.Account{Driver: "Local", VirtualPath: "/local", Addition: "{}"}, iserr: true},
		{account: model.Account{Driver: "None", VirtualPath: "/none", Addition: "{}"}, iserr: true},
	}
	for _, account := range accounts {
		err := operations.CreateAccount(context.Background(), account.account)
		if err != nil {
			if !account.iserr {
				t.Errorf("failed to create account: %+v", err)
			} else {
				t.Logf("expect failed to create account: %+v", err)
			}
		}
	}
}

func TestGetAccountVirtualFilesByPath(t *testing.T) {
	Setup(t)
	virtualFiles := operations.GetAccountVirtualFilesByPath("/a")
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

func Setup(t *testing.T) {
	var accounts = []model.Account{
		{Driver: "Local", VirtualPath: "/a/b", Index: 0, Addition: "{}"},
		{Driver: "Local", VirtualPath: "/a/c", Index: 1, Addition: "{}"},
		{Driver: "Local", VirtualPath: "/a/d", Index: 2, Addition: "{}"},
		{Driver: "Local", VirtualPath: "/a/d/e", Index: 3, Addition: "{}"},
		{Driver: "Local", VirtualPath: "/a/d/e.balance", Index: 4, Addition: "{}"},
	}
	for _, account := range accounts {
		err := operations.CreateAccount(context.Background(), account)
		if err != nil {
			t.Fatalf("failed to create account: %+v", err)
		}
	}
}

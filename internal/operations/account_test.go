package operations_test

import (
	"context"
	"testing"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/internal/store"
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

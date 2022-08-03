package db

import (
	"testing"

	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/pkg/errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	Init(db)
}

func TestCreateMeta(t *testing.T) {
	metas := []model.Meta{
		{Path: "/"},
		{Path: "/a"},
		{Path: "/a/b"},
		{Path: "/a/b/c"},
	}
	for _, meta := range metas {
		err := CreateMeta(&meta)
		if err != nil {
			t.Errorf("failed to create meta: %+v", err)
		}
	}
}

func TestUpdateMeta(t *testing.T) {
	meta := model.Meta{ID: 1, Path: "/b"}
	err := UpdateMeta(&meta)
	if err != nil {
		t.Errorf("failed to update meta: %+v", err)
	}
}

func TestGetNearestMeta1(t *testing.T) {
	meta, err := GetNearestMeta("/b/c/d")
	if err != nil {
		t.Errorf("failed to get nearest meta: %+v", err)
	}
	if meta.Path != "/b" {
		t.Errorf("unexpected meta: %+v", meta)
	}
}

func TestGetNearestMeta2(t *testing.T) {
	meta, err := GetNearestMeta("/c/d/e")
	if errors.Cause(err) != errs.MetaNotFound {
		t.Errorf("unexpected error: %+v", err)
		t.Errorf("unexpected meta: %+v", meta)
	}
}

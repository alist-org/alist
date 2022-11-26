package search

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
)

func Update(path string, objs []model.Obj) {
	
}

func init() {
	op.RegisterObjsUpdateHook(Update)
}

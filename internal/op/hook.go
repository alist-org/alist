package op

import "github.com/alist-org/alist/v3/internal/model"

type objsUpdateHook = func(path string, objs []model.Obj)

var (
	ObjsUpdateHooks = make([]objsUpdateHook, 0)
)

func RegisterObjsUpdateHook(hook objsUpdateHook) {
	ObjsUpdateHooks = append(ObjsUpdateHooks, hook)
}

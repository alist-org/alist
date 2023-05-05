package yaegi_storage

import (
	"context"
	"time"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
)

// api == v1
type DriverPluginHelper_V1 struct {
	// 必要字段

	IValue any

	WGetAddition func() driver.Additional
	WGetStorage  func() *model.Storage
	WSetStorage  func(model.Storage)

	WConfig func() driver.Config
	WInit   func(context.Context) error
	WDrop   func(context.Context) error

	WGetRoot func(context.Context) (model.Obj, error)
	WList    func(context.Context, model.Obj, model.ListArgs) ([]model.Obj, error)
	WLink    func(context.Context, model.Obj, model.LinkArgs) (*model.Link, error)

	// 可选字段

	WRemove func(context.Context, model.Obj) error
	WOther  func(context.Context, model.OtherArgs) (any, error)

	WCopy    func(context.Context, model.Obj, model.Obj) error
	WMakeDir func(context.Context, model.Obj, string) error
	WPut     func(context.Context, model.Obj, model.FileStreamer, driver.UpdateProgress) error
	WRename  func(context.Context, model.Obj, string) error
	WMove    func(context.Context, model.Obj, model.Obj) error

	WCopyResult    func(ctx context.Context, srcObj model.Obj, dstDir model.Obj) (model.Obj, error)
	WMakeDirResult func(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error)
	WPutResult     func(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error)
	WRenameResult  func(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error)
	WMoveResult    func(ctx context.Context, srcObj model.Obj, dstDir model.Obj) (model.Obj, error)
}

func (W *DriverPluginHelper_V1) GetAddition() driver.Additional {
	return W.WGetAddition()
}
func (W *DriverPluginHelper_V1) GetStorage() *model.Storage {
	return W.WGetStorage()
}
func (W *DriverPluginHelper_V1) SetStorage(s model.Storage) {
	W.WSetStorage(s)
}

func (W *DriverPluginHelper_V1) Config() driver.Config {
	return W.WConfig()
}
func (W *DriverPluginHelper_V1) Init(ctx context.Context) error {
	return W.WInit(ctx)
}
func (W *DriverPluginHelper_V1) Drop(ctx context.Context) error {
	if W.WDrop != nil {
		return W.WDrop(ctx)
	}
	return nil
}

func (W *DriverPluginHelper_V1) GetRoot(ctx context.Context) (model.Obj, error) {
	return W.WGetRoot(ctx)
}
func (W *DriverPluginHelper_V1) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	return W.WList(ctx, dir, args)
}
func (W *DriverPluginHelper_V1) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	return W.WLink(ctx, file, args)
}

func (W *DriverPluginHelper_V1) Remove(ctx context.Context, obj model.Obj) error {
	return W.WRemove(ctx, obj)
}
func (W *DriverPluginHelper_V1) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
	return W.WOther(ctx, args)
}

func (W *DriverPluginHelper_V1) Copy(ctx context.Context, srcObj model.Obj, dstDir model.Obj) (model.Obj, error) {
	if W.WCopyResult != nil {
		return W.WCopyResult(ctx, srcObj, dstDir)
	}
	if W.WCopy != nil {
		return nil, W.WCopy(ctx, srcObj, dstDir)
	}
	return nil, errs.NotImplement
}
func (W *DriverPluginHelper_V1) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	if W.WMakeDirResult != nil {
		return W.WMakeDirResult(ctx, parentDir, dirName)
	}
	if W.WMakeDir != nil {
		return nil, W.WMakeDir(ctx, parentDir, dirName)
	}
	return nil, errs.NotImplement
}
func (W *DriverPluginHelper_V1) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	if W.WPutResult != nil {
		return W.WPutResult(ctx, dstDir, stream, up)
	}
	if W.WPut != nil {
		return nil, W.WPut(ctx, dstDir, stream, up)
	}
	return nil, errs.NotImplement
}
func (W *DriverPluginHelper_V1) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	if W.WRenameResult != nil {
		return W.WRenameResult(ctx, srcObj, newName)
	}
	if W.WRename != nil {
		return nil, W.WRename(ctx, srcObj, newName)
	}
	return nil, errs.NotImplement
}
func (W *DriverPluginHelper_V1) Move(ctx context.Context, srcObj model.Obj, dstDir model.Obj) (model.Obj, error) {
	if W.WRenameResult != nil {
		return W.WMoveResult(ctx, srcObj, dstDir)
	}
	if W.WRename != nil {
		return nil, W.WMove(ctx, srcObj, dstDir)
	}
	return nil, errs.NotImplement
}

var _ interface {
	driver.Driver

	driver.GetRooter
	driver.Other

	driver.MkdirResult
	driver.MoveResult
	driver.RenameResult
	driver.CopyResult
	driver.PutResult
	driver.Remove
} = (*DriverPluginHelper_V1)(nil)

type ObjectHelper struct {
	IValue   any
	ID       string
	Path     string
	Name     string
	Size     int64
	Modified time.Time
	IsFolder bool
}

func (o *ObjectHelper) GetName() string {
	return o.Name
}

func (o *ObjectHelper) GetSize() int64 {
	return o.Size
}

func (o *ObjectHelper) ModTime() time.Time {
	return o.Modified
}

func (o *ObjectHelper) IsDir() bool {
	return o.IsFolder
}

func (o *ObjectHelper) GetID() string {
	return o.ID
}

func (o *ObjectHelper) GetPath() string {
	return o.Path
}

func (o *ObjectHelper) SetPath(id string) {
	o.Path = id
}

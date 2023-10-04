package driver

import (
	"context"

	"github.com/alist-org/alist/v3/internal/model"
)

type Driver interface {
	Meta
	Reader
	//Writer
	//Other
}

type Meta interface {
	Config() Config
	// GetStorage just get raw storage, no need to implement, because model.Storage have implemented
	GetStorage() *model.Storage
	SetStorage(model.Storage)
	// GetAddition Additional is used for unmarshal of JSON, so need return pointer
	GetAddition() Additional
	// Init If already initialized, drop first
	Init(ctx context.Context) error
	Drop(ctx context.Context) error
}

type Other interface {
	Other(ctx context.Context, args model.OtherArgs) (interface{}, error)
}

type Reader interface {
	// List files in the path
	// if identify files by path, need to set ID with path,like path.Join(dir.GetID(), obj.GetName())
	// if identify files by id, need to set ID with corresponding id
	List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error)
	// Link get url/filepath/reader of file
	Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error)
}

type GetRooter interface {
	GetRoot(ctx context.Context) (model.Obj, error)
}

type Getter interface {
	// Get file by path, the path haven't been joined with root path
	Get(ctx context.Context, path string) (model.Obj, error)
}

//type Writer interface {
//	Mkdir
//	Move
//	Rename
//	Copy
//	Remove
//	Put
//}

type Mkdir interface {
	MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error
}

type Move interface {
	Move(ctx context.Context, srcObj, dstDir model.Obj) error
}

type Rename interface {
	Rename(ctx context.Context, srcObj model.Obj, newName string) error
}

type Copy interface {
	Copy(ctx context.Context, srcObj, dstDir model.Obj) error
}

type Remove interface {
	Remove(ctx context.Context, obj model.Obj) error
}

type Put interface {
	Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up UpdateProgress) error
}

//type WriteResult interface {
//	MkdirResult
//	MoveResult
//	RenameResult
//	CopyResult
//	PutResult
//	Remove
//}

type MkdirResult interface {
	MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error)
}

type MoveResult interface {
	Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error)
}

type RenameResult interface {
	Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error)
}

type CopyResult interface {
	Copy(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error)
}

type PutResult interface {
	Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up UpdateProgress) (model.Obj, error)
}

type UpdateProgress func(percentage float64)

type Progress struct {
	Total int64
	Done  int64
	up    UpdateProgress
}

func (p *Progress) Write(b []byte) (n int, err error) {
	n = len(b)
	p.Done += int64(n)
	p.up(float64(p.Done) / float64(p.Total) * 100)
	return
}

func NewProgress(total int64, up UpdateProgress) *Progress {
	return &Progress{
		Total: total,
		up:    up,
	}
}

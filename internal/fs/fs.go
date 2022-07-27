package fs

import (
	"context"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	log "github.com/sirupsen/logrus"
	stdpath "path"
)

// the param named path of functions in this package is a virtual path
// So, the purpose of this package is to convert virtual path to actual path
// then pass the actual path to the operations package

type Fs interface {
	List(ctx context.Context, path string) ([]model.Obj, error)
	Get(ctx context.Context, path string) (model.Obj, error)
	Link(ctx context.Context, path string, args model.LinkArgs) (*model.Link, model.Obj, error)
	MakeDir(ctx context.Context, path string) error
	Move(ctx context.Context, srcPath string, dstDirPath string) error
	Copy(ctx context.Context, srcObjPath string, dstDirPath string) (bool, error)
	Rename(ctx context.Context, srcPath, dstName string) error
	Remove(ctx context.Context, path string) error
	PutDirectly(ctx context.Context, dstDirPath string, file model.FileStreamer) error
	PutAsTask(dstDirPath string, file model.FileStreamer) error
	GetStorage(path string) (driver.Driver, error)
}

type FileSystem struct {
	User *model.User
}

func (fs *FileSystem) List(ctx context.Context, path string) ([]model.Obj, error) {
	path = stdpath.Join(fs.User.BasePath, path)
	res, err := list(ctx, path)
	if err != nil {
		log.Errorf("failed list %s: %+v", path, err)
		return nil, err
	}
	return res, nil
}

func (fs *FileSystem) Get(ctx context.Context, path string) (model.Obj, error) {
	path = stdpath.Join(fs.User.BasePath, path)
	res, err := get(ctx, path)
	if err != nil {
		log.Errorf("failed get %s: %+v", path, err)
		return nil, err
	}
	return res, nil
}

func (fs *FileSystem) Link(ctx context.Context, path string, args model.LinkArgs) (*model.Link, model.Obj, error) {
	path = stdpath.Join(fs.User.BasePath, path)
	res, file, err := link(ctx, path, args)
	if err != nil {
		log.Errorf("failed link %s: %+v", path, err)
		return nil, nil, err
	}
	return res, file, nil
}

func (fs *FileSystem) MakeDir(ctx context.Context, path string) error {
	path = stdpath.Join(fs.User.BasePath, path)
	err := makeDir(ctx, path)
	if err != nil {
		log.Errorf("failed make dir %s: %+v", path, err)
	}
	return err
}

func (fs *FileSystem) Move(ctx context.Context, srcPath string, dstDirPath string) error {
	srcPath = stdpath.Join(fs.User.BasePath, srcPath)
	dstDirPath = stdpath.Join(fs.User.BasePath, dstDirPath)
	err := move(ctx, srcPath, dstDirPath)
	if err != nil {
		log.Errorf("failed move %s to %s: %+v", srcPath, dstDirPath, err)
	}
	return err
}

func (fs *FileSystem) Copy(ctx context.Context, srcPath string, dstDirPath string) (bool, error) {
	srcPath = stdpath.Join(fs.User.BasePath, srcPath)
	dstDirPath = stdpath.Join(fs.User.BasePath, dstDirPath)
	res, err := _copy(ctx, srcPath, dstDirPath)
	if err != nil {
		log.Errorf("failed copy %s to %s: %+v", srcPath, dstDirPath, err)
	}
	return res, err
}

func (fs *FileSystem) Rename(ctx context.Context, srcPath, dstName string) error {
	srcPath = stdpath.Join(fs.User.BasePath, srcPath)
	err := rename(ctx, srcPath, dstName)
	if err != nil {
		log.Errorf("failed rename %s to %s: %+v", srcPath, dstName, err)
	}
	return err
}

func (fs *FileSystem) Remove(ctx context.Context, path string) error {
	path = stdpath.Join(fs.User.BasePath, path)
	err := remove(ctx, path)
	if err != nil {
		log.Errorf("failed remove %s: %+v", path, err)
	}
	return err
}

func (fs *FileSystem) PutDirectly(ctx context.Context, dstDirPath string, file model.FileStreamer) error {
	dstDirPath = stdpath.Join(fs.User.BasePath, dstDirPath)
	err := putDirectly(ctx, dstDirPath, file)
	if err != nil {
		log.Errorf("failed put %s: %+v", dstDirPath, err)
	}
	return err
}

func (fs *FileSystem) PutAsTask(dstDirPath string, file model.FileStreamer) error {
	dstDirPath = stdpath.Join(fs.User.BasePath, dstDirPath)
	err := putAsTask(dstDirPath, file)
	if err != nil {
		log.Errorf("failed put %s: %+v", dstDirPath, err)
	}
	return err
}

func (fs *FileSystem) GetStorage(path string) (driver.Driver, error) {
	path = stdpath.Join(fs.User.BasePath, path)
	storageDriver, _, err := operations.GetStorageAndActualPath(path)
	if err != nil {
		return nil, err
	}
	return storageDriver, nil
}

var _ Fs = (*FileSystem)(nil)

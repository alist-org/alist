package operations

import (
	"context"
	stdpath "path"
	"time"

	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/singleflight"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
)

// In order to facilitate adding some other things before and after file operations

var filesCache = cache.NewMemCache(cache.WithShards[[]model.Obj](64))
var filesG singleflight.Group[[]model.Obj]

// List files in storage, not contains virtual file
func List(ctx context.Context, account driver.Driver, path string, refresh ...bool) ([]model.Obj, error) {
	dir, err := Get(ctx, account, path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get dir")
	}
	if account.Config().NoCache {
		return account.List(ctx, dir)
	}
	key := stdpath.Join(account.GetAccount().VirtualPath, path)
	if len(refresh) == 0 || !refresh[0] {
		if files, ok := filesCache.Get(key); ok {
			return files, nil
		}
	}
	files, err, _ := filesG.Do(key, func() ([]model.Obj, error) {
		files, err := account.List(ctx, dir)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to list files")
		}
		// TODO: maybe can get duration from account's config
		filesCache.Set(key, files, cache.WithEx[[]model.Obj](time.Minute*time.Duration(conf.Conf.CaCheExpiration)))
		return files, nil
	})
	return files, err
}

// Get object from list of files
func Get(ctx context.Context, account driver.Driver, path string) (model.Obj, error) {
	// is root folder
	if r, ok := account.GetAddition().(driver.IRootFolderId); ok && utils.PathEqual(path, "/") {
		return model.Object{
			ID:       r.GetRootFolderId(),
			Name:     "root",
			Size:     0,
			Modified: account.GetAccount().Modified,
			IsFolder: true,
		}, nil
	}
	if r, ok := account.GetAddition().(driver.IRootFolderPath); ok && utils.PathEqual(path, r.GetRootFolderPath()) {
		return model.Object{
			ID:       r.GetRootFolderPath(),
			Name:     "root",
			Size:     0,
			Modified: account.GetAccount().Modified,
			IsFolder: true,
		}, nil
	}
	// not root folder
	dir, name := stdpath.Split(path)
	files, err := List(ctx, account, dir)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get parent list")
	}
	for _, f := range files {
		if f.GetName() == name {
			// use path as id, why don't set id in List function?
			// because files maybe cache, set id here can reduce memory usage
			if f.GetID() == "" {
				if s, ok := f.(model.SetID); ok {
					s.SetID(path)
				}
			}
			return f, nil
		}
	}
	return nil, errors.WithStack(driver.ErrorObjectNotFound)
}

var linkCache = cache.NewMemCache(cache.WithShards[*model.Link](16))
var linkG singleflight.Group[*model.Link]

// Link get link, if is an url. should have an expiry time
func Link(ctx context.Context, account driver.Driver, path string, args model.LinkArgs) (*model.Link, error) {
	key := stdpath.Join(account.GetAccount().VirtualPath, path)
	if link, ok := linkCache.Get(key); ok {
		return link, nil
	}
	fn := func() (*model.Link, error) {
		file, err := Get(ctx, account, path)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to get file")
		}
		link, err := account.Link(ctx, file, args)
		if err != nil {
			return nil, errors.WithMessage(err, "failed get link")
		}
		if link.Expiration != nil {
			linkCache.Set(key, link, cache.WithEx[*model.Link](*link.Expiration))
		}
		return link, nil
	}
	link, err, _ := linkG.Do(key, fn)
	return link, err
}

func MakeDir(ctx context.Context, account driver.Driver, path string) error {
	// check if dir exists
	f, err := Get(ctx, account, path)
	if err != nil {
		if driver.IsErrObjectNotFound(err) {
			parentPath, dirName := stdpath.Split(path)
			err = MakeDir(ctx, account, parentPath)
			if err != nil {
				return errors.WithMessagef(err, "failed to make parent dir [%s]", parentPath)
			}
			parentDir, err := Get(ctx, account, parentPath)
			// this should not happen
			if err != nil {
				return errors.WithMessagef(err, "failed to get parent dir [%s]", parentPath)
			}
			return account.MakeDir(ctx, parentDir, dirName)
		} else {
			return errors.WithMessage(err, "failed to check if dir exists")
		}
	} else {
		// dir exists
		if f.IsDir() {
			return nil
		} else {
			// dir to make is a file
			return errors.New("file exists")
		}
	}
}

func Move(ctx context.Context, account driver.Driver, srcPath, dstDirPath string) error {
	srcObj, err := Get(ctx, account, srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get src object")
	}
	dstDir, err := Get(ctx, account, dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get dst dir")
	}
	return account.Move(ctx, srcObj, dstDir)
}

func Rename(ctx context.Context, account driver.Driver, srcPath, dstName string) error {
	srcObj, err := Get(ctx, account, srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get src object")
	}
	return account.Rename(ctx, srcObj, dstName)
}

// Copy Just copy file[s] in an account
func Copy(ctx context.Context, account driver.Driver, srcPath, dstDirPath string) error {
	srcObj, err := Get(ctx, account, srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get src object")
	}
	dstDir, err := Get(ctx, account, dstDirPath)
	return account.Copy(ctx, srcObj, dstDir)
}

func Remove(ctx context.Context, account driver.Driver, path string) error {
	obj, err := Get(ctx, account, path)
	if err != nil {
		// if object not found, it's ok
		if driver.IsErrObjectNotFound(err) {
			return nil
		}
		return errors.WithMessage(err, "failed to get object")
	}
	return account.Remove(ctx, obj)
}

func Put(ctx context.Context, account driver.Driver, parentPath string, file model.FileStreamer, up driver.UpdateProgress) error {
	err := MakeDir(ctx, account, parentPath)
	if err != nil {
		return errors.WithMessagef(err, "failed to make dir [%s]", parentPath)
	}
	parentDir, err := Get(ctx, account, parentPath)
	// this should not happen
	if err != nil {
		return errors.WithMessagef(err, "failed to get dir [%s]", parentPath)
	}
	// if up is nil, set a default to prevent panic
	if up == nil {
		up = func(p int) {}
	}
	return account.Put(ctx, parentDir, file, up)
}

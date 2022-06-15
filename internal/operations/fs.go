package operations

import (
	"context"
	stdpath "path"
	"time"

	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/singleflight"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
)

// In order to facilitate adding some other things before and after file operations

var filesCache = cache.NewMemCache(cache.WithShards[[]model.FileInfo](64))
var filesG singleflight.Group[[]model.FileInfo]

// List files in storage, not contains virtual file
func List(ctx context.Context, account driver.Driver, path string) ([]model.FileInfo, error) {
	if account.Config().NoCache {
		return account.List(ctx, path)
	}
	key := stdpath.Join(account.GetAccount().VirtualPath, path)
	if files, ok := filesCache.Get(key); ok {
		return files, nil
	}
	files, err, _ := filesG.Do(key, func() ([]model.FileInfo, error) {
		files, err := account.List(ctx, path)
		if err != nil {
			return nil, errors.WithMessage(err, "failed to list files")
		}
		// TODO: get duration from global config or account's config
		filesCache.Set(key, files, cache.WithEx[[]model.FileInfo](time.Minute*30))
		return files, nil
	})
	return files, err
}

func Get(ctx context.Context, account driver.Driver, path string) (model.FileInfo, error) {
	if r, ok := account.GetAddition().(driver.RootFolderId); ok && utils.PathEqual(path, "/") {
		return model.FileWithId{
			Id: r.GetRootFolderId(),
			File: model.File{
				Name:     "root",
				Size:     0,
				Modified: account.GetAccount().Modified,
				IsFolder: true,
			},
		}, nil
	}
	if r, ok := account.GetAddition().(driver.IRootFolderPath); ok && utils.PathEqual(path, r.GetRootFolderPath()) {
		return model.File{
			Name:     "root",
			Size:     0,
			Modified: account.GetAccount().Modified,
			IsFolder: true,
		}, nil
	}
	dir, name := stdpath.Split(path)
	files, err := List(ctx, account, dir)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get parent list")
	}
	for _, f := range files {
		if f.GetName() == name {
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
		link, err := account.Link(ctx, path, args)
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
	if f != nil && f.IsDir() {
		return nil
	}
	if err != nil && !driver.IsErrObjectNotFound(err) {
		return errors.WithMessage(err, "failed to check if dir exists")
	}
	parentPath := stdpath.Dir(path)
	err = MakeDir(ctx, account, parentPath)
	if err != nil {
		return errors.WithMessagef(err, "failed to make parent dir [%s]", parentPath)
	}
	return account.MakeDir(ctx, path)
}

func Move(ctx context.Context, account driver.Driver, srcPath, dstPath string) error {
	return account.Move(ctx, srcPath, dstPath)
}

func Rename(ctx context.Context, account driver.Driver, srcPath, dstName string) error {
	return account.Rename(ctx, srcPath, dstName)
}

// Copy Just copy file[s] in an account
func Copy(ctx context.Context, account driver.Driver, srcPath, dstPath string) error {
	return account.Copy(ctx, srcPath, dstPath)
}

func Remove(ctx context.Context, account driver.Driver, path string) error {
	return account.Remove(ctx, path)
}

func Put(ctx context.Context, account driver.Driver, parentPath string, file model.FileStreamer) error {
	return account.Put(ctx, parentPath, file)
}

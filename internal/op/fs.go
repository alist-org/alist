package op

import (
	"context"
	"os"
	stdpath "path"
	"strings"
	"time"

	"github.com/Xhofe/go-cache"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/singleflight"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// In order to facilitate adding some other things before and after file op

var listCache = cache.NewMemCache(cache.WithShards[[]model.Obj](64))
var listG singleflight.Group[[]model.Obj]

func ClearCache(storage driver.Driver, path string) {
	key := stdpath.Join(storage.GetStorage().MountPath, path)
	listCache.Del(key)
}

func Key(storage driver.Driver, path string) string {
	return stdpath.Join(storage.GetStorage().MountPath, utils.StandardizePath(path))
}

// List files in storage, not contains virtual file
func List(ctx context.Context, storage driver.Driver, path string, args model.ListArgs, refresh ...bool) ([]model.Obj, error) {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return nil, errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	path = utils.StandardizePath(path)
	log.Debugf("op.List %s", path)
	dir, err := Get(ctx, storage, path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get dir")
	}
	log.Debugf("list dir: %+v", dir)
	if !dir.IsDir() {
		return nil, errors.WithStack(errs.NotFolder)
	}
	if storage.Config().NoCache {
		objs, err := storage.List(ctx, dir, args)
		return objs, errors.WithStack(err)
	}
	key := Key(storage, path)
	if len(refresh) == 0 || !refresh[0] {
		if files, ok := listCache.Get(key); ok && len(files) > 0 {
			return files, nil
		}
	}
	objs, err, _ := listG.Do(key, func() ([]model.Obj, error) {
		files, err := storage.List(ctx, dir, args)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list objs")
		}
		listCache.Set(key, files, cache.WithEx[[]model.Obj](time.Minute*time.Duration(storage.GetStorage().CacheExpiration)))
		return files, nil
	})
	return objs, err
}

func isRoot(path, rootFolderPath string) bool {
	if utils.PathEqual(path, rootFolderPath) {
		return true
	}
	rootFolderPath = strings.TrimSuffix(rootFolderPath, "/")
	rootFolderPath = strings.TrimPrefix(rootFolderPath, "\\")
	// relative path, this shouldn't happen, because root folder path is absolute
	if utils.PathEqual(path, "/") && rootFolderPath == "." {
		return true
	}
	return false
}

// Get object from list of files
func Get(ctx context.Context, storage driver.Driver, path string) (model.Obj, error) {
	path = utils.StandardizePath(path)
	log.Debugf("op.Get %s", path)
	if g, ok := storage.(driver.Getter); ok {
		obj, err := g.Get(ctx, path)
		if err == nil {
			return obj, nil
		}
	}
	// is root folder
	if r, ok := storage.GetAddition().(driver.IRootId); ok && utils.PathEqual(path, "/") {
		return &model.Object{
			ID:       r.GetRootId(),
			Name:     "root",
			Size:     0,
			Modified: storage.GetStorage().Modified,
			IsFolder: true,
		}, nil
	}
	if r, ok := storage.GetAddition().(driver.IRootPath); ok && isRoot(path, r.GetRootPath()) {
		return &model.Object{
			Path:     r.GetRootPath(),
			Name:     "root",
			Size:     0,
			Modified: storage.GetStorage().Modified,
			IsFolder: true,
		}, nil
	}
	// not root folder
	dir, name := stdpath.Split(path)
	files, err := List(ctx, storage, dir, model.ListArgs{})
	if err != nil {
		return nil, errors.WithMessage(err, "failed get parent list")
	}
	for _, f := range files {
		// TODO maybe copy obj here
		if f.GetName() == name {
			// use path as id, why don't set id in List function?
			// because files maybe cache, set id here can reduce memory usage
			if f.GetPath() == "" {
				if s, ok := f.(model.SetPath); ok {
					s.SetPath(path)
				}
			}
			return f, nil
		}
	}
	return nil, errors.WithStack(errs.ObjectNotFound)
}

var linkCache = cache.NewMemCache(cache.WithShards[*model.Link](16))
var linkG singleflight.Group[*model.Link]

// Link get link, if is an url. should have an expiry time
func Link(ctx context.Context, storage driver.Driver, path string, args model.LinkArgs) (*model.Link, model.Obj, error) {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return nil, nil, errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	file, err := Get(ctx, storage, path)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "failed to get file")
	}
	if file.IsDir() {
		return nil, nil, errors.WithStack(errs.NotFile)
	}
	key := stdpath.Join(storage.GetStorage().MountPath, path) + ":" + args.IP
	if link, ok := linkCache.Get(key); ok {
		return link, file, nil
	}
	fn := func() (*model.Link, error) {
		link, err := storage.Link(ctx, file, args)
		if err != nil {
			return nil, errors.Wrapf(err, "failed get link")
		}
		if link.Expiration != nil {
			linkCache.Set(key, link, cache.WithEx[*model.Link](*link.Expiration))
		}
		return link, nil
	}
	link, err, _ := linkG.Do(key, fn)
	return link, file, err
}

// Other api
func Other(ctx context.Context, storage driver.Driver, args model.FsOtherArgs) (interface{}, error) {
	obj, err := Get(ctx, storage, args.Path)
	if err != nil {
		return nil, errors.WithMessagef(err, "failed to get obj")
	}
	if o, ok := storage.(driver.Other); ok {
		return o.Other(ctx, model.OtherArgs{
			Obj:    obj,
			Method: args.Method,
			Data:   args.Data,
		})
	} else {
		return nil, errs.NotImplement
	}
}

var mkdirG singleflight.Group[interface{}]

func MakeDir(ctx context.Context, storage driver.Driver, path string) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	path = utils.StandardizePath(path)
	key := Key(storage, path)
	_, err, _ := mkdirG.Do(key, func() (interface{}, error) {
		// check if dir exists
		f, err := Get(ctx, storage, path)
		if err != nil {
			if errs.IsObjectNotFound(err) {
				parentPath, dirName := stdpath.Split(path)
				err = MakeDir(ctx, storage, parentPath)
				if err != nil {
					return nil, errors.WithMessagef(err, "failed to make parent dir [%s]", parentPath)
				}
				parentDir, err := Get(ctx, storage, parentPath)
				// this should not happen
				if err != nil {
					return nil, errors.WithMessagef(err, "failed to get parent dir [%s]", parentPath)
				}
				err = storage.MakeDir(ctx, parentDir, dirName)
				if err == nil {
					ClearCache(storage, parentPath)
				}
				return nil, errors.WithStack(err)
			} else {
				return nil, errors.WithMessage(err, "failed to check if dir exists")
			}
		} else {
			// dir exists
			if f.IsDir() {
				return nil, nil
			} else {
				// dir to make is a file
				return nil, errors.New("file exists")
			}
		}
	})
	return err

}

func Move(ctx context.Context, storage driver.Driver, srcPath, dstDirPath string) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	srcObj, err := Get(ctx, storage, srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get src object")
	}
	dstDir, err := Get(ctx, storage, dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get dst dir")
	}
	return errors.WithStack(storage.Move(ctx, srcObj, dstDir))
}

func Rename(ctx context.Context, storage driver.Driver, srcPath, dstName string) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	srcObj, err := Get(ctx, storage, srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get src object")
	}
	return errors.WithStack(storage.Rename(ctx, srcObj, dstName))
}

// Copy Just copy file[s] in a storage
func Copy(ctx context.Context, storage driver.Driver, srcPath, dstDirPath string) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	srcObj, err := Get(ctx, storage, srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get src object")
	}
	dstDir, err := Get(ctx, storage, dstDirPath)
	return errors.WithStack(storage.Copy(ctx, srcObj, dstDir))
}

func Remove(ctx context.Context, storage driver.Driver, path string) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	obj, err := Get(ctx, storage, path)
	if err != nil {
		// if object not found, it's ok
		if errs.IsObjectNotFound(err) {
			return nil
		}
		return errors.WithMessage(err, "failed to get object")
	}
	err = storage.Remove(ctx, obj)
	if err == nil {
		key := Key(storage, stdpath.Dir(path))
		if objs, ok := listCache.Get(key); ok {
			j := -1
			for i, m := range objs {
				if m.GetName() == obj.GetName() {
					j = i
					break
				}
			}
			if j >= 0 && j < len(objs) {
				objs = append(objs[:j], objs[j+1:]...)
				listCache.Set(key, objs)
			} else {
				log.Debugf("not found obj")
			}
		} else {
			log.Debugf("not found parent cache")
		}
	}
	return errors.WithStack(err)
}

func Put(ctx context.Context, storage driver.Driver, dstDirPath string, file model.FileStreamer, up driver.UpdateProgress) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	defer func() {
		if f, ok := file.GetReadCloser().(*os.File); ok {
			err := os.RemoveAll(f.Name())
			if err != nil {
				log.Errorf("failed to remove file [%s]", f.Name())
			}
		}
	}()
	defer func() {
		if err := file.Close(); err != nil {
			log.Errorf("failed to close file streamer, %v", err)
		}
	}()
	// if file exist and size = 0, delete it
	dstPath := stdpath.Join(dstDirPath, file.GetName())
	fi, err := Get(ctx, storage, dstPath)
	if err == nil {
		if fi.GetSize() == 0 {
			err = Remove(ctx, storage, dstPath)
			if err != nil {
				return errors.WithMessagef(err, "failed remove file that exist and have size 0")
			}
		}
	}

	err = MakeDir(ctx, storage, dstDirPath)
	if err != nil {
		return errors.WithMessagef(err, "failed to make dir [%s]", dstDirPath)
	}
	parentDir, err := Get(ctx, storage, dstDirPath)
	// this should not happen
	if err != nil {
		return errors.WithMessagef(err, "failed to get dir [%s]", dstDirPath)
	}
	// if up is nil, set a default to prevent panic
	if up == nil {
		up = func(p int) {}
	}
	err = storage.Put(ctx, parentDir, file, up)
	log.Debugf("put file [%s] done", file.GetName())
	//if err == nil {
	//	//clear cache
	//	key := stdpath.Join(storage.GetStorage().MountPath, dstDirPath)
	//	listCache.Del(key)
	//}
	return errors.WithStack(err)
}

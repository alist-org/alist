package op

import (
	"context"
	"os"
	stdpath "path"
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
	listCache.Del(Key(storage, path))
}

func Key(storage driver.Driver, path string) string {
	return stdpath.Join(storage.GetStorage().MountPath, utils.FixAndCleanPath(path))
}

// List files in storage, not contains virtual file
func List(ctx context.Context, storage driver.Driver, path string, args model.ListArgs, refresh ...bool) ([]model.Obj, error) {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return nil, errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	path = utils.FixAndCleanPath(path)
	log.Debugf("op.List %s", path)
	key := Key(storage, path)
	if len(refresh) == 0 || !refresh[0] {
		if files, ok := listCache.Get(key); ok {
			log.Debugf("use cache when list %s", path)
			return files, nil
		}
	}
	dir, err := GetUnwrap(ctx, storage, path)
	if err != nil {
		return nil, errors.WithMessage(err, "failed get dir")
	}
	log.Debugf("list dir: %+v", dir)
	if !dir.IsDir() {
		return nil, errors.WithStack(errs.NotFolder)
	}
	objs, err, _ := listG.Do(key, func() ([]model.Obj, error) {
		files, err := storage.List(ctx, dir, args)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to list objs")
		}
		// set path
		for _, f := range files {
			if s, ok := f.(model.SetPath); ok && f.GetPath() == "" && dir.GetPath() != "" {
				s.SetPath(stdpath.Join(dir.GetPath(), f.GetName()))
			}
		}
		// warp obj name
		model.WrapObjsName(files)
		// call hooks
		go func(reqPath string, files []model.Obj) {
			for _, hook := range objsUpdateHooks {
				hook(args.ReqPath, files)
			}
		}(args.ReqPath, files)
		if !storage.Config().NoCache {
			if len(files) > 0 {
				log.Debugf("set cache: %s => %+v", key, files)
				listCache.Set(key, files, cache.WithEx[[]model.Obj](time.Minute*time.Duration(storage.GetStorage().CacheExpiration)))
			} else {
				log.Debugf("del cache: %s", key)
				listCache.Del(key)
			}
		}
		return files, nil
	})
	return objs, err
}

// Get object from list of files
func Get(ctx context.Context, storage driver.Driver, path string) (model.Obj, error) {
	path = utils.FixAndCleanPath(path)
	log.Debugf("op.Get %s", path)

	// is root folder
	if utils.PathEqual(path, "/") {
		var rootObj model.Obj
		switch r := storage.GetAddition().(type) {
		case driver.IRootId:
			rootObj = &model.Object{
				ID:       r.GetRootId(),
				Name:     RootName,
				Size:     0,
				Modified: storage.GetStorage().Modified,
				IsFolder: true,
			}
		case driver.IRootPath:
			rootObj = &model.Object{
				Path:     r.GetRootPath(),
				Name:     RootName,
				Size:     0,
				Modified: storage.GetStorage().Modified,
				IsFolder: true,
			}
		default:
			if storage, ok := storage.(driver.Getter); ok {
				obj, err := storage.GetRoot(ctx)
				if err != nil {
					return nil, errors.WithMessage(err, "failed get root obj")
				}
				rootObj = obj
			}
		}
		if rootObj == nil {
			return nil, errors.Errorf("please implement IRootPath or IRootId or Getter method")
		}
		return &model.ObjWrapName{
			Name: RootName,
			Obj:  rootObj,
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
			return f, nil
		}
	}
	log.Debugf("cant find obj with name: %s", name)
	return nil, errors.WithStack(errs.ObjectNotFound)
}

func GetUnwrap(ctx context.Context, storage driver.Driver, path string) (model.Obj, error) {
	obj, err := Get(ctx, storage, path)
	if err != nil {
		return nil, err
	}
	if unwrap, ok := obj.(model.UnwrapObj); ok {
		obj = unwrap.Unwrap()
	}
	return obj, err
}

var linkCache = cache.NewMemCache(cache.WithShards[*model.Link](16))
var linkG singleflight.Group[*model.Link]

// Link get link, if is an url. should have an expiry time
func Link(ctx context.Context, storage driver.Driver, path string, args model.LinkArgs) (*model.Link, model.Obj, error) {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return nil, nil, errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	file, err := GetUnwrap(ctx, storage, path)
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
	obj, err := GetUnwrap(ctx, storage, args.Path)
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
	path = utils.FixAndCleanPath(path)
	key := Key(storage, path)
	_, err, _ := mkdirG.Do(key, func() (interface{}, error) {
		// check if dir exists
		f, err := GetUnwrap(ctx, storage, path)
		if err != nil {
			if errs.IsObjectNotFound(err) {
				parentPath, dirName := stdpath.Split(path)
				err = MakeDir(ctx, storage, parentPath)
				if err != nil {
					return nil, errors.WithMessagef(err, "failed to make parent dir [%s]", parentPath)
				}
				parentDir, err := GetUnwrap(ctx, storage, parentPath)
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
	srcObj, err := GetUnwrap(ctx, storage, srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get src object")
	}
	dstDir, err := GetUnwrap(ctx, storage, dstDirPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get dst dir")
	}
	return errors.WithStack(storage.Move(ctx, srcObj, dstDir))
}

func Rename(ctx context.Context, storage driver.Driver, srcPath, dstName string) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	srcObj, err := GetUnwrap(ctx, storage, srcPath)
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
	srcObj, err := GetUnwrap(ctx, storage, srcPath)
	if err != nil {
		return errors.WithMessage(err, "failed to get src object")
	}
	dstDir, err := GetUnwrap(ctx, storage, dstDirPath)
	return errors.WithStack(storage.Copy(ctx, srcObj, dstDir))
}

func Remove(ctx context.Context, storage driver.Driver, path string) error {
	if storage.Config().CheckStatus && storage.GetStorage().Status != WORK {
		return errors.Errorf("storage not init: %s", storage.GetStorage().Status)
	}
	obj, err := GetUnwrap(ctx, storage, path)
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

func Put(ctx context.Context, storage driver.Driver, dstDirPath string, file *model.FileStream, up driver.UpdateProgress) error {
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
	fi, err := GetUnwrap(ctx, storage, dstPath)
	if err == nil {
		if fi.GetSize() == 0 {
			err = Remove(ctx, storage, dstPath)
			if err != nil {
				return errors.WithMessagef(err, "failed remove file that exist and have size 0")
			}
		} else {
			file.Old = fi
		}
	}
	err = MakeDir(ctx, storage, dstDirPath)
	if err != nil {
		return errors.WithMessagef(err, "failed to make dir [%s]", dstDirPath)
	}
	parentDir, err := GetUnwrap(ctx, storage, dstDirPath)
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

package local

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	stdpath "path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/disintegration/imaging"
)

type Local struct {
	model.Storage
	Addition
}

func (d *Local) Config() driver.Config {
	return config
}

func (d *Local) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
	if !utils.Exists(d.GetRootPath()) {
		err = fmt.Errorf("root folder %s not exists", d.GetRootPath())
	} else {
		if !filepath.IsAbs(d.GetRootPath()) {
			abs, err := filepath.Abs(d.GetRootPath())
			if err != nil {
				return err
			}
			d.SetRootPath(abs)
		}
	}
	return err
}

func (d *Local) Drop(ctx context.Context) error {
	return nil
}

func (d *Local) GetAddition() driver.Additional {
	return d.Addition
}

func (d *Local) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	fullPath := dir.GetPath()
	rawFiles, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	var files []model.Obj
	for _, f := range rawFiles {
		if !d.ShowHidden && strings.HasPrefix(f.Name(), ".") {
			continue
		}
		thumb := ""
		if d.Thumbnail && utils.GetFileType(f.Name()) == conf.IMAGE {
			thumb = common.GetApiUrl(nil) + stdpath.Join("/d", args.ReqPath, f.Name())
			thumb = utils.EncodePath(thumb, true)
			thumb += "?type=thumb"
		}
		isFolder := f.IsDir() || isSymlinkDir(f, fullPath)
		size := f.Size()
		if isFolder {
			size = 0
		}
		file := model.ObjThumb{
			Object: model.Object{
				Name:     f.Name(),
				Modified: f.ModTime(),
				Size:     size,
				IsFolder: isFolder,
			},
			Thumbnail: model.Thumbnail{
				Thumbnail: thumb,
			},
		}
		files = append(files, &file)
	}
	return files, nil
}

func (d *Local) Get(ctx context.Context, path string) (model.Obj, error) {
	f, err := os.Stat(path)
	if err != nil {
		if strings.Contains(err.Error(), "cannot find the file") {
			return nil, errs.ObjectNotFound
		}
		return nil, err
	}
	isFolder := f.IsDir() || isSymlinkDir(f, path)
	size := f.Size()
	if isFolder {
		size = 0
	}
	file := model.Object{
		Path:     path,
		Name:     f.Name(),
		Modified: f.ModTime(),
		Size:     size,
		IsFolder: isFolder,
	}
	return &file, nil
}

func (d *Local) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	fullPath := file.GetPath()
	var link model.Link
	if args.Type == "thumb" && utils.Ext(file.GetName()) != "svg" {
		imgData, err := ioutil.ReadFile(fullPath)
		if err != nil {
			return nil, err
		}
		srcBuf := bytes.NewBuffer(imgData)
		image, err := imaging.Decode(srcBuf)
		if err != nil {
			return nil, err
		}
		thumbImg := imaging.Resize(image, 144, 0, imaging.Lanczos)
		var buf bytes.Buffer
		err = imaging.Encode(&buf, thumbImg, imaging.PNG)
		if err != nil {
			return nil, err
		}
		size := buf.Len()
		link.Data = io.NopCloser(&buf)
		link.Header = http.Header{
			"Content-Length": []string{strconv.Itoa(size)},
		}
	} else {
		link.FilePath = &fullPath
	}
	return &link, nil
}

func (d *Local) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	fullPath := filepath.Join(parentDir.GetPath(), dirName)
	err := os.MkdirAll(fullPath, 0700)
	if err != nil {
		return err
	}
	return nil
}

func (d *Local) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	srcPath := srcObj.GetPath()
	dstPath := filepath.Join(dstDir.GetPath(), srcObj.GetName())
	err := os.Rename(srcPath, dstPath)
	if err != nil {
		return err
	}
	return nil
}

func (d *Local) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	srcPath := srcObj.GetPath()
	dstPath := filepath.Join(filepath.Dir(srcPath), newName)
	err := os.Rename(srcPath, dstPath)
	if err != nil {
		return err
	}
	return nil
}

func (d *Local) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	srcPath := srcObj.GetPath()
	dstPath := filepath.Join(dstDir.GetPath(), srcObj.GetName())
	var err error
	if srcObj.IsDir() {
		err = utils.CopyDir(srcPath, dstPath)
	} else {
		err = utils.CopyFile(srcPath, dstPath)
	}
	if err != nil {
		return err
	}
	return nil
}

func (d *Local) Remove(ctx context.Context, obj model.Obj) error {
	var err error
	if obj.IsDir() {
		err = os.RemoveAll(obj.GetPath())
	} else {
		err = os.Remove(obj.GetPath())
	}
	if err != nil {
		return err
	}
	return nil
}

func (d *Local) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	fullPath := filepath.Join(dstDir.GetPath(), stream.GetName())
	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
		if errors.Is(err, context.Canceled) {
			_ = os.Remove(fullPath)
		}
	}()
	err = utils.CopyWithCtx(ctx, out, stream, stream.GetSize(), up)
	if err != nil {
		return err
	}
	return nil
}

var _ driver.Driver = (*Local)(nil)

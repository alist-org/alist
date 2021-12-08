// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webdav

import (
	"context"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type FileSystem struct{}

func ParsePath(rawPath string) (*model.Account, string, base.Driver, error) {
	var internalPath, name string
	switch model.AccountsCount() {
	case 0:
		return nil, "", nil, fmt.Errorf("no accounts,please add one first")
	case 1:
		internalPath = rawPath
		break
	default:
		paths := strings.Split(rawPath, "/")
		internalPath = "/" + strings.Join(paths[2:], "/")
		name = paths[1]
	}
	account, ok := model.GetAccount(name)
	if !ok {
		return nil, "", nil, fmt.Errorf("no [%s] account", name)
	}
	driver, ok := base.GetDriver(account.Type)
	if !ok {
		return nil, "", nil, fmt.Errorf("no [%s] driver", account.Type)
	}
	return &account, internalPath, driver, nil
}

func (fs *FileSystem) File(rawPath string) (*model.File, error) {
	rawPath = utils.ParsePath(rawPath)
	if model.AccountsCount() > 1 && rawPath == "/" {
		now := time.Now()
		return &model.File{
			Name:      "root",
			Size:      0,
			Type:      conf.FOLDER,
			Driver:    "root",
			UpdatedAt: &now,
		}, nil
	}
	account, path_, driver, err := ParsePath(rawPath)
	if err != nil {
		return nil, err
	}
	return driver.File(path_, account)
}

func (fs *FileSystem) Files(rawPath string) ([]model.File, error) {
	rawPath = utils.ParsePath(rawPath)
	if model.AccountsCount() > 1 && rawPath == "/" {
		files, err := model.GetAccountFiles()
		if err != nil {
			return nil, err
		}
		return files, nil
	}
	account, path_, driver, err := ParsePath(rawPath)
	if err != nil {
		return nil, err
	}
	return driver.Files(path_, account)
}

func GetPW(path string, name string) string {
	if !conf.CheckDown {
		return ""
	}
	meta, err := model.GetMetaByPath(path)
	if err == nil {
		if meta.Password != "" {
			return utils.SignWithPassword(name, meta.Password)
		}
		return ""
	} else {
		if !conf.CheckParent {
			return ""
		}
		if path == "/" {
			return ""
		}
		return GetPW(utils.Dir(path), name)
	}
}

func (fs *FileSystem) Link(r *http.Request, rawPath string) (string, error) {
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("get link path: %s", rawPath)
	if model.AccountsCount() > 1 && rawPath == "/" {
		// error
	}
	account, path_, driver, err := ParsePath(rawPath)
	if err != nil {
		return "", err
	}
	link := ""
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}
	if driver.Config().OnlyProxy || account.WebdavProxy {
		link = fmt.Sprintf("%s://%s/p%s", protocol, r.Host, rawPath)
		if conf.CheckDown {
			pw := GetPW(utils.Dir(rawPath), utils.Base(rawPath))
			link += "?pw" + pw
		}
	} else {
		link, err = driver.Link(path_, account)
	}
	log.Debugf("webdav get link: %s", link)
	return link, err
}

func (fs *FileSystem) CreateDirectory(ctx context.Context, rawPath string) error {
	rawPath = utils.ParsePath(rawPath)
	if rawPath == "/" {
		return ErrNotImplemented
	}
	if model.AccountsCount() > 1 && len(strings.Split(rawPath, "/")) < 2 {
		return ErrNotImplemented
	}
	account, path_, driver, err := ParsePath(rawPath)
	if err != nil {
		return err
	}
	return driver.MakeDir(path_,account)
}

func (fs *FileSystem) Upload(ctx context.Context, r *http.Request, rawPath string) error {
	rawPath = utils.ParsePath(rawPath)
	if model.AccountsCount() > 1 && rawPath == "/" {
		return ErrNotImplemented
	}
	account, path_, driver, err := ParsePath(rawPath)
	if err != nil {
		return err
	}
	//fileSize, err := strconv.ParseUint(r.Header.Get("Content-Length"), 10, 64)
	fileSize := uint64(r.ContentLength)
	//if err != nil {
	//	return err
	//}
	filePath, fileName := filepath.Split(path_)
	fileData := model.FileStream{
		MIMEType:   r.Header.Get("Content-Type"),
		File:       r.Body,
		Size:       fileSize,
		Name:       fileName,
		ParentPath: filePath,
	}
	return driver.Upload(&fileData, account)
}

func (fs *FileSystem) Delete(rawPath string) error {
	rawPath = utils.ParsePath(rawPath)
	if rawPath == "/" {
		return ErrNotImplemented
	}
	if model.AccountsCount() > 1 && len(strings.Split(rawPath, "/")) < 2 {
		return ErrNotImplemented
	}
	account, path_, driver, err := ParsePath(rawPath)
	if err != nil {
		return err
	}
	return driver.Delete(path_, account)
}

// slashClean is equivalent to but slightly more efficient than
// path.Clean("/" + name).
func slashClean(name string) string {
	if name == "" || name[0] != '/' {
		name = "/" + name
	}
	return path.Clean(name)
}

// moveFiles moves files and/or directories from src to dst.
//
// See section 9.9.4 for when various HTTP status codes apply.
func moveFiles(ctx context.Context, fs *FileSystem, src string, dst string, overwrite bool) (status int, err error) {
	src = utils.ParsePath(src)
	dst = utils.ParsePath(dst)
	if src == dst {
		return http.StatusMethodNotAllowed, errDestinationEqualsSource
	}
	srcAccount, srcPath, driver, err := ParsePath(src)
	if err != nil {
		return http.StatusMethodNotAllowed, err
	}
	dstAccount, dstPath, _, err := ParsePath(dst)
	if err != nil {
		return http.StatusMethodNotAllowed, err
	}
	if srcAccount.Name != dstAccount.Name {
		return http.StatusMethodNotAllowed, errInvalidDestination
	}
	err = driver.Move(srcPath,dstPath,srcAccount)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusNoContent, nil
}

// copyFiles copies files and/or directories from src to dst.
//
// See section 9.8.5 for when various HTTP status codes apply.
func copyFiles(ctx context.Context, fs *FileSystem, src string, dst string, overwrite bool, depth int, recursion int) (status int, err error) {
	src = utils.ParsePath(src)
	dst = utils.ParsePath(dst)
	if src == dst {
		return http.StatusMethodNotAllowed, errDestinationEqualsSource
	}
	srcAccount, srcPath, driver, err := ParsePath(src)
	if err != nil {
		return http.StatusMethodNotAllowed, err
	}
	dstAccount, dstPath, _, err := ParsePath(dst)
	if err != nil {
		return http.StatusMethodNotAllowed, err
	}
	if srcAccount.Name != dstAccount.Name {
		// TODO 跨账号复制
		return http.StatusMethodNotAllowed, errInvalidDestination
	}
	err = driver.Copy(srcPath,dstPath,srcAccount)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusNoContent, nil
}

// walkFS traverses filesystem fs starting at name up to depth levels.
//
// Allowed values for depth are 0, 1 or infiniteDepth. For each visited node,
// walkFS calls walkFn. If a visited file system node is a directory and
// walkFn returns filepath.SkipDir, walkFS will skip traversal of this node.
func walkFS(
	ctx context.Context,
	fs *FileSystem,
	depth int,
	name string,
	info FileInfo,
	walkFn func(reqPath string, info FileInfo, err error) error) error {
	// This implementation is based on Walk's code in the standard path/filepath package.
	err := walkFn(name, info, nil)
	if err != nil {
		if info.IsDir() && err == filepath.SkipDir {
			return nil
		}
		return err
	}
	if !info.IsDir() || depth == 0 {
		return nil
	}
	if depth == 1 {
		depth = 0
	}

	files, err := fs.Files(name)
	if err != nil {
		return err
	}
	for _, fileInfo := range files {
		filename := path.Join(name, fileInfo.Name)
		err = walkFS(ctx, fs, depth, filename, &fileInfo, walkFn)
		if err != nil {
			if !fileInfo.IsDir() || err != filepath.SkipDir {
				return err
			}
		}
	}
	return nil
}

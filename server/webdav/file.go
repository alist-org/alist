// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webdav

import (
	"context"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

type FileSystem struct {}

func ParsePath(rawPath string) (*model.Account, string, drivers.Driver, error) {
	var path, name string
	switch model.AccountsCount() {
	case 0:
		return nil, "", nil, fmt.Errorf("no accounts,please add one first")
	case 1:
		path = rawPath
		break
	default:
		paths := strings.Split(rawPath, "/")
		path = "/" + strings.Join(paths[2:], "/")
		name = paths[1]
	}
	account, ok := model.GetAccount(name)
	if !ok {
		return nil, "", nil, fmt.Errorf("no [%s] account", name)
	}
	driver, ok := drivers.GetDriver(account.Type)
	if !ok {
		return nil, "", nil, fmt.Errorf("no [%s] driver", account.Type)
	}
	return &account, path, driver, nil
}

func (fs *FileSystem) File(rawPath string) (*model.File,error) {
	rawPath = utils.ParsePath(rawPath)
	if model.AccountsCount() > 1 && rawPath == "/" {
		return &model.File{
			Name:      "root",
			Size:      0,
			Type:      conf.FOLDER,
			Driver:    "root",
			UpdatedAt: nil,
		},nil
	}
	account, path_, driver, err := ParsePath(rawPath)
	if err != nil {
		return nil, err
	}
	return driver.File(path_,account)
}

func (fs *FileSystem) Files(rawPath string) ([]model.File,error) {
	rawPath = utils.ParsePath(rawPath)
	if model.AccountsCount() > 1 && rawPath == "/" {
		files, err := model.GetAccountFiles()
		if err != nil {
			return nil, err
		}
		return files,nil
	}
	account, path_, driver, err := ParsePath(rawPath)
	if err != nil {
		return nil, err
	}
	return driver.Files(path_,account)
}

func (fs *FileSystem) Link(rawPath string) (string, error) {
	rawPath = utils.ParsePath(rawPath)
	if model.AccountsCount() > 1 && rawPath == "/" {
		// error
	}
	account, path_, driver, err := ParsePath(rawPath)
	if err != nil {
		return "", err
	}
	return driver.Link(path_,account)
}

func (fs *FileSystem) CreateDirectory(ctx context.Context, reqPath string) (interface{}, error) {
	return nil, nil
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
func moveFiles(ctx context.Context, fs *FileSystem, src FileInfo, dst string, overwrite bool) (status int, err error) {

	return http.StatusNoContent, nil
}

// copyFiles copies files and/or directories from src to dst.
//
// See section 9.8.5 for when various HTTP status codes apply.
func copyFiles(ctx context.Context, fs *FileSystem, src FileInfo, dst string, overwrite bool, depth int, recursion int) (status int, err error) {

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

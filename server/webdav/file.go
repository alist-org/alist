// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webdav

import (
	"context"
	"fmt"
	"mime"
	"net"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/drivers/operate"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
)

type FileSystem struct{}

var upFileMap = make(map[string]*model.File)

func (fs *FileSystem) File(rawPath string) (*model.File, error) {
	rawPath = utils.ParsePath(rawPath)
	if f, ok := upFileMap[rawPath]; ok {
		return f, nil
	}
	account, path_, driver, err := common.ParsePath(rawPath)
	log.Debugln(account, path_, driver, err)
	if err != nil {
		if err.Error() == "path not found" {
			accountFiles := model.GetAccountFilesByPath(rawPath)
			if len(accountFiles) != 0 {
				now := time.Now()
				return &model.File{
					Name:      "root",
					Size:      0,
					Type:      conf.FOLDER,
					UpdatedAt: &now,
				}, nil
			}
		}
		return nil, err
	}
	file, err := operate.File(driver, account, path_)
	if err != nil && err.Error() == "path not found" {
		accountFiles := model.GetAccountFilesByPath(rawPath)
		if len(accountFiles) != 0 {
			now := time.Now()
			return &model.File{
				Name:      "root",
				Size:      0,
				Type:      conf.FOLDER,
				UpdatedAt: &now,
			}, nil
		}
	}
	return file, err
}

func (fs *FileSystem) Files(ctx context.Context, rawPath string) ([]model.File, error) {
	rawPath = utils.ParsePath(rawPath)
	//var files []model.File
	//var err error
	//if model.AccountsCount() > 1 && rawPath == "/" {
	//	files, err = model.GetAccountFilesByPath("/")
	//} else {
	//	account, path_, driver, err := common.ParsePath(rawPath)
	//	if err != nil {
	//		return nil, err
	//	}
	//	files, err = operate.Files(driver, account, path_)
	//}
	_, files, _, _, _, err := common.Path(rawPath)
	if err != nil {
		return nil, err
	}
	meta, _ := model.GetMetaByPath(rawPath)
	if visitor := ctx.Value("visitor"); visitor != nil {
		if visitor.(bool) {
			log.Debug("visitor")
			files = common.Hide(meta, files)
		}
	} else {
		log.Debug("admin")
	}
	return files, nil
}

func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

func (fs *FileSystem) Link(w http.ResponseWriter, r *http.Request, rawPath string) (string, error) {
	rawPath = utils.ParsePath(rawPath)
	log.Debugf("get link path: %s", rawPath)
	if model.AccountsCount() > 1 && rawPath == "/" {
		// error
	}
	account, path_, driver, err := common.ParsePath(rawPath)
	if err != nil {
		return "", err
	}
	link := ""
	protocol := "http"
	if r.TLS != nil {
		protocol = "https"
	}
	// 直接返回
	if account.WebdavDirect {
		file, err := fs.File(rawPath)
		if err != nil {
			return "", err
		}
		link_, err := driver.Link(base.Args{Path: path_, Header: r.Header}, account)
		if err != nil {
			return "", err
		}
		err = common.Proxy(w, r, link_, file)
		return "", err
	}
	if driver.Config().OnlyProxy || account.WebdavProxy {
		link = fmt.Sprintf("%s://%s/p%s", protocol, r.Host, rawPath)
		//if conf.GetBool("check down link") {
		sign := utils.SignWithToken(utils.Base(rawPath), conf.Token)
		link += "?sign=" + sign
		//}
	} else {
		link_, err := driver.Link(base.Args{Path: path_, IP: ClientIP(r)}, account)
		if err != nil {
			return "", err
		}
		link = link_.Url
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
	account, path_, driver, err := common.ParsePath(rawPath)
	if err != nil {
		return err
	}
	log.Debugf("mkdir: %s", path_)
	return operate.MakeDir(driver, account, path_, true)
}

func (fs *FileSystem) Upload(ctx context.Context, r *http.Request, rawPath string) (FileInfo, error) {
	rawPath = utils.ParsePath(rawPath)
	if model.AccountsCount() > 1 && rawPath == "/" {
		return nil, ErrNotImplemented
	}
	account, path_, driver, err := common.ParsePath(rawPath)
	if err != nil {
		return nil, err
	}
	fileSize := uint64(r.ContentLength)
	filePath, fileName := filepath.Split(path_)
	now := time.Now()
	fi := &model.File{
		Name:      fileName,
		Size:      0,
		UpdatedAt: &now,
	}
	if fileSize == 0 {
		// 如果文件大小为0，默认成功
		upFileMap[rawPath] = fi
		return fi, nil
	} else {
		delete(upFileMap, rawPath)
	}
	mimeType := r.Header.Get("Content-Type")
	if mimeType == "" || strings.ToLower(mimeType) == "application/octet-stream" {
		mimeTypeTmp := mime.TypeByExtension(path.Ext(fileName))
		if mimeTypeTmp != "" {
			mimeType = mimeTypeTmp
		} else {
			mimeType = "application/octet-stream"
		}
	}
	fileData := model.FileStream{
		MIMEType:   mimeType,
		File:       r.Body,
		Size:       fileSize,
		Name:       fileName,
		ParentPath: filePath,
	}
	return fi, operate.Upload(driver, account, &fileData, true)
}

func (fs *FileSystem) Delete(rawPath string) error {
	rawPath = utils.ParsePath(rawPath)
	if rawPath == "/" {
		return ErrNotImplemented
	}
	if model.AccountsCount() > 1 && len(strings.Split(rawPath, "/")) < 2 {
		return ErrNotImplemented
	}
	account, path_, driver, err := common.ParsePath(rawPath)
	if err != nil {
		return err
	}
	return operate.Delete(driver, account, path_, true)
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
	log.Debugf("move %s -> %s", src, dst)
	if src == dst {
		return http.StatusMethodNotAllowed, errDestinationEqualsSource
	}
	srcAccount, srcPath, driver, err := common.ParsePath(src)
	if err != nil {
		return http.StatusMethodNotAllowed, err
	}
	dstAccount, dstPath, _, err := common.ParsePath(dst)
	if err != nil {
		return http.StatusMethodNotAllowed, err
	}
	if srcAccount.Name != dstAccount.Name {
		return http.StatusMethodNotAllowed, errInvalidDestination
	}
	err = operate.Move(driver, srcAccount, srcPath, dstPath, true)
	if err != nil {
		log.Debug(err)
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
	log.Debugf("move %s -> %s", src, dst)
	if src == dst {
		return http.StatusMethodNotAllowed, errDestinationEqualsSource
	}
	srcAccount, srcPath, driver, err := common.ParsePath(src)
	if err != nil {
		return http.StatusMethodNotAllowed, err
	}
	dstAccount, dstPath, _, err := common.ParsePath(dst)
	if err != nil {
		return http.StatusMethodNotAllowed, err
	}
	if srcAccount.Name != dstAccount.Name {
		// TODO 跨账号复制
		return http.StatusMethodNotAllowed, errInvalidDestination
	}
	err = operate.Copy(driver, srcAccount, srcPath, dstPath, true)
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

	files, err := fs.Files(ctx, name)
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

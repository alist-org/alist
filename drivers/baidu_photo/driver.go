package baiduphoto

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

type BaiduPhoto struct {
	model.Storage
	Addition

	AccessToken string
	Uk          int64
	root        model.Obj
}

func (d *BaiduPhoto) Config() driver.Config {
	return config
}

func (d *BaiduPhoto) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *BaiduPhoto) Init(ctx context.Context) error {
	if err := d.refreshToken(); err != nil {
		return err
	}

	// root
	if d.AlbumID != "" {
		albumID := strings.Split(d.AlbumID, "|")[0]
		album, err := d.GetAlbumDetail(ctx, albumID)
		if err != nil {
			return err
		}
		d.root = album
	} else {
		d.root = &Root{
			Name:     "root",
			Modified: d.Modified,
			IsFolder: true,
		}
	}

	// uk
	info, err := d.uInfo()
	if err != nil {
		return err
	}
	d.Uk, err = strconv.ParseInt(info.YouaID, 10, 64)
	return err
}

func (d *BaiduPhoto) GetRoot(ctx context.Context) (model.Obj, error) {
	return d.root, nil
}

func (d *BaiduPhoto) Drop(ctx context.Context) error {
	d.AccessToken = ""
	d.Uk = 0
	d.root = nil
	return nil
}

func (d *BaiduPhoto) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	var err error

	/* album */
	if album, ok := dir.(*Album); ok {
		var files []AlbumFile
		files, err = d.GetAllAlbumFile(ctx, album, "")
		if err != nil {
			return nil, err
		}

		return utils.MustSliceConvert(files, func(file AlbumFile) model.Obj {
			return &file
		}), nil
	}

	/* root */
	var albums []Album
	if d.ShowType != "root_only_file" {
		albums, err = d.GetAllAlbum(ctx)
		if err != nil {
			return nil, err
		}
	}

	var files []File
	if d.ShowType != "root_only_album" {
		files, err = d.GetAllFile(ctx)
		if err != nil {
			return nil, err
		}
	}

	return append(
		utils.MustSliceConvert(albums, func(album Album) model.Obj {
			return &album
		}),
		utils.MustSliceConvert(files, func(album File) model.Obj {
			return &album
		})...,
	), nil

}

func (d *BaiduPhoto) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	switch file := file.(type) {
	case *File:
		return d.linkFile(ctx, file, args)
	case *AlbumFile:
		return d.linkAlbum(ctx, file, args)
	}
	return nil, errs.NotFile
}

var joinReg = regexp.MustCompile(`(?i)join:([\S]*)`)

func (d *BaiduPhoto) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	if _, ok := parentDir.(*Root); ok {
		code := joinReg.FindStringSubmatch(dirName)
		if len(code) > 1 {
			return d.JoinAlbum(ctx, code[1])
		}
		return d.CreateAlbum(ctx, dirName)
	}
	return nil, errs.NotSupport
}

func (d *BaiduPhoto) Copy(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	switch file := srcObj.(type) {
	case *File:
		if album, ok := dstDir.(*Album); ok {
			//rootfile ->  album
			return d.AddAlbumFile(ctx, album, file)
		}
	case *AlbumFile:
		switch album := dstDir.(type) {
		case *Root:
			//albumfile -> root
			return d.CopyAlbumFile(ctx, file)
		case *Album:
			// albumfile -> root -> album
			rootfile, err := d.CopyAlbumFile(ctx, file)
			if err != nil {
				return nil, err
			}
			return d.AddAlbumFile(ctx, album, rootfile)
		}
	}
	return nil, errs.NotSupport
}

func (d *BaiduPhoto) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	// 仅支持相册之间移动
	if file, ok := srcObj.(*AlbumFile); ok {
		if _, ok := dstDir.(*Album); ok {
			newObj, err := d.Copy(ctx, srcObj, dstDir)
			if err != nil {
				return nil, err
			}
			// 删除原相册文件
			_ = d.DeleteAlbumFile(ctx, file)
			return newObj, nil
		}
	}
	return nil, errs.NotSupport
}

func (d *BaiduPhoto) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	// 仅支持相册改名
	if album, ok := srcObj.(*Album); ok {
		return d.SetAlbumName(ctx, album, newName)
	}
	return nil, errs.NotSupport
}

func (d *BaiduPhoto) Remove(ctx context.Context, obj model.Obj) error {
	switch obj := obj.(type) {
	case *File:
		return d.DeleteFile(ctx, obj)
	case *AlbumFile:
		return d.DeleteAlbumFile(ctx, obj)
	case *Album:
		return d.DeleteAlbum(ctx, obj)
	}
	return errs.NotSupport
}

func (d *BaiduPhoto) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	// 需要获取完整文件md5,必须支持 io.Seek
	tempFile, err := utils.CreateTempFile(stream.GetReadCloser())
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	// 计算需要的数据
	const DEFAULT = 1 << 22
	const SliceSize = 1 << 18
	count := int(math.Ceil(float64(stream.GetSize()) / float64(DEFAULT)))

	sliceMD5List := make([]string, 0, count)
	fileMd5 := md5.New()
	sliceMd5 := md5.New()
	sliceMd52 := md5.New()
	slicemd52Write := utils.LimitWriter(sliceMd52, SliceSize)
	for i := 1; i <= count; i++ {
		if utils.IsCanceled(ctx) {
			return nil, ctx.Err()
		}

		_, err := io.CopyN(io.MultiWriter(fileMd5, sliceMd5, slicemd52Write), tempFile, DEFAULT)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, err
		}
		sliceMD5List = append(sliceMD5List, hex.EncodeToString(sliceMd5.Sum(nil)))
		sliceMd5.Reset()
	}
	if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	content_md5 := hex.EncodeToString(fileMd5.Sum(nil))
	slice_md5 := hex.EncodeToString(sliceMd52.Sum(nil))

	// 开始执行上传
	params := map[string]string{
		"autoinit":    "1",
		"isdir":       "0",
		"rtype":       "1",
		"ctype":       "11",
		"path":        stream.GetName(),
		"size":        fmt.Sprint(stream.GetSize()),
		"slice-md5":   slice_md5,
		"content-md5": content_md5,
		"block_list":  MustString(utils.Json.MarshalToString(sliceMD5List)),
	}

	// 预上传
	var precreateResp PrecreateResp
	_, err = d.Post(FILE_API_URL_V1+"/precreate", func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetFormData(params)
	}, &precreateResp)
	if err != nil {
		return nil, err
	}

	switch precreateResp.ReturnType {
	case 1: // 上传文件
		uploadParams := map[string]string{
			"method":   "upload",
			"path":     params["path"],
			"uploadid": precreateResp.UploadID,
		}

		for i := 0; i < count; i++ {
			if utils.IsCanceled(ctx) {
				return nil, ctx.Err()
			}
			uploadParams["partseq"] = fmt.Sprint(i)
			_, err = d.Post("https://c3.pcs.baidu.com/rest/2.0/pcs/superfile2", func(r *resty.Request) {
				r.SetContext(ctx)
				r.SetQueryParams(uploadParams)
				r.SetFileReader("file", stream.GetName(), io.LimitReader(tempFile, DEFAULT))
			}, nil)
			if err != nil {
				return nil, err
			}
			up(i * 100 / count)
		}
		fallthrough
	case 2: // 创建文件
		params["uploadid"] = precreateResp.UploadID
		_, err = d.Post(FILE_API_URL_V1+"/create", func(r *resty.Request) {
			r.SetContext(ctx)
			r.SetFormData(params)
		}, &precreateResp)
		if err != nil {
			return nil, err
		}
		fallthrough
	case 3: // 增加到相册
		rootfile := precreateResp.Data.toFile()
		if album, ok := dstDir.(*Album); ok {
			return d.AddAlbumFile(ctx, album, rootfile)
		}
		return rootfile, nil
	}
	return nil, errs.NotSupport
}

var _ driver.Driver = (*BaiduPhoto)(nil)
var _ driver.Getter = (*BaiduPhoto)(nil)
var _ driver.MkdirResult = (*BaiduPhoto)(nil)
var _ driver.CopyResult = (*BaiduPhoto)(nil)
var _ driver.MoveResult = (*BaiduPhoto)(nil)
var _ driver.Remove = (*BaiduPhoto)(nil)
var _ driver.PutResult = (*BaiduPhoto)(nil)
var _ driver.RenameResult = (*BaiduPhoto)(nil)

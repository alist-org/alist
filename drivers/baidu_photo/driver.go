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
}

func (d *BaiduPhoto) Config() driver.Config {
	return config
}

func (d *BaiduPhoto) GetAddition() driver.Additional {
	return d.Addition
}

func (d *BaiduPhoto) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
	return d.refreshToken()
}

func (d *BaiduPhoto) Drop(ctx context.Context) error {
	return nil
}

func (d *BaiduPhoto) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	var objs []model.Obj
	var err error
	if IsRoot(dir) {
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

		alubmName := make(map[string]int)
		objs, _ = utils.SliceConvert(albums, func(album Album) (model.Obj, error) {
			i := alubmName[album.GetName()]
			if i != 0 {
				alubmName[album.GetName()]++
				album.Title = fmt.Sprintf("%s(%d)", album.Title, i)
			}
			alubmName[album.GetName()]++
			return &album, nil
		})
		for i := 0; i < len(files); i++ {
			objs = append(objs, &files[i])
		}
	} else if IsAlbum(dir) || IsAlbumRoot(dir) {
		var files []AlbumFile
		files, err = d.GetAllAlbumFile(ctx, splitID(dir.GetID())[0], "")
		if err != nil {
			return nil, err
		}
		objs = make([]model.Obj, 0, len(files))
		for i := 0; i < len(files); i++ {
			objs = append(objs, &files[i])
		}
	}
	return objs, nil
}

func (d *BaiduPhoto) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	if IsAlbumFile(file) {
		return d.linkAlbum(ctx, file, args)
	} else if IsFile(file) {
		return d.linkFile(ctx, file, args)
	}
	return nil, errs.NotFile
}

func (d *BaiduPhoto) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	if IsRoot(parentDir) {
		code := regexp.MustCompile(`(?i)join:([\S]*)`).FindStringSubmatch(dirName)
		if len(code) > 1 {
			return d.JoinAlbum(ctx, code[1])
		}
		return d.CreateAlbum(ctx, dirName)
	}
	return errs.NotSupport
}

func (d *BaiduPhoto) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	if IsFile(srcObj) {
		if IsAlbum(dstDir) {
			//rootfile ->  album
			e := splitID(dstDir.GetID())
			return d.AddAlbumFile(ctx, e[0], e[1], srcObj.GetID())
		}
	} else if IsAlbumFile(srcObj) {
		if IsRoot(dstDir) {
			//albumfile -> root
			e := splitID(srcObj.GetID())
			_, err := d.CopyAlbumFile(ctx, e[1], e[2], e[3], srcObj.GetID())
			return err
		} else if IsAlbum(dstDir) {
			// albumfile -> root -> album
			e := splitID(srcObj.GetID())
			file, err := d.CopyAlbumFile(ctx, e[1], e[2], e[3], srcObj.GetID())
			if err != nil {
				return err
			}
			e = splitID(dstDir.GetID())
			return d.AddAlbumFile(ctx, e[0], e[1], fmt.Sprint(file.Fsid))
		}
	}
	return errs.NotSupport
}

func (d *BaiduPhoto) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// 仅支持相册之间移动
	if IsAlbumFile(srcObj) && IsAlbum(dstDir) {
		err := d.Copy(ctx, srcObj, dstDir)
		if err != nil {
			return err
		}
		e := splitID(srcObj.GetID())
		return d.DeleteAlbumFile(ctx, e[1], e[2], srcObj.GetID())
	}
	return errs.NotSupport
}

func (d *BaiduPhoto) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	// 仅支持相册改名
	if IsAlbum(srcObj) {
		e := splitID(srcObj.GetID())
		return d.SetAlbumName(ctx, e[0], e[1], newName)
	}
	return errs.NotSupport
}

func (d *BaiduPhoto) Remove(ctx context.Context, obj model.Obj) error {
	e := splitID(obj.GetID())
	if IsFile(obj) {
		return d.DeleteFile(ctx, e[0])
	} else if IsAlbum(obj) {
		return d.DeleteAlbum(ctx, e[0], e[1])
	} else if IsAlbumFile(obj) {
		return d.DeleteAlbumFile(ctx, e[1], e[2], obj.GetID())
	}
	return errs.NotSupport
}

func (d *BaiduPhoto) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// 需要获取完整文件md5,必须支持 io.Seek
	tempFile, err := utils.CreateTempFile(stream.GetReadCloser())
	if err != nil {
		return err
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
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		_, err := io.CopyN(io.MultiWriter(fileMd5, sliceMd5, slicemd52Write), tempFile, DEFAULT)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return err
		}
		sliceMD5List = append(sliceMD5List, hex.EncodeToString(sliceMd5.Sum(nil)))
		sliceMd5.Reset()
	}
	if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
		return err
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
		return err
	}

	switch precreateResp.ReturnType {
	case 1: // 上传文件
		uploadParams := map[string]string{
			"method":   "upload",
			"path":     params["path"],
			"uploadid": precreateResp.UploadID,
		}

		for i := 0; i < count; i++ {
			uploadParams["partseq"] = fmt.Sprint(i)
			_, err = d.Post("https://c3.pcs.baidu.com/rest/2.0/pcs/superfile2", func(r *resty.Request) {
				r.SetContext(ctx)
				r.SetQueryParams(uploadParams)
				r.SetFileReader("file", stream.GetName(), io.LimitReader(tempFile, DEFAULT))
			}, nil)
			if err != nil {
				return err
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
			return err
		}
		fallthrough
	case 3: // 增加到相册
		if IsAlbum(dstDir) || IsAlbumRoot(dstDir) {
			e := splitID(dstDir.GetID())
			err = d.AddAlbumFile(ctx, e[0], e[1], fmt.Sprint(precreateResp.Data.FsID))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

var _ driver.Driver = (*BaiduPhoto)(nil)

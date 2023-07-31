package weiyun

import (
	"context"
	"io"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/cron"
	"github.com/alist-org/alist/v3/pkg/utils"
	weiyunsdkgo "github.com/foxxorcat/weiyun-sdk-go"
)

type WeiYun struct {
	model.Storage
	Addition

	client     *weiyunsdkgo.WeiYunClient
	cron       *cron.Cron
	rootFolder *Folder
}

func (d *WeiYun) Config() driver.Config {
	return config
}

func (d *WeiYun) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *WeiYun) Init(ctx context.Context) error {
	d.client = weiyunsdkgo.NewWeiYunClientWithRestyClient(base.RestyClient)
	err := d.client.SetCookiesStr(d.Cookies).RefreshCtoken()
	if err != nil {
		return err
	}

	// Cookie过期回调
	d.client.SetOnCookieExpired(func(err error) {
		d.Status = err.Error()
		op.MustSaveDriverStorage(d)
	})

	// cookie更新回调
	d.client.SetOnCookieUpload(func(c []*http.Cookie) {
		d.Cookies = weiyunsdkgo.CookieToString(weiyunsdkgo.ClearCookie(c))
		op.MustSaveDriverStorage(d)
	})

	// qqCookie保活
	if d.client.LoginType() == 1 {
		d.cron = cron.NewCron(time.Minute * 5)
		d.cron.Do(func() {
			d.client.KeepAlive()
		})
	}

	// 获取默认根目录dirKey
	if d.RootFolderID == "" {
		userInfo, err := d.client.DiskUserInfoGet()
		if err != nil {
			return err
		}
		d.RootFolderID = userInfo.MainDirKey
	}

	// 处理目录ID，找到PdirKey
	folders, err := d.client.LibDirPathGet(d.RootFolderID)
	if err != nil {
		return err
	}
	folder := folders[len(folders)-1]
	d.rootFolder = &Folder{
		PFolder: &Folder{
			Folder: weiyunsdkgo.Folder{
				DirKey: folder.PdirKey,
			},
		},
		Folder: folder.Folder,
	}
	return nil
}

func (d *WeiYun) Drop(ctx context.Context) error {
	d.client = nil
	if d.cron != nil {
		d.cron.Stop()
		d.cron = nil
	}
	return nil
}

func (d *WeiYun) GetRoot(ctx context.Context) (model.Obj, error) {
	return d.rootFolder, nil
}

func (d *WeiYun) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	if folder, ok := dir.(*Folder); ok {
		var files []model.Obj
		for {
			data, err := d.client.DiskDirFileList(folder.GetID(), weiyunsdkgo.WarpParamOption(
				weiyunsdkgo.QueryFileOptionOffest(int64(len(files))),
				weiyunsdkgo.QueryFileOptionGetType(weiyunsdkgo.FileAndDir),
				weiyunsdkgo.QueryFileOptionSort(func() weiyunsdkgo.OrderBy {
					switch d.OrderBy {
					case "name":
						return weiyunsdkgo.FileName
					case "size":
						return weiyunsdkgo.FileSize
					case "updated_at":
						return weiyunsdkgo.FileMtime
					default:
						return weiyunsdkgo.FileName
					}
				}(), d.OrderDirection == "desc"),
			))
			if err != nil {
				return nil, err
			}

			if files == nil {
				files = make([]model.Obj, 0, data.TotalDirCount+data.TotalFileCount)
			}

			for _, dir := range data.DirList {
				files = append(files, &Folder{
					PFolder: folder,
					Folder:  dir,
				})
			}

			for _, file := range data.FileList {
				files = append(files, &File{
					PFolder: folder,
					File:    file,
				})
			}

			if data.FinishFlag || len(data.DirList)+len(data.FileList) == 0 {
				return files, nil
			}
		}
	}
	return nil, errs.NotSupport
}

func (d *WeiYun) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	if file, ok := file.(*File); ok {
		data, err := d.client.DiskFileDownload(weiyunsdkgo.FileParam{PdirKey: file.GetPKey(), FileID: file.GetID()})
		if err != nil {
			return nil, err
		}
		return &model.Link{
			URL: data.DownloadUrl,
			Header: http.Header{
				"Cookie": []string{data.CookieName + "=" + data.CookieValue},
			},
		}, nil
	}
	return nil, errs.NotSupport
}

func (d *WeiYun) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	if folder, ok := parentDir.(*Folder); ok {
		newFolder, err := d.client.DiskDirCreate(weiyunsdkgo.FolderParam{
			PPdirKey: folder.GetPKey(),
			PdirKey:  folder.DirKey,
			DirName:  dirName,
		})
		if err != nil {
			return nil, err
		}
		return &Folder{
			PFolder: folder,
			Folder:  *newFolder,
		}, nil
	}
	return nil, errs.NotSupport
}

func (d *WeiYun) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	if dstDir, ok := dstDir.(*Folder); ok {
		dstParam := weiyunsdkgo.FolderParam{
			PdirKey: dstDir.GetPKey(),
			DirKey:  dstDir.GetID(),
			DirName: dstDir.GetName(),
		}
		switch srcObj := srcObj.(type) {
		case *File:
			err := d.client.DiskFileMove(weiyunsdkgo.FileParam{
				PPdirKey: srcObj.PFolder.GetPKey(),
				PdirKey:  srcObj.GetPKey(),
				FileID:   srcObj.GetID(),
				FileName: srcObj.GetName(),
			}, dstParam)
			if err != nil {
				return nil, err
			}

			return &File{
				PFolder: dstDir,
				File:    srcObj.File,
			}, nil
		case *Folder:
			err := d.client.DiskDirMove(weiyunsdkgo.FolderParam{
				PPdirKey: srcObj.PFolder.GetPKey(),
				PdirKey:  srcObj.GetPKey(),
				DirKey:   srcObj.GetID(),
				DirName:  srcObj.GetName(),
			}, dstParam)
			if err != nil {
				return nil, err
			}

			return &Folder{
				PFolder: dstDir,
				Folder:  srcObj.Folder,
			}, nil
		}
	}
	return nil, errs.NotSupport
}

func (d *WeiYun) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	switch srcObj := srcObj.(type) {
	case *File:
		err := d.client.DiskFileRename(weiyunsdkgo.FileParam{
			PPdirKey: srcObj.PFolder.GetPKey(),
			PdirKey:  srcObj.GetPKey(),
			FileID:   srcObj.GetID(),
			FileName: srcObj.GetName(),
		}, newName)
		if err != nil {
			return nil, err
		}
		newFile := srcObj.File
		newFile.FileName = newName
		newFile.FileCtime = weiyunsdkgo.TimeStamp(time.Now())
		return &File{
			PFolder: srcObj.PFolder,
			File:    newFile,
		}, nil
	case *Folder:
		err := d.client.DiskDirAttrModify(weiyunsdkgo.FolderParam{
			PPdirKey: srcObj.PFolder.GetPKey(),
			PdirKey:  srcObj.GetPKey(),
			DirKey:   srcObj.GetID(),
			DirName:  srcObj.GetName(),
		}, newName)
		if err != nil {
			return nil, err
		}

		newFolder := srcObj.Folder
		newFolder.DirName = newName
		newFolder.DirCtime = weiyunsdkgo.TimeStamp(time.Now())
		return &Folder{
			PFolder: srcObj.PFolder,
			Folder:  newFolder,
		}, nil
	}
	return nil, errs.NotSupport
}

func (d *WeiYun) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj, optional
	return errs.NotImplement
}

func (d *WeiYun) Remove(ctx context.Context, obj model.Obj) error {
	switch obj := obj.(type) {
	case *File:
		return d.client.DiskFileDelete(weiyunsdkgo.FileParam{
			PPdirKey: obj.PFolder.GetPKey(),
			PdirKey:  obj.GetPKey(),
			FileID:   obj.GetID(),
			FileName: obj.GetName(),
		})
	case *Folder:
		return d.client.DiskDirDelete(weiyunsdkgo.FolderParam{
			PPdirKey: obj.PFolder.GetPKey(),
			PdirKey:  obj.GetPKey(),
			DirKey:   obj.GetID(),
			DirName:  obj.GetName(),
		})
	}
	// TODO remove obj, optional
	return errs.NotSupport
}

func (d *WeiYun) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	if folder, ok := dstDir.(*Folder); ok {
		file, err := utils.CreateTempFile(stream)
		if err != nil {
			return nil, err
		}
		defer func() {
			_ = file.Close()
			_ = os.Remove(file.Name())
		}()

		// step 1.
		preData, err := d.client.PreUpload(ctx, weiyunsdkgo.UpdloadFileParam{
			PdirKey: folder.GetPKey(),
			DirKey:  folder.DirKey,

			FileName: stream.GetName(),
			FileSize: stream.GetSize(),
			File:     file,

			ChannelCount:    4,
			FileExistOption: 1,
		})
		if err != nil {
			return nil, err
		}

		// fast upload
		if !preData.FileExist {
			// step 2.
			upCtx, cancel := context.WithCancelCause(ctx)
			var wg sync.WaitGroup
			for _, channel := range preData.ChannelList {
				wg.Add(1)
				go func(channel weiyunsdkgo.UploadChannelData) {
					defer wg.Done()
					if utils.IsCanceled(upCtx) {
						return
					}
					for {
						channel.Len = int(math.Min(float64(stream.GetSize()-channel.Offset), float64(channel.Len)))
						upData, err := d.client.UploadFile(upCtx, channel, preData.UploadAuthData,
							io.NewSectionReader(file, channel.Offset, int64(channel.Len)))
						if err != nil {
							cancel(err)
							return
						}
						// 上传完成
						if upData.UploadState != 1 {
							return
						}
						channel = upData.Channel
					}
				}(channel)
			}
			wg.Wait()
			if utils.IsCanceled(upCtx) {
				return nil, context.Cause(upCtx)
			}
		}

		return &File{
			PFolder: folder,
			File:    preData.File,
		}, nil
	}
	return nil, errs.NotSupport
}

// func (d *WeiYun) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
// 	return nil, errs.NotSupport
// }

var _ driver.Driver = (*WeiYun)(nil)
var _ driver.GetRooter = (*WeiYun)(nil)
var _ driver.MkdirResult = (*WeiYun)(nil)

// var _ driver.CopyResult = (*WeiYun)(nil)
var _ driver.MoveResult = (*WeiYun)(nil)
var _ driver.Remove = (*WeiYun)(nil)

var _ driver.PutResult = (*WeiYun)(nil)
var _ driver.RenameResult = (*WeiYun)(nil)

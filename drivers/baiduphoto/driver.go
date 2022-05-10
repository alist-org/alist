package baiduphoto

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"os"

	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
)

type Baidu struct{}

func init() {
	base.RegisterDriver(new(Baidu))
}

func (driver Baidu) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:      "Baidu.Photo",
		LocalSort: true,
	}
}

func (driver Baidu) Items() []base.Item {
	return []base.Item{
		{
			Name:     "refresh_token",
			Label:    "refresh token",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:  "root_folder",
			Label: "album_id",
			Type:  base.TypeString,
		},
		{
			Name:     "internal_type",
			Label:    "download api",
			Type:     base.TypeSelect,
			Required: true,
			Values:   "file,album",
			Default:  "album",
		},
		{
			Name:     "client_id",
			Label:    "client id",
			Default:  "iYCeC9g08h5vuP9UqvPHKKSVrKFXGa1v",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "client_secret",
			Label:    "client secret",
			Default:  "jXiFMOPVPCWlO2M5CwWQzffpNPaGTRBG",
			Type:     base.TypeString,
			Required: true,
		},
	}
}

func (driver Baidu) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	return driver.RefreshToken(account)
}

func (driver Baidu) File(path string, account *model.Account) (*model.File, error) {
	path = utils.ParsePath(path)
	if path == "/" {
		return &model.File{
			Id:        account.RootFolder,
			Name:      account.Name,
			Size:      0,
			Type:      conf.FOLDER,
			Driver:    driver.Config().Name,
			UpdatedAt: account.UpdatedAt,
		}, nil
	}

	dir, name := utils.Split(path)
	files, err := driver.Files(dir, account)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		if file.Name == name {
			return &file, nil
		}
	}
	return nil, base.ErrPathNotFound
}

func (driver Baidu) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	var files []model.File
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ = cache.([]model.File)
		return files, nil
	}

	file, err := driver.File(path, account)
	if err != nil {
		return nil, err
	}

	if IsAlbum(file) {
		albumFiles, err := driver.GetAllAlbumFile(file.Id, account)
		if err != nil {
			return nil, err
		}
		files = make([]model.File, 0, len(albumFiles))
		for _, file := range albumFiles {
			var thumbnail string
			if len(file.Thumburl) > 0 {
				thumbnail = file.Thumburl[0]
			}
			files = append(files, model.File{
				Id:        joinID(file.Fsid, file.Uk, file.Tid),
				Name:      file.Name(),
				Size:      file.Size,
				Type:      utils.GetFileType(utils.Ext(file.Path)),
				Driver:    driver.Config().Name,
				UpdatedAt: getTime(file.Mtime),
				Thumbnail: thumbnail,
			})
		}
	} else if IsRoot(file) {
		albums, err := driver.GetAllAlbum(account)
		if err != nil {
			return nil, err
		}

		files = make([]model.File, 0, len(albums))
		for _, album := range albums {
			files = append(files, model.File{
				Id:        joinID(album.AlbumID, album.Tid),
				Name:      album.Title,
				Size:      0,
				Type:      conf.FOLDER,
				Driver:    driver.Config().Name,
				UpdatedAt: getTime(album.Mtime),
			})
		}
	} else {
		return nil, base.ErrNotSupport
	}

	if len(files) > 0 {
		_ = base.SetCache(path, files, account)
	}
	return files, nil
}

func (driver Baidu) Link(args base.Args, account *model.Account) (*base.Link, error) {
	if account.InternalType == "file" {
		return driver.LinkFile(args, account)
	}
	return driver.LinkAlbum(args, account)
}

func (driver Baidu) LinkAlbum(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}
	if !IsAlbumFile(file) {
		return nil, base.ErrNotSupport
	}

	album, err := driver.File(utils.Dir(utils.ParsePath(args.Path)), account)
	if err != nil {
		return nil, err
	}

	e := splitID(file.Id)
	res, err := base.NoRedirectClient.R().
		SetQueryParams(map[string]string{
			"access_token": account.AccessToken,
			"album_id":     splitID(album.Id)[0],
			"tid":          e[2],
			"fsid":         e[0],
			"uk":           e[1],
		}).
		Head(ALBUM_API_URL + "/download")
	if err != nil {
		return nil, err
	}
	return &base.Link{
		Headers: []base.Header{
			{Name: "User-Agent", Value: base.UserAgent},
		},
		Url: res.Header().Get("location"),
	}, nil
}

func (driver Baidu) LinkFile(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}

	if !IsAlbumFile(file) {
		return nil, base.ErrNotSupport
	}

	album, err := driver.File(utils.Dir(utils.ParsePath(args.Path)), account)
	if err != nil {
		return nil, err
	}
	// 拷贝到根目录
	cfile, err := driver.CopyAlbumFile(album.Id, account, file.Id)
	if err != nil {
		return nil, err
	}

	res, err := driver.Request(http.MethodGet, FILE_API_URL_V2+"/download", func(r *resty.Request) {
		r.SetQueryParams(map[string]string{
			"fsid": fmt.Sprint(cfile.Fsid),
		})
	}, account)
	if err != nil {
		return nil, err
	}
	return &base.Link{
		Headers: []base.Header{
			{Name: "User-Agent", Value: base.UserAgent},
		},
		Url: utils.Json.Get(res.Body(), "dlink").ToString(),
	}, nil
}

func (driver Baidu) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
	file, err := driver.File(path, account)
	if err != nil {
		return nil, nil, err
	}
	if !file.IsDir() {
		return file, nil, nil
	}
	files, err := driver.Files(path, account)
	if err != nil {
		return nil, nil, err
	}
	return nil, files, nil
}

func (driver Baidu) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver Baidu) Rename(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	if IsAlbum(srcFile) {
		return driver.SetAlbumName(srcFile.Id, utils.Base(dst), account)
	}
	return base.ErrNotSupport
}

func (driver Baidu) MakeDir(path string, account *model.Account) error {
	dir, name := utils.Split(path)
	parentFile, err := driver.File(dir, account)
	if err != nil {
		return err
	}

	if !IsRoot(parentFile) {
		return base.ErrNotSupport
	}
	return driver.CreateAlbum(name, account)
}

func (driver Baidu) Move(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	if IsAlbumFile(srcFile) {
		// 移动相册文件
		dstAlbum, err := driver.File(utils.Dir(dst), account)
		if err != nil {
			return err
		}
		if !IsAlbum(dstAlbum) {
			return base.ErrNotSupport
		}

		srcAlbum, err := driver.File(utils.Dir(src), account)
		if err != nil {
			return err
		}

		newFile, err := driver.CopyAlbumFile(srcAlbum.Id, account, srcFile.Id)
		if err != nil {
			return err
		}
		err = driver.DeleteAlbumFile(srcAlbum.Id, account, srcFile.Id)
		if err != nil {
			return err
		}
		err = driver.AddAlbumFile(dstAlbum.Id, account, joinID(newFile.Fsid))
		if err != nil {
			return err
		}
		return nil
	}
	return base.ErrNotSupport
}

func (driver Baidu) Copy(src string, dst string, account *model.Account) error {
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}

	if IsAlbumFile(srcFile) {
		// 复制相册文件
		dstAlbum, err := driver.File(utils.Dir(dst), account)
		if err != nil {
			return err
		}
		if !IsAlbum(dstAlbum) {
			return base.ErrNotSupport
		}

		srcAlbum, err := driver.File(utils.Dir(src), account)
		if err != nil {
			return err
		}

		newFile, err := driver.CopyAlbumFile(srcAlbum.Id, account, srcFile.Id)
		if err != nil {
			return err
		}
		err = driver.AddAlbumFile(dstAlbum.Id, account, joinID(newFile.Fsid))
		if err != nil {
			return err
		}
		return nil
	}
	return base.ErrNotSupport
}

func (driver Baidu) Delete(path string, account *model.Account) error {
	file, err := driver.File(path, account)
	if err != nil {
		return err
	}

	// 删除相册
	if IsAlbum(file) {
		return driver.DeleteAlbum(file.Id, account)
	}

	// 生成相册文件
	if IsAlbumFile(file) {
		// 删除相册文件
		album, err := driver.File(utils.Dir(path), account)
		if err != nil {
			return err
		}
		return driver.DeleteAlbumFile(album.Id, account, file.Id)
	}
	return base.ErrNotSupport
}

func (driver Baidu) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}

	parentFile, err := driver.File(file.ParentPath, account)
	if err != nil {
		return err
	}

	if !IsAlbum(parentFile) {
		return base.ErrNotSupport
	}

	tempFile, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
	if err != nil {
		return err
	}
	defer func() {
		tempFile.Close()
		os.Remove(tempFile.Name())
	}()

	// 计算需要的数据
	const DEFAULT = 1 << 22
	const SliceSize = 1 << 18
	count := int(math.Ceil(float64(file.Size) / float64(DEFAULT)))

	sliceMD5List := make([]string, 0, count)
	fileMd5 := md5.New()
	sliceMd5 := md5.New()
	for i := 1; i <= count; i++ {
		if n, err := io.CopyN(io.MultiWriter(fileMd5, sliceMd5, tempFile), file, DEFAULT); err != io.EOF && n == 0 {
			return err
		}
		sliceMD5List = append(sliceMD5List, hex.EncodeToString(sliceMd5.Sum(nil)))
		sliceMd5.Reset()
	}

	if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
		return err
	}

	content_md5 := hex.EncodeToString(fileMd5.Sum(nil))
	slice_md5 := content_md5
	if file.GetSize() > SliceSize {
		sliceData := make([]byte, SliceSize)
		if _, err = io.ReadFull(tempFile, sliceData); err != nil {
			return err
		}
		sliceMd5.Write(sliceData)
		slice_md5 = hex.EncodeToString(sliceMd5.Sum(nil))
		if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
			return err
		}
	}

	// 开始执行上传
	params := map[string]string{
		"autoinit":    "1",
		"isdir":       "0",
		"rtype":       "1",
		"ctype":       "11",
		"path":        utils.ParsePath(file.Name),
		"size":        fmt.Sprint(file.Size),
		"slice-md5":   slice_md5,
		"content-md5": content_md5,
		"block_list":  MustString(utils.Json.MarshalToString(sliceMD5List)),
	}

	// 预上传
	var precreateResp PrecreateResp
	_, err = driver.Request(http.MethodPost, FILE_API_URL_V1+"/precreate", func(r *resty.Request) {
		r.SetFormData(params)
		r.SetResult(&precreateResp)
	}, account)
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
			_, err = driver.Request(http.MethodPost, "https://c3.pcs.baidu.com/rest/2.0/pcs/superfile2", func(r *resty.Request) {
				r.SetQueryParams(uploadParams)
				r.SetFileReader("file", file.Name, io.LimitReader(tempFile, DEFAULT))
			}, account)
			if err != nil {
				return err
			}
		}
		fallthrough
	case 2: // 创建文件
		params["uploadid"] = precreateResp.UploadID
		_, err = driver.Request(http.MethodPost, FILE_API_URL_V1+"/create", func(r *resty.Request) {
			r.SetFormData(params)
			r.SetResult(&precreateResp)
		}, account)
		if err != nil {
			return err
		}
		fallthrough
	case 3: // 增加到相册
		err = driver.AddAlbumFile(parentFile.Id, account, joinID(precreateResp.Data.FsID))
		if err != nil {
			return err
		}
	}
	return nil
}

var _ base.Driver = (*Baidu)(nil)

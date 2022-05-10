package baiduphoto

import (
	"fmt"
	"net/http"

	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

func (driver Baidu) RefreshToken(account *model.Account) error {
	err := driver.refreshToken(account)
	if err != nil && err == base.ErrEmptyToken {
		err = driver.refreshToken(account)
	}
	if err != nil {
		account.Status = err.Error()
	}
	_ = model.SaveAccount(account)
	return err
}

func (driver Baidu) refreshToken(account *model.Account) error {
	u := "https://openapi.baidu.com/oauth/2.0/token"
	var resp base.TokenResp
	var e TokenErrResp
	_, err := base.RestyClient.R().
		SetResult(&resp).
		SetError(&e).
		SetQueryParams(map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": account.RefreshToken,
			"client_id":     account.ClientId,
			"client_secret": account.ClientSecret,
		}).Get(u)
	if err != nil {
		return err
	}
	if e.ErrorMsg != "" {
		return &e
	}
	if resp.RefreshToken == "" {
		return base.ErrEmptyToken
	}
	account.Status = "work"
	account.AccessToken, account.RefreshToken = resp.AccessToken, resp.RefreshToken
	return nil
}

func (driver Baidu) Request(method string, url string, callback func(*resty.Request), account *model.Account) (*resty.Response, error) {
	req := base.RestyClient.R()
	req.SetQueryParam("access_token", account.AccessToken)
	if callback != nil {
		callback(req)
	}

	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	log.Debug(res.String())

	var erron Erron
	if err = utils.Json.Unmarshal(res.Body(), &erron); err != nil {
		return nil, err
	}

	switch erron.Errno {
	case 0:
		return res, nil
	case -6:
		if err = driver.RefreshToken(account); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("errno: %d, refer to https://photo.baidu.com/union/doc", erron.Errno)
	}
	return driver.Request(method, url, callback, account)
}

// 获取所有根文件
func (driver Baidu) GetAllFile(account *model.Account) (files []File, err error) {
	var cursor string

	for {
		var resp FileListResp
		_, err = driver.Request(http.MethodGet, FILE_API_URL_V1+"/list", func(r *resty.Request) {
			r.SetQueryParams(map[string]string{
				"need_thumbnail":     "1",
				"need_filter_hidden": "0",
				"cursor":             cursor,
			})
			r.SetResult(&resp)
		}, account)
		if err != nil {
			return
		}

		cursor = resp.Cursor
		files = append(files, resp.List...)

		if !resp.HasNextPage() {
			return
		}
	}
}

// 获取所有相册
func (driver Baidu) GetAllAlbum(account *model.Account) (albums []Album, err error) {
	var cursor string
	for {
		var resp AlbumListResp
		_, err = driver.Request(http.MethodGet, ALBUM_API_URL+"/list", func(r *resty.Request) {
			r.SetQueryParams(map[string]string{
				"need_amount": "1",
				"limit":       "100",
				"cursor":      cursor,
			})
			r.SetResult(&resp)
		}, account)
		if err != nil {
			return
		}
		if albums == nil {
			albums = make([]Album, 0, resp.TotalCount)
		}

		cursor = resp.Cursor
		albums = append(albums, resp.List...)

		if !resp.HasNextPage() {
			return
		}
	}
}

// 获取相册中所有文件
func (driver Baidu) GetAllAlbumFile(albumID string, account *model.Account) (files []AlbumFile, err error) {
	var cursor string
	for {
		var resp AlbumFileListResp
		_, err = driver.Request(http.MethodGet, ALBUM_API_URL+"/listfile", func(r *resty.Request) {
			r.SetQueryParams(map[string]string{
				"album_id":    splitID(albumID)[0],
				"need_amount": "1",
				"limit":       "1000",
				"cursor":      cursor,
			})
			r.SetResult(&resp)
		}, account)
		if err != nil {
			return
		}
		if files == nil {
			files = make([]AlbumFile, 0, resp.TotalCount)
		}

		cursor = resp.Cursor
		files = append(files, resp.List...)

		if !resp.HasNextPage() {
			return
		}
	}
}

// 创建相册
func (driver Baidu) CreateAlbum(name string, account *model.Account) error {
	if !checkName(name) {
		return ErrNotSupportName
	}
	_, err := driver.Request(http.MethodPost, ALBUM_API_URL+"/create", func(r *resty.Request) {
		r.SetQueryParams(map[string]string{
			"title":  name,
			"tid":    getTid(),
			"source": "0",
		})
	}, account)
	return err
}

// 相册改名
func (driver Baidu) SetAlbumName(albumID string, name string, account *model.Account) error {
	if !checkName(name) {
		return ErrNotSupportName
	}

	e := splitID(albumID)
	_, err := driver.Request(http.MethodPost, ALBUM_API_URL+"/settitle", func(r *resty.Request) {
		r.SetFormData(map[string]string{
			"title":    name,
			"album_id": e[0],
			"tid":      e[1],
		})
	}, account)
	return err
}

// 删除相册
func (driver Baidu) DeleteAlbum(albumID string, account *model.Account) error {
	e := splitID(albumID)
	_, err := driver.Request(http.MethodPost, ALBUM_API_URL+"/delete", func(r *resty.Request) {
		r.SetFormData(map[string]string{
			"album_id":            e[0],
			"tid":                 e[1],
			"delete_origin_image": "0", // 是否删除原图 0 不删除
		})
	}, account)
	return err
}

// 删除相册文件
func (driver Baidu) DeleteAlbumFile(albumID string, account *model.Account, fileIDs ...string) error {
	e := splitID(albumID)
	_, err := driver.Request(http.MethodPost, ALBUM_API_URL+"/delfile", func(r *resty.Request) {
		r.SetFormData(map[string]string{
			"album_id":   e[0],
			"tid":        e[1],
			"list":       fsidsFormat(fileIDs...),
			"del_origin": "0", // 是否删除原图 0 不删除 1 删除
		})
	}, account)
	return err
}

// 增加相册文件
func (driver Baidu) AddAlbumFile(albumID string, account *model.Account, fileIDs ...string) error {
	e := splitID(albumID)
	_, err := driver.Request(http.MethodGet, ALBUM_API_URL+"/addfile", func(r *resty.Request) {
		r.SetQueryParams(map[string]string{
			"album_id": e[0],
			"tid":      e[1],
			"list":     fsidsFormatNotUk(fileIDs...),
		})
	}, account)
	return err
}

// 保存相册文件为根文件
func (driver Baidu) CopyAlbumFile(albumID string, account *model.Account, fileID string) (*CopyFile, error) {
	var resp CopyFileResp
	e := splitID(fileID)
	_, err := driver.Request(http.MethodPost, ALBUM_API_URL+"/copyfile", func(r *resty.Request) {
		r.SetFormData(map[string]string{
			"album_id": splitID(albumID)[0],
			"tid":      e[2],
			"uk":       e[1],
			"list":     fsidsFormatNotUk(fileID),
		})
		r.SetResult(&resp)
	}, account)
	if err != nil {
		return nil, err
	}
	return &resp.List[0], err
}

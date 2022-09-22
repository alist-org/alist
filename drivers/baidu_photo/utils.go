package baiduphoto

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
)

const (
	API_URL         = "https://photo.baidu.com/youai"
	ALBUM_API_URL   = API_URL + "/album/v1"
	FILE_API_URL_V1 = API_URL + "/file/v1"
	FILE_API_URL_V2 = API_URL + "/file/v2"
)

var (
	ErrNotSupportName = errors.New("only chinese and english, numbers and underscores are supported, and the length is no more than 20")
)

func (p *BaiduPhoto) Request(furl string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	req := base.RestyClient.R().
		SetQueryParam("access_token", p.AccessToken)
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	res, err := req.Execute(method, furl)
	if err != nil {
		return nil, err
	}

	erron := utils.Json.Get(res.Body(), "errno").ToInt()
	switch erron {
	case 0:
		break
	case 50805:
		return nil, fmt.Errorf("you have joined album")
	case 50820:
		return nil, fmt.Errorf("no shared albums found")
	case -6:
		if err = p.refreshToken(); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("errno: %d, refer to https://photo.baidu.com/union/doc", erron)
	}
	return res.Body(), nil
}

func (p *BaiduPhoto) refreshToken() error {
	u := "https://openapi.baidu.com/oauth/2.0/token"
	var resp base.TokenResp
	var e TokenErrResp
	_, err := base.RestyClient.R().SetResult(&resp).SetError(&e).SetQueryParams(map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": p.RefreshToken,
		"client_id":     p.ClientID,
		"client_secret": p.ClientSecret,
	}).Get(u)
	if err != nil {
		return err
	}
	if e.ErrorMsg != "" {
		return &e
	}
	if resp.RefreshToken == "" {
		return errs.EmptyToken
	}
	p.AccessToken, p.RefreshToken = resp.AccessToken, resp.RefreshToken
	op.MustSaveDriverStorage(p)
	return nil
}

func (p *BaiduPhoto) Get(furl string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	return p.Request(furl, http.MethodGet, callback, resp)
}

func (p *BaiduPhoto) Post(furl string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	return p.Request(furl, http.MethodPost, callback, resp)
}

// 获取所有文件
func (p *BaiduPhoto) GetAllFile(ctx context.Context) (files []File, err error) {
	var cursor string
	for {
		var resp FileListResp
		_, err = p.Get(FILE_API_URL_V1+"/list", func(r *resty.Request) {
			r.SetContext(ctx)
			r.SetQueryParams(map[string]string{
				"need_thumbnail":     "1",
				"need_filter_hidden": "0",
				"cursor":             cursor,
			})
		}, &resp)
		if err != nil {
			return
		}

		files = append(files, resp.List...)
		if !resp.HasNextPage() {
			return
		}
		cursor = resp.Cursor
	}
}

// 删除根文件
func (p *BaiduPhoto) DeleteFile(ctx context.Context, fileIDs ...string) error {
	_, err := p.Get(FILE_API_URL_V1+"/delete", func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetQueryParams(map[string]string{
			"fsid_list": fmt.Sprintf("[%s]", strings.Join(fileIDs, ",")),
		})
	}, nil)
	return err
}

// 获取所有相册
func (p *BaiduPhoto) GetAllAlbum(ctx context.Context) (albums []Album, err error) {
	var cursor string
	for {
		var resp AlbumListResp
		_, err = p.Get(ALBUM_API_URL+"/list", func(r *resty.Request) {
			r.SetContext(ctx)
			r.SetQueryParams(map[string]string{
				"need_amount": "1",
				"limit":       "100",
				"cursor":      cursor,
			})
		}, &resp)
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
func (p *BaiduPhoto) GetAllAlbumFile(ctx context.Context, albumID, passwd string) (files []AlbumFile, err error) {
	var cursor string
	for {
		var resp AlbumFileListResp
		_, err = p.Get(ALBUM_API_URL+"/listfile", func(r *resty.Request) {
			r.SetContext(ctx)
			r.SetQueryParams(map[string]string{
				"album_id":    albumID,
				"need_amount": "1",
				"limit":       "1000",
				"passwd":      passwd,
				"cursor":      cursor,
			})
		}, &resp)
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
func (p *BaiduPhoto) CreateAlbum(ctx context.Context, name string) error {
	if !checkName(name) {
		return ErrNotSupportName
	}
	_, err := p.Post(ALBUM_API_URL+"/create", func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetQueryParams(map[string]string{
			"title":  name,
			"tid":    getTid(),
			"source": "0",
		})
	}, nil)
	return err
}

// 相册改名
func (p *BaiduPhoto) SetAlbumName(ctx context.Context, albumID, tID, name string) error {
	if !checkName(name) {
		return ErrNotSupportName
	}

	_, err := p.Post(ALBUM_API_URL+"/settitle", func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetFormData(map[string]string{
			"title":    name,
			"album_id": albumID,
			"tid":      tID,
		})
	}, nil)
	return err
}

// 删除相册
func (p *BaiduPhoto) DeleteAlbum(ctx context.Context, albumID, tID string) error {
	_, err := p.Post(ALBUM_API_URL+"/delete", func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetFormData(map[string]string{
			"album_id":            albumID,
			"tid":                 tID,
			"delete_origin_image": "0", // 是否删除原图 0 不删除 1 删除
		})
	}, nil)
	return err
}

// 删除相册文件
func (p *BaiduPhoto) DeleteAlbumFile(ctx context.Context, albumID, tID string, fileIDs ...string) error {
	_, err := p.Post(ALBUM_API_URL+"/delfile", func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetFormData(map[string]string{
			"album_id":   albumID,
			"tid":        tID,
			"list":       fsidsFormat(fileIDs...),
			"del_origin": "0", // 是否删除原图 0 不删除 1 删除
		})
	}, nil)
	return err
}

// 增加相册文件
func (p *BaiduPhoto) AddAlbumFile(ctx context.Context, albumID, tID string, fileIDs ...string) error {
	_, err := p.Get(ALBUM_API_URL+"/addfile", func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetQueryParams(map[string]string{
			"album_id": albumID,
			"tid":      tID,
			"list":     fsidsFormatNotUk(fileIDs...),
		})
	}, nil)
	return err
}

// 保存相册文件为根文件
func (p *BaiduPhoto) CopyAlbumFile(ctx context.Context, albumID, tID, uk string, fileID ...string) (*CopyFile, error) {
	var resp CopyFileResp
	_, err := p.Post(ALBUM_API_URL+"/copyfile", func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetFormData(map[string]string{
			"album_id": albumID,
			"tid":      tID,
			"uk":       uk,
			"list":     fsidsFormatNotUk(fileID...),
		})
		r.SetResult(&resp)
	}, nil)
	if err != nil {
		return nil, err
	}
	return &resp.List[0], nil
}

// 加入相册
func (p *BaiduPhoto) JoinAlbum(ctx context.Context, code string) error {
	var resp InviteResp
	_, err := p.Get(ALBUM_API_URL+"/querypcode", func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetQueryParams(map[string]string{
			"pcode": code,
			"web":   "1",
		})
	}, &resp)
	if err != nil {
		return err
	}
	_, err = p.Get(ALBUM_API_URL+"/join", func(req *resty.Request) {
		req.SetContext(ctx)
		req.SetQueryParams(map[string]string{
			"invite_code": resp.Pdata.InviteCode,
		})
	}, nil)
	return err
}

func (d *BaiduPhoto) linkAlbum(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	headers := map[string]string{
		"User-Agent": base.UserAgent,
	}
	if args.Header.Get("User-Agent") != "" {
		headers["User-Agent"] = args.Header.Get("User-Agent")
	}
	if !utils.IsLocalIPAddr(args.IP) {
		headers["X-Forwarded-For"] = args.IP
	}

	e := splitID(file.GetID())
	res, err := base.NoRedirectClient.R().
		SetContext(ctx).
		SetHeaders(headers).
		SetQueryParams(map[string]string{
			"access_token": d.AccessToken,
			"fsid":         e[0],
			"album_id":     e[1],
			"tid":          e[2],
			"uk":           e[3],
		}).
		Head(ALBUM_API_URL + "/download")

	if err != nil {
		return nil, err
	}

	//exp := 8 * time.Hour
	link := &model.Link{
		URL: res.Header().Get("location"),
		Header: http.Header{
			"User-Agent": []string{headers["User-Agent"]},
			"Referer":    []string{"https://photo.baidu.com/"},
		},
		//Expiration: &exp,
	}
	return link, nil
}

func (d *BaiduPhoto) linkFile(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	headers := map[string]string{
		"User-Agent": base.UserAgent,
	}
	if args.Header.Get("User-Agent") != "" {
		headers["User-Agent"] = args.Header.Get("User-Agent")
	}
	if !utils.IsLocalIPAddr(args.IP) {
		headers["X-Forwarded-For"] = args.IP
	}

	var downloadUrl struct {
		Dlink string `json:"dlink"`
	}
	_, err := d.Get(FILE_API_URL_V2+"/download", func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetHeaders(headers)
		r.SetQueryParams(map[string]string{
			"fsid": splitID(file.GetID())[0],
		})
	}, &downloadUrl)
	if err != nil {
		return nil, err
	}

	//exp := 8 * time.Hour
	link := &model.Link{
		URL: downloadUrl.Dlink,
		Header: http.Header{
			"User-Agent": []string{headers["User-Agent"]},
			"Referer":    []string{"https://photo.baidu.com/"},
		},
		//Expiration: &exp,
	}
	return link, nil
}

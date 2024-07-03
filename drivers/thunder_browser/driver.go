package thunder_browser

import (
	"context"
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	hash_extend "github.com/alist-org/alist/v3/pkg/utils/hash"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/go-resty/resty/v2"
	"net/http"
	"regexp"
	"strings"
)

type ThunderBrowser struct {
	*XunLeiBrowserCommon
	model.Storage
	Addition

	identity string
}

func (x *ThunderBrowser) Config() driver.Config {
	return config
}

func (x *ThunderBrowser) GetAddition() driver.Additional {
	return &x.Addition
}

func (x *ThunderBrowser) Init(ctx context.Context) (err error) {

	spaceTokenFunc := func() error {
		// 如果用户未设置 "超级保险柜" 密码 则直接返回
		if x.SafePassword == "" {
			return nil
		}
		// 通过 GetSafeAccessToken 获取
		token, err := x.GetSafeAccessToken(x.SafePassword)
		x.SetSpaceTokenResp(token)
		return err
	}

	// 初始化所需参数
	if x.XunLeiBrowserCommon == nil {
		x.XunLeiBrowserCommon = &XunLeiBrowserCommon{
			Common: &Common{
				client:            base.NewRestyClient(),
				Algorithms:        Algorithms,
				DeviceID:          utils.GetMD5EncodeStr(x.Username + x.Password),
				ClientID:          ClientID,
				ClientSecret:      ClientSecret,
				ClientVersion:     ClientVersion,
				PackageName:       PackageName,
				UserAgent:         BuildCustomUserAgent(utils.GetMD5EncodeStr(x.Username+x.Password), PackageName, SdkVersion, ClientVersion, PackageName),
				DownloadUserAgent: DownloadUserAgent,
				UseVideoUrl:       x.UseVideoUrl,
				RemoveWay:         x.Addition.RemoveWay,
				refreshCTokenCk: func(token string) {
					x.CaptchaToken = token
					op.MustSaveDriverStorage(x)
				},
			},
			refreshTokenFunc: func() error {
				// 通过RefreshToken刷新
				token, err := x.RefreshToken(x.TokenResp.RefreshToken)
				if err != nil {
					// 重新登录
					token, err = x.Login(x.Username, x.Password)
					if err != nil {
						x.GetStorage().SetStatus(fmt.Sprintf("%+v", err.Error()))
						op.MustSaveDriverStorage(x)
					}
				}
				x.SetTokenResp(token)
				return err
			},
		}
	}

	// 自定义验证码token
	ctoekn := strings.TrimSpace(x.CaptchaToken)
	if ctoekn != "" {
		x.SetCaptchaToken(ctoekn)
	}
	if x.DeviceID == "" {
		x.SetDeviceID(utils.GetMD5EncodeStr(x.Username + x.Password))
	}
	x.XunLeiBrowserCommon.UseVideoUrl = x.UseVideoUrl
	x.Addition.RootFolderID = x.RootFolderID
	// 防止重复登录
	identity := x.GetIdentity()
	if x.identity != identity || !x.IsLogin() {
		x.identity = identity
		// 登录
		token, err := x.Login(x.Username, x.Password)
		if err != nil {
			return err
		}
		x.SetTokenResp(token)
	}

	// 获取 spaceToken
	err = spaceTokenFunc()
	if err != nil {
		return err
	}

	return nil
}

func (x *ThunderBrowser) Drop(ctx context.Context) error {
	return nil
}

type ThunderBrowserExpert struct {
	*XunLeiBrowserCommon
	model.Storage
	ExpertAddition

	identity string
}

func (x *ThunderBrowserExpert) Config() driver.Config {
	return configExpert
}

func (x *ThunderBrowserExpert) GetAddition() driver.Additional {
	return &x.ExpertAddition
}

func (x *ThunderBrowserExpert) Init(ctx context.Context) (err error) {

	spaceTokenFunc := func() error {
		// 如果用户未设置 "超级保险柜" 密码 则直接返回
		if x.SafePassword == "" {
			return nil
		}
		// 通过 GetSafeAccessToken 获取
		token, err := x.GetSafeAccessToken(x.SafePassword)
		x.SetSpaceTokenResp(token)
		return err
	}

	// 防止重复登录
	identity := x.GetIdentity()
	if identity != x.identity || !x.IsLogin() {
		x.identity = identity
		x.XunLeiBrowserCommon = &XunLeiBrowserCommon{
			Common: &Common{
				client: base.NewRestyClient(),
				DeviceID: func() string {
					if len(x.DeviceID) != 32 {
						if x.LoginType == "user" {
							return utils.GetMD5EncodeStr(x.Username + x.Password)
						}
						return utils.GetMD5EncodeStr(x.ExpertAddition.RefreshToken)
					}
					return x.DeviceID
				}(),
				ClientID:      x.ClientID,
				ClientSecret:  x.ClientSecret,
				ClientVersion: x.ClientVersion,
				PackageName:   x.PackageName,
				UserAgent: func() string {
					if x.ExpertAddition.UserAgent != "" {
						return x.ExpertAddition.UserAgent
					}
					if x.LoginType == "user" {
						return BuildCustomUserAgent(utils.GetMD5EncodeStr(x.Username+x.Password), x.PackageName, SdkVersion, x.ClientVersion, x.PackageName)
					}
					return BuildCustomUserAgent(utils.GetMD5EncodeStr(x.ExpertAddition.RefreshToken), x.PackageName, SdkVersion, x.ClientVersion, x.PackageName)
				}(),
				DownloadUserAgent: func() string {
					if x.ExpertAddition.DownloadUserAgent != "" {
						return x.ExpertAddition.DownloadUserAgent
					}
					return DownloadUserAgent
				}(),
				UseVideoUrl: x.UseVideoUrl,
				RemoveWay:   x.ExpertAddition.RemoveWay,
				refreshCTokenCk: func(token string) {
					x.CaptchaToken = token
					op.MustSaveDriverStorage(x)
				},
			},
		}

		if x.ExpertAddition.CaptchaToken != "" {
			x.SetCaptchaToken(x.ExpertAddition.CaptchaToken)
			op.MustSaveDriverStorage(x)
		}
		if x.Common.DeviceID != "" {
			x.ExpertAddition.DeviceID = x.Common.DeviceID
			op.MustSaveDriverStorage(x)
		}
		if x.Common.UserAgent != "" {
			x.ExpertAddition.UserAgent = x.Common.UserAgent
			op.MustSaveDriverStorage(x)
		}
		if x.Common.DownloadUserAgent != "" {
			x.ExpertAddition.DownloadUserAgent = x.Common.DownloadUserAgent
			op.MustSaveDriverStorage(x)
		}
		x.XunLeiBrowserCommon.UseVideoUrl = x.UseVideoUrl
		x.ExpertAddition.RootFolderID = x.RootFolderID
		// 签名方法
		if x.SignType == "captcha_sign" {
			x.Common.Timestamp = x.Timestamp
			x.Common.CaptchaSign = x.CaptchaSign
		} else {
			x.Common.Algorithms = strings.Split(x.Algorithms, ",")
		}

		// 登录方式
		if x.LoginType == "refresh_token" {
			// 通过RefreshToken登录
			token, err := x.XunLeiBrowserCommon.RefreshToken(x.ExpertAddition.RefreshToken)
			if err != nil {
				return err
			}
			x.SetTokenResp(token)

			// 刷新token方法
			x.SetRefreshTokenFunc(func() error {
				token, err := x.XunLeiBrowserCommon.RefreshToken(x.TokenResp.RefreshToken)
				if err != nil {
					x.GetStorage().SetStatus(fmt.Sprintf("%+v", err.Error()))
				}
				x.SetTokenResp(token)
				op.MustSaveDriverStorage(x)
				return err
			})

			err = spaceTokenFunc()
			if err != nil {
				return err
			}

		} else {
			// 通过用户密码登录
			token, err := x.Login(x.Username, x.Password)
			if err != nil {
				return err
			}
			x.SetTokenResp(token)
			x.SetRefreshTokenFunc(func() error {
				token, err := x.XunLeiBrowserCommon.RefreshToken(x.TokenResp.RefreshToken)
				if err != nil {
					token, err = x.Login(x.Username, x.Password)
					if err != nil {
						x.GetStorage().SetStatus(fmt.Sprintf("%+v", err.Error()))
					}
				}
				x.SetTokenResp(token)
				op.MustSaveDriverStorage(x)
				return err
			})

			err = spaceTokenFunc()
			if err != nil {
				return err
			}
		}
	} else {
		// 仅修改验证码token
		if x.CaptchaToken != "" {
			x.SetCaptchaToken(x.CaptchaToken)
		}

		err = spaceTokenFunc()
		if err != nil {
			return err
		}

		x.XunLeiBrowserCommon.UserAgent = x.UserAgent
		x.XunLeiBrowserCommon.DownloadUserAgent = x.DownloadUserAgent
		x.XunLeiBrowserCommon.UseVideoUrl = x.UseVideoUrl
		x.ExpertAddition.RootFolderID = x.RootFolderID
	}

	return nil
}

func (x *ThunderBrowserExpert) Drop(ctx context.Context) error {
	return nil
}

func (x *ThunderBrowserExpert) SetTokenResp(token *TokenResp) {
	x.XunLeiBrowserCommon.SetTokenResp(token)
	if token != nil {
		x.ExpertAddition.RefreshToken = token.RefreshToken
	}
}

type XunLeiBrowserCommon struct {
	*Common
	*TokenResp // 登录信息

	refreshTokenFunc func() error
}

func (xc *XunLeiBrowserCommon) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	return xc.getFiles(ctx, dir.GetID(), args.ReqPath)
}

func (xc *XunLeiBrowserCommon) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var lFile Files

	params := map[string]string{
		"_magic":         "2021",
		"space":          "SPACE_BROWSER",
		"thumbnail_size": "SIZE_LARGE",
		"with":           "url",
	}
	// 对 "迅雷云盘" 内的文件 特殊处理
	if file.GetPath() == ThunderDriveFileID {
		params = map[string]string{}
	} else if file.GetPath() == ThunderBrowserDriveSafeFileID {
		// 对 "超级保险箱" 内的文件 特殊处理
		params["space"] = "SPACE_BROWSER_SAFE"
	}

	_, err := xc.Request(FILE_API_URL+"/{fileID}", http.MethodGet, func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetPathParam("fileID", file.GetID())
		r.SetQueryParams(params)
		//r.SetQueryParam("space", "")
	}, &lFile)
	if err != nil {
		return nil, err
	}
	link := &model.Link{
		URL: lFile.WebContentLink,
		Header: http.Header{
			"User-Agent": {xc.DownloadUserAgent},
		},
	}

	if xc.UseVideoUrl {
		for _, media := range lFile.Medias {
			if media.Link.URL != "" {
				link.URL = media.Link.URL
				break
			}
		}
	}
	return link, nil
}

func (xc *XunLeiBrowserCommon) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	js := base.Json{
		"kind":      FOLDER,
		"name":      dirName,
		"parent_id": parentDir.GetID(),
		"space":     "SPACE_BROWSER",
	}
	if parentDir.GetPath() == ThunderDriveFileID {
		js = base.Json{
			"kind":      FOLDER,
			"name":      dirName,
			"parent_id": parentDir.GetID(),
		}
	} else if parentDir.GetPath() == ThunderBrowserDriveSafeFileID {
		js = base.Json{
			"kind":      FOLDER,
			"name":      dirName,
			"parent_id": parentDir.GetID(),
			"space":     "SPACE_BROWSER_SAFE",
		}
	}
	_, err := xc.Request(FILE_API_URL, http.MethodPost, func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetBody(&js)
	}, nil)
	return err
}

func (xc *XunLeiBrowserCommon) Move(ctx context.Context, srcObj, dstDir model.Obj) error {

	srcSpace := "SPACE_BROWSER"
	dstSpace := "SPACE_BROWSER"

	// 对 "超级保险箱" 内的文件 特殊处理
	if srcObj.GetPath() == ThunderBrowserDriveSafeFileID {
		srcSpace = "SPACE_BROWSER_SAFE"
	}
	if dstDir.GetPath() == ThunderBrowserDriveSafeFileID {
		dstSpace = "SPACE_BROWSER_SAFE"
	}

	params := map[string]string{
		"_from": dstSpace,
	}
	js := base.Json{
		"to":    base.Json{"parent_id": dstDir.GetID(), "space": dstSpace},
		"space": srcSpace,
		"ids":   []string{srcObj.GetID()},
	}
	// 对 "迅雷云盘" 内的文件 特殊处理
	if srcObj.GetPath() == ThunderDriveFileID {
		params = map[string]string{}
		js = base.Json{
			"to":  base.Json{"parent_id": dstDir.GetID()},
			"ids": []string{srcObj.GetID()},
		}
	}

	_, err := xc.Request(FILE_API_URL+":batchMove", http.MethodPost, func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetBody(&js)
		r.SetQueryParams(params)
	}, nil)
	return err
}

func (xc *XunLeiBrowserCommon) Rename(ctx context.Context, srcObj model.Obj, newName string) error {

	params := map[string]string{
		"space": "SPACE_BROWSER",
	}
	// 对 "迅雷云盘" 内的文件 特殊处理
	if srcObj.GetPath() == ThunderDriveFileID {
		params = map[string]string{}
	} else if srcObj.GetPath() == ThunderBrowserDriveSafeFileID {
		// 对 "超级保险箱" 内的文件 特殊处理
		params = map[string]string{
			"space": "SPACE_BROWSER_SAFE",
		}
	}

	_, err := xc.Request(FILE_API_URL+"/{fileID}", http.MethodPatch, func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetPathParam("fileID", srcObj.GetID())
		r.SetBody(&base.Json{"name": newName})
		r.SetQueryParams(params)
	}, nil)
	return err
}

func (xc *XunLeiBrowserCommon) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {

	srcSpace := "SPACE_BROWSER"
	dstSpace := "SPACE_BROWSER"

	// 对 "超级保险箱" 内的文件 特殊处理
	if srcObj.GetPath() == ThunderBrowserDriveSafeFileID {
		srcSpace = "SPACE_BROWSER_SAFE"
	}
	if dstDir.GetPath() == ThunderBrowserDriveSafeFileID {
		dstSpace = "SPACE_BROWSER_SAFE"
	}

	params := map[string]string{
		"_from": dstSpace,
	}
	js := base.Json{
		"to":    base.Json{"parent_id": dstDir.GetID(), "space": dstSpace},
		"space": srcSpace,
		"ids":   []string{srcObj.GetID()},
	}
	// 对 "迅雷云盘" 内的文件 特殊处理
	if srcObj.GetPath() == ThunderDriveFileID {
		params = map[string]string{}
		js = base.Json{
			"to":  base.Json{"parent_id": dstDir.GetID()},
			"ids": []string{srcObj.GetID()},
		}
	}

	_, err := xc.Request(FILE_API_URL+":batchCopy", http.MethodPost, func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetBody(&js)
		r.SetQueryParams(params)
	}, nil)
	return err
}

func (xc *XunLeiBrowserCommon) Remove(ctx context.Context, obj model.Obj) error {

	js := base.Json{
		"ids":   []string{obj.GetID()},
		"space": "SPACE_BROWSER",
	}
	// 对 "迅雷云盘" 内的文件 特殊处理
	if obj.GetPath() == ThunderDriveFileID {
		js = base.Json{
			"ids": []string{obj.GetID()},
		}
	} else if obj.GetPath() == ThunderBrowserDriveSafeFileID {
		// 对 "超级保险箱" 内的文件 特殊处理
		js = base.Json{
			"ids":   []string{obj.GetID()},
			"space": "SPACE_BROWSER_SAFE",
		}
	}

	// 先判断是否是特殊情况
	if obj.GetPath() == ThunderDriveFileID {
		_, err := xc.Request(FILE_API_URL+"/{fileID}/trash", http.MethodPatch, func(r *resty.Request) {
			r.SetContext(ctx)
			r.SetPathParam("fileID", obj.GetID())
			r.SetBody("{}")
		}, nil)
		return err
	} else if obj.GetPath() == ThunderBrowserDriveSafeFileID {
		_, err := xc.Request(FILE_API_URL+":batchDelete", http.MethodPost, func(r *resty.Request) {
			r.SetContext(ctx)
			r.SetBody(&js)
		}, nil)
		return err
	}

	// 根据用户选择的删除方式进行删除
	if xc.RemoveWay == "delete" {
		_, err := xc.Request(FILE_API_URL+":batchDelete", http.MethodPost, func(r *resty.Request) {
			r.SetContext(ctx)
			r.SetBody(&js)
		}, nil)
		return err
	} else {
		_, err := xc.Request(FILE_API_URL+":batchTrash", http.MethodPost, func(r *resty.Request) {
			r.SetContext(ctx)
			r.SetBody(&js)
		}, nil)
		return err
	}
}

func (xc *XunLeiBrowserCommon) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	hi := stream.GetHash()
	gcid := hi.GetHash(hash_extend.GCID)
	if len(gcid) < hash_extend.GCID.Width {
		tFile, err := stream.CacheFullInTempFile()
		if err != nil {
			return err
		}

		gcid, err = utils.HashFile(hash_extend.GCID, tFile, stream.GetSize())
		if err != nil {
			return err
		}
	}

	js := base.Json{
		"kind":        FILE,
		"parent_id":   dstDir.GetID(),
		"name":        stream.GetName(),
		"size":        stream.GetSize(),
		"hash":        gcid,
		"upload_type": UPLOAD_TYPE_RESUMABLE,
		"space":       "SPACE_BROWSER",
	}
	// 对 "迅雷云盘" 内的文件 特殊处理
	if dstDir.GetPath() == ThunderDriveFileID {
		js = base.Json{
			"kind":        FILE,
			"parent_id":   dstDir.GetID(),
			"name":        stream.GetName(),
			"size":        stream.GetSize(),
			"hash":        gcid,
			"upload_type": UPLOAD_TYPE_RESUMABLE,
		}
	} else if dstDir.GetPath() == ThunderBrowserDriveSafeFileID {
		// 对 "超级保险箱" 内的文件 特殊处理
		js = base.Json{
			"kind":        FILE,
			"parent_id":   dstDir.GetID(),
			"name":        stream.GetName(),
			"size":        stream.GetSize(),
			"hash":        gcid,
			"upload_type": UPLOAD_TYPE_RESUMABLE,
			"space":       "SPACE_BROWSER_SAFE",
		}
	}

	var resp UploadTaskResponse
	_, err := xc.Request(FILE_API_URL, http.MethodPost, func(r *resty.Request) {
		r.SetContext(ctx)
		r.SetBody(&js)
	}, &resp)
	if err != nil {
		return err
	}

	param := resp.Resumable.Params
	if resp.UploadType == UPLOAD_TYPE_RESUMABLE {
		param.Endpoint = strings.TrimLeft(param.Endpoint, param.Bucket+".")
		s, err := session.NewSession(&aws.Config{
			Credentials: credentials.NewStaticCredentials(param.AccessKeyID, param.AccessKeySecret, param.SecurityToken),
			Region:      aws.String("xunlei"),
			Endpoint:    aws.String(param.Endpoint),
		})
		if err != nil {
			return err
		}
		uploader := s3manager.NewUploader(s)
		if stream.GetSize() > s3manager.MaxUploadParts*s3manager.DefaultUploadPartSize {
			uploader.PartSize = stream.GetSize() / (s3manager.MaxUploadParts - 1)
		}
		_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
			Bucket:  aws.String(param.Bucket),
			Key:     aws.String(param.Key),
			Expires: aws.Time(param.Expiration),
			Body:    stream,
		})
		return err
	}
	return nil
}

func (xc *XunLeiBrowserCommon) getFiles(ctx context.Context, folderId string, path string) ([]model.Obj, error) {
	files := make([]model.Obj, 0)
	var pageToken string
	for {
		var fileList FileList
		folderSpace := "SPACE_BROWSER"
		params := map[string]string{
			"parent_id":      folderId,
			"page_token":     pageToken,
			"space":          folderSpace,
			"filters":        `{"trashed":{"eq":false}}`,
			"with_audit":     "true",
			"thumbnail_size": "SIZE_LARGE",
		}
		var fileType int8
		// 处理特殊目录 “迅雷云盘” 设置特殊的 params 以便正常访问
		pattern1 := fmt.Sprintf(`^/.*/%s(/.*)?$`, ThunderDriveFolderName)
		thunderDriveMatch, _ := regexp.MatchString(pattern1, path)
		// 处理特殊目录 “超级保险箱” 设置特殊的 params 以便正常访问
		pattern2 := fmt.Sprintf(`^/.*/%s(/.*)?$`, ThunderBrowserDriveSafeFolderName)
		thunderBrowserDriveSafeMatch, _ := regexp.MatchString(pattern2, path)

		// 如果是 "迅雷云盘" 内的
		if folderId == ThunderDriveFileID || thunderDriveMatch {
			params = map[string]string{
				"space":      "",
				"__type":     "drive",
				"refresh":    "true",
				"__sync":     "true",
				"parent_id":  folderId,
				"page_token": pageToken,
				"with_audit": "true",
				"limit":      "100",
				"filters":    `{"phase":{"eq":"PHASE_TYPE_COMPLETE"},"trashed":{"eq":false}}`,
			}
			// 如果不是 "迅雷云盘"的"根目录"
			if folderId == ThunderDriveFileID {
				params["parent_id"] = ""
			}
			fileType = ThunderDriveType
		} else if thunderBrowserDriveSafeMatch {
			// 如果是 "超级保险箱" 内的
			fileType = ThunderBrowserDriveSafeType
			params["space"] = "SPACE_BROWSER_SAFE"
		}

		_, err := xc.Request(FILE_API_URL, http.MethodGet, func(r *resty.Request) {
			r.SetContext(ctx)
			r.SetQueryParams(params)
		}, &fileList)
		if err != nil {
			return nil, err
		}
		// 对文件夹也进行处理
		fileList.FolderType = fileType

		for i := 0; i < len(fileList.Files); i++ {
			file := &fileList.Files[i]
			// 标记 文件夹内的文件
			file.FileType = fileList.FolderType
			// 解决 "迅雷云盘" 重复出现问题————迅雷后端发送错误
			if file.Name == ThunderDriveFolderName && file.ID == "" && file.FolderType == ThunderDriveFolderType && folderId != "" {
				continue
			}
			// 处理特殊目录 “迅雷云盘” 设置特殊的文件夹ID
			if file.Name == ThunderDriveFolderName && file.ID == "" && file.FolderType == ThunderDriveFolderType {
				file.ID = ThunderDriveFileID
			} else if file.Name == ThunderBrowserDriveSafeFolderName && file.FolderType == ThunderBrowserDriveSafeFolderType {
				file.FileType = ThunderBrowserDriveSafeType
			}
			files = append(files, file)
		}

		if fileList.NextPageToken == "" {
			break
		}
		pageToken = fileList.NextPageToken
	}
	return files, nil
}

// SetRefreshTokenFunc 设置刷新Token的方法
func (xc *XunLeiBrowserCommon) SetRefreshTokenFunc(fn func() error) {
	xc.refreshTokenFunc = fn
}

// SetTokenResp 设置Token
func (xc *XunLeiBrowserCommon) SetTokenResp(tr *TokenResp) {
	xc.TokenResp = tr
}

// SetSpaceTokenResp 设置Token
func (xc *XunLeiBrowserCommon) SetSpaceTokenResp(spaceToken string) {
	xc.TokenResp.Token = spaceToken
}

// Request 携带Authorization和CaptchaToken的请求
func (xc *XunLeiBrowserCommon) Request(url string, method string, callback base.ReqCallback, resp interface{}) ([]byte, error) {
	data, err := xc.Common.Request(url, method, func(req *resty.Request) {
		req.SetHeaders(map[string]string{
			"Authorization":         xc.GetToken(),
			"X-Captcha-Token":       xc.GetCaptchaToken(),
			"X-Space-Authorization": xc.GetSpaceToken(),
		})
		if callback != nil {
			callback(req)
		}
	}, resp)

	errResp, ok := err.(*ErrResp)
	if !ok {
		return nil, err
	}

	switch errResp.ErrorCode {
	case 0:
		return data, nil
	case 4122, 4121, 10, 16:
		if xc.refreshTokenFunc != nil {
			if err = xc.refreshTokenFunc(); err == nil {
				break
			}
		}
		return nil, err
	case 9:
		// space_token 获取失败
		if errResp.ErrorMsg == "space_token_invalid" {
			if token, err := xc.GetSafeAccessToken(xc.Token); err != nil {
				return nil, err
			} else {
				xc.SetSpaceTokenResp(token)
			}

		}
		if errResp.ErrorMsg == "captcha_invalid" {
			// 验证码token过期
			if err = xc.RefreshCaptchaTokenAtLogin(GetAction(method, url), xc.UserID); err != nil {
				return nil, err
			}
		}
		return nil, err
	default:
		return nil, err
	}
	return xc.Request(url, method, callback, resp)
}

// RefreshToken 刷新Token
func (xc *XunLeiBrowserCommon) RefreshToken(refreshToken string) (*TokenResp, error) {
	var resp TokenResp
	_, err := xc.Common.Request(XLUSER_API_URL+"/auth/token", http.MethodPost, func(req *resty.Request) {
		req.SetBody(&base.Json{
			"grant_type":    "refresh_token",
			"refresh_token": refreshToken,
			"client_id":     xc.ClientID,
			"client_secret": xc.ClientSecret,
		})
	}, &resp)
	if err != nil {
		return nil, err
	}

	if resp.RefreshToken == "" {
		return nil, errors.New("refresh token is empty")
	}
	return &resp, nil
}

// GetSafeAccessToken 获取 超级保险柜 AccessToken
func (xc *XunLeiBrowserCommon) GetSafeAccessToken(safePassword string) (string, error) {
	var resp TokenResp
	_, err := xc.Request(XLUSER_API_URL+"/password/check", http.MethodPost, func(req *resty.Request) {
		req.SetBody(&base.Json{
			"scene":    "box",
			"password": EncryptPassword(safePassword),
		})
	}, &resp)
	if err != nil {
		return "", err
	}

	if resp.Token == "" {
		return "", errors.New("SafePassword is incorrect ")
	}
	return resp.Token, nil
}

// Login 登录
func (xc *XunLeiBrowserCommon) Login(username, password string) (*TokenResp, error) {
	url := XLUSER_API_URL + "/auth/signin"
	err := xc.RefreshCaptchaTokenInLogin(GetAction(http.MethodPost, url), username)
	if err != nil {
		return nil, err
	}

	var resp TokenResp
	_, err = xc.Common.Request(url, http.MethodPost, func(req *resty.Request) {
		req.SetBody(&SignInRequest{
			CaptchaToken: xc.GetCaptchaToken(),
			ClientID:     xc.ClientID,
			ClientSecret: xc.ClientSecret,
			Username:     username,
			Password:     password,
		})
	}, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (xc *XunLeiBrowserCommon) IsLogin() bool {
	if xc.TokenResp == nil {
		return false
	}
	_, err := xc.Request(XLUSER_API_URL+"/user/me", http.MethodGet, nil, nil)
	return err == nil
}

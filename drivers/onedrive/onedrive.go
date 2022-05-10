package onedrive

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/go-resty/resty/v2"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
)

type Host struct {
	Oauth string
	Api   string
}

var onedriveHostMap = map[string]Host{
	"global": {
		Oauth: "https://login.microsoftonline.com",
		Api:   "https://graph.microsoft.com",
	},
	"cn": {
		Oauth: "https://login.chinacloudapi.cn",
		Api:   "https://microsoftgraph.chinacloudapi.cn",
	},
	"us": {
		Oauth: "https://login.microsoftonline.us",
		Api:   "https://graph.microsoft.us",
	},
	"de": {
		Oauth: "https://login.microsoftonline.de",
		Api:   "https://graph.microsoft.de",
	},
}

func (driver Onedrive) GetMetaUrl(account *model.Account, auth bool, path string) string {
	path = filepath.Join(account.RootFolder, path)
	//log.Debugf(path)
	host, _ := onedriveHostMap[account.Zone]
	if auth {
		return host.Oauth
	}
	switch account.InternalType {
	case "onedrive":
		{
			if path == "/" || path == "\\" {
				return fmt.Sprintf("%s/v1.0/me/drive/root", host.Api)
			} else {
				return fmt.Sprintf("%s/v1.0/me/drive/root:%s:", host.Api, path)
			}
		}
	case "sharepoint":
		{
			if path == "/" || path == "\\" {
				return fmt.Sprintf("%s/v1.0/sites/%s/drive/root", host.Api, account.SiteId)
			} else {
				return fmt.Sprintf("%s/v1.0/sites/%s/drive/root:%s:", host.Api, account.SiteId, path)
			}
		}
	default:
		return ""
	}
}

type OneTokenErr struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (driver Onedrive) RefreshToken(account *model.Account) error {
	err := driver.refreshToken(account)
	if err != nil && err == base.ErrEmptyToken {
		return driver.refreshToken(account)
	}
	return err
}

func (driver Onedrive) refreshToken(account *model.Account) error {
	url := driver.GetMetaUrl(account, true, "") + "/common/oauth2/v2.0/token"
	var resp base.TokenResp
	var e OneTokenErr
	_, err := base.RestyClient.R().SetResult(&resp).SetError(&e).SetFormData(map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     account.ClientId,
		"client_secret": account.ClientSecret,
		"redirect_uri":  account.RedirectUri,
		"refresh_token": account.RefreshToken,
	}).Post(url)
	if err != nil {
		account.Status = err.Error()
		return err
	}
	if e.Error != "" {
		account.Status = e.ErrorDescription
		return fmt.Errorf("%s", e.ErrorDescription)
	} else {
		account.Status = "work"
	}
	if resp.RefreshToken == "" {
		account.Status = base.ErrEmptyToken.Error()
		return base.ErrEmptyToken
	}
	account.RefreshToken, account.AccessToken = resp.RefreshToken, resp.AccessToken
	return nil
}

type OneFile struct {
	Id                   string     `json:"id"`
	Name                 string     `json:"name"`
	Size                 int64      `json:"size"`
	LastModifiedDateTime *time.Time `json:"lastModifiedDateTime"`
	Url                  string     `json:"@microsoft.graph.downloadUrl"`
	File                 *struct {
		MimeType string `json:"mimeType"`
	} `json:"file"`
	Thumbnails []struct {
		Medium struct {
			Url string `json:"url"`
		} `json:"medium"`
	} `json:"thumbnails"`
	ParentReference struct {
		DriveId string `json:"driveId"`
	} `json:"parentReference"`
}

type OneFiles struct {
	Value    []OneFile `json:"value"`
	NextLink string    `json:"@odata.nextLink"`
}

type OneRespErr struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (driver Onedrive) FormatFile(file *OneFile) *model.File {
	f := &model.File{
		Name:      file.Name,
		Size:      file.Size,
		UpdatedAt: file.LastModifiedDateTime,
		Driver:    driver.Config().Name,
		Url:       file.Url,
		Id:        file.Id,
	}
	if len(file.Thumbnails) > 0 {
		f.Thumbnail = file.Thumbnails[0].Medium.Url
	}
	if file.File == nil {
		f.Type = conf.FOLDER
	} else {
		f.Type = utils.GetFileType(filepath.Ext(file.Name))
	}
	return f
}

func (driver Onedrive) GetFiles(account *model.Account, path string) ([]OneFile, error) {
	var res []OneFile
	nextLink := driver.GetMetaUrl(account, false, path) + "/children?$expand=thumbnails"
	if account.OrderBy != "" {
		nextLink += fmt.Sprintf("&orderby=%s", account.OrderBy)
		if account.OrderDirection != "" {
			nextLink += fmt.Sprintf("%%20%s", account.OrderDirection)
		}
	}
	for nextLink != "" {
		var files OneFiles
		_, err := driver.Request(nextLink, base.Get, nil, nil, nil, nil, &files, account)
		//var e OneRespErr
		//_, err := oneClient.R().SetResult(&files).SetError(&e).
		//	SetHeader("Authorization", "Bearer  "+account.AccessToken).
		//	Get(nextLink)
		if err != nil {
			return nil, err
		}
		//if e.Error.Code != "" {
		//	return nil, fmt.Errorf("%s", e.Error.Message)
		//}
		res = append(res, files.Value...)
		nextLink = files.NextLink
	}
	return res, nil
}

func (driver Onedrive) GetFile(account *model.Account, path string) (*OneFile, error) {
	var file OneFile
	//var e OneRespErr
	u := driver.GetMetaUrl(account, false, path)
	_, err := driver.Request(u, base.Get, nil, nil, nil, nil, &file, account)
	//_, err := oneClient.R().SetResult(&file).SetError(&e).
	//	SetHeader("Authorization", "Bearer  "+account.AccessToken).
	//	Get(driver.GetMetaUrl(account, false, path))
	if err != nil {
		return nil, err
	}
	//if e.Error.Code != "" {
	//	return nil, fmt.Errorf("%s", e.Error.Message)
	//}
	return &file, nil
}

func (driver Onedrive) Request(url string, method int, headers, query, form map[string]string, data interface{}, resp interface{}, account *model.Account) ([]byte, error) {
	rawUrl := url
	if account.APIProxyUrl != "" {
		url = fmt.Sprintf("%s/%s", account.APIProxyUrl, url)
	}
	req := base.RestyClient.R()
	req.SetHeader("Authorization", "Bearer "+account.AccessToken)
	if headers != nil {
		req.SetHeaders(headers)
	}
	if query != nil {
		req.SetQueryParams(query)
	}
	if form != nil {
		req.SetFormData(form)
	}
	if data != nil {
		req.SetBody(data)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	var res *resty.Response
	var err error
	var e OneRespErr
	req.SetError(&e)
	switch method {
	case base.Get:
		res, err = req.Get(url)
	case base.Post:
		res, err = req.Post(url)
	case base.Patch:
		res, err = req.Patch(url)
	case base.Delete:
		res, err = req.Delete(url)
	case base.Put:
		res, err = req.Put(url)
	default:
		return nil, base.ErrNotSupport
	}
	if err != nil {
		return nil, err
	}
	//log.Debug(res.String())
	if e.Error.Code != "" {
		if e.Error.Code == "InvalidAuthenticationToken" {
			err = driver.RefreshToken(account)
			if err != nil {
				_ = model.SaveAccount(account)
				return nil, err
			}
			return driver.Request(rawUrl, method, headers, query, form, data, resp, account)
		}
		return nil, errors.New(e.Error.Message)
	}
	return res.Body(), nil
}

func (driver Onedrive) UploadSmall(file *model.FileStream, account *model.Account) error {
	url := driver.GetMetaUrl(account, false, utils.Join(file.ParentPath, file.Name)) + "/content"
	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	_, err = driver.Request(url, base.Put, nil, nil, nil, data, nil, account)
	return err
}

func (driver Onedrive) UploadBig(file *model.FileStream, account *model.Account) error {
	url := driver.GetMetaUrl(account, false, utils.Join(file.ParentPath, file.Name)) + "/createUploadSession"
	res, err := driver.Request(url, base.Post, nil, nil, nil, nil, nil, account)
	if err != nil {
		return err
	}
	uploadUrl := jsoniter.Get(res, "uploadUrl").ToString()
	var finish uint64 = 0
	const DEFAULT = 4 * 1024 * 1024
	for finish < file.GetSize() {
		log.Debugf("upload: %d", finish)
		var byteSize uint64 = DEFAULT
		left := file.GetSize() - finish
		if left < DEFAULT {
			byteSize = left
		}
		byteData := make([]byte, byteSize)
		n, err := io.ReadFull(file, byteData)
		log.Debug(err, n)
		if err != nil {
			return err
		}
		req, err := http.NewRequest("PUT", uploadUrl, bytes.NewBuffer(byteData))
		req.Header.Set("Content-Length", strconv.Itoa(int(byteSize)))
		req.Header.Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", finish, finish+byteSize-1, file.Size))
		finish += byteSize
		res, err := base.HttpClient.Do(req)
		if res.StatusCode != 201 && res.StatusCode != 202 {
			data, _ := ioutil.ReadAll(res.Body)
			res.Body.Close()
			return errors.New(string(data))
		}
		res.Body.Close()
	}
	return nil
}

func init() {
	base.RegisterDriver(&Onedrive{})
}

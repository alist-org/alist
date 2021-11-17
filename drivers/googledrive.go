package drivers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strconv"
	"time"
)

type GoogleDrive struct {
}

var googleClient = resty.New()

func (g GoogleDrive) Items() []Item {
	return []Item{
		{
			Name:     "client_id",
			Label:    "client id",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "client_secret",
			Label:    "client secret",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "refresh_token",
			Label:    "refresh token",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     "string",
			Required: true,
		},
	}
}

type GoogleTokenError struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (g GoogleDrive) RefreshToken(account *model.Account) error {
	url := "https://www.googleapis.com/oauth2/v4/token"
	var resp TokenResp
	var e GoogleTokenError
	_, err := googleClient.R().SetResult(&resp).SetError(&e).
		SetFormData(map[string]string{
			"client_id":     account.ClientId,
			"client_secret": account.ClientSecret,
			"refresh_token": account.RefreshToken,
			"grant_type":    "refresh_token",
		}).Post(url)
	if err != nil {
		return err
	}
	if e.Error != "" {
		return fmt.Errorf(e.Error)
	}
	account.AccessToken = resp.AccessToken
	account.Status = "work"
	return nil
}

func (g GoogleDrive) Save(account *model.Account, old *model.Account) error {
	account.Proxy = true
	err := g.RefreshToken(account)
	if err != nil {
		account.Status = err.Error()
		_ = model.SaveAccount(account)
		return err
	}
	account.Status = "work"
	_ = model.SaveAccount(account)
	return nil
}

type GoogleFile struct {
	Id           string     `json:"id"`
	Name         string     `json:"name"`
	MimeType     string     `json:"mimeType"`
	ModifiedTime *time.Time `json:"modifiedTime"`
	Size         string     `json:"size"`
}

func (g GoogleDrive) IsDir(mimeType string) bool {
	return mimeType == "application/vnd.google-apps.folder" || mimeType == "application/vnd.google-apps.shortcut"
}

func (g GoogleDrive) FormatFile(file *GoogleFile) *model.File {
	f := &model.File{
		Name:      file.Name,
		Driver:    "GoogleDrive",
		UpdatedAt: file.ModifiedTime,
		Thumbnail: "",
		Url:       "",
	}
	if g.IsDir(file.MimeType) {
		f.Type = conf.FOLDER
	} else {
		size, _ := strconv.ParseInt(file.Size, 10, 64)
		f.Size = size
		f.Type = utils.GetFileType(filepath.Ext(file.Name))
	}
	return f
}

type GoogleFiles struct {
	NextPageToken string       `json:"nextPageToken"`
	Files         []GoogleFile `json:"files"`
}

type GoogleError struct {
	Error struct {
		Errors []struct {
			Domain       string `json:"domain"`
			Reason       string `json:"reason"`
			Message      string `json:"message"`
			LocationType string `json:"location_type"`
			Location     string `json:"location"`
		}
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (g GoogleDrive) GetFiles(id string, account *model.Account) ([]GoogleFile, error) {
	pageToken := "first"
	res := make([]GoogleFile, 0)
	for pageToken != "" {
		if pageToken == "first" {
			pageToken = ""
		}
		var resp GoogleFiles
		var e GoogleError
		_, err := googleClient.R().SetResult(&resp).SetError(&e).
			SetHeader("Authorization", "Bearer "+account.AccessToken).
			SetQueryParams(map[string]string{
				"orderBy":                   "folder,name,modifiedTime desc",
				"fields":                    "files(id,name,mimeType,size,modifiedTime),nextPageToken",
				"pageSize":                  "1000",
				"q":                         fmt.Sprintf("'%s' in parents and trashed = false", id),
				"includeItemsFromAllDrives": "true",
				"supportsAllDrives":         "true",
				"pageToken":                 pageToken,
			}).Get("https://www.googleapis.com/drive/v3/files")
		if err != nil {
			return nil, err
		}
		if e.Error.Code != 0 {
			if e.Error.Code == 401 {
				err = g.RefreshToken(account)
				if err != nil {
					_ = model.SaveAccount(account)
					return nil, err
				}
				return g.GetFiles(id, account)
			}
			return nil, fmt.Errorf("%s: %v", e.Error.Message, e.Error.Errors)
		}
		pageToken = resp.NextPageToken
		res = append(res, resp.Files...)
	}
	return res, nil
}

func (g GoogleDrive) GetFile(path string, account *model.Account) (*GoogleFile, error) {
	dir, name := filepath.Split(path)
	dir = utils.ParsePath(dir)
	_, _, err := g.Path(dir, account)
	if err != nil {
		return nil, err
	}
	parentFiles_, _ := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, dir))
	parentFiles, _ := parentFiles_.([]GoogleFile)
	for _, file := range parentFiles {
		if file.Name == name {
			if !g.IsDir(file.MimeType) {
				return &file, err
			} else {
				return nil, fmt.Errorf("not file")
			}
		}
	}
	return nil, fmt.Errorf("path not found")
}

func (g GoogleDrive) Path(path string, account *model.Account) (*model.File, []*model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("ali path: %s", path)
	cache, err := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path))
	if err == nil {
		files, _ := cache.([]GoogleFile)
		if len(files) != 0 {
			res := make([]*model.File, 0)
			for _, file := range files {
				res = append(res, g.FormatFile(&file))
			}
			return nil, res, nil
		}
	}
	// no cache or len(files) == 0
	fileId := account.RootFolder
	if path != "/" {
		dir, name := filepath.Split(path)
		dir = utils.ParsePath(dir)
		_, _, err = g.Path(dir, account)
		if err != nil {
			return nil, nil, err
		}
		parentFiles_, _ := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, dir))
		parentFiles, _ := parentFiles_.([]GoogleFile)
		found := false
		for _, file := range parentFiles {
			if file.Name == name {
				found = true
				if !g.IsDir(file.MimeType) {
					return g.FormatFile(&file), nil, nil
				} else {
					fileId = file.Id
					break
				}
			}
		}
		if !found {
			return nil, nil, fmt.Errorf("path not found")
		}
	}
	files, err := g.GetFiles(fileId, account)
	if err != nil {
		return nil, nil, err
	}
	_ = conf.Cache.Set(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path), files, nil)
	res := make([]*model.File, 0)
	for _, file := range files {
		res = append(res, g.FormatFile(&file))
	}
	return nil, res, nil
}

func (g GoogleDrive) Link(path string, account *model.Account) (string, error) {
	file, err := g.GetFile(utils.ParsePath(path), account)
	if err != nil {
		return "", err
	}
	link := fmt.Sprintf("https://www.googleapis.com/drive/v3/files/%s?includeItemsFromAllDrives=true&supportsAllDrives=true", file.Id)
	var e GoogleError
	_, _ = googleClient.R().SetError(&e).
		SetHeader("Authorization", "Bearer "+account.AccessToken).
		Get(link)
	if e.Error.Code != 0 {
		if e.Error.Code == 401 {
			err = g.RefreshToken(account)
			if err != nil {
				_ = model.SaveAccount(account)
				return "", err
			}
			return g.Link(path, account)
		}
		return "", fmt.Errorf("%s: %v", e.Error.Message, e.Error.Errors)
	}
	return link + "&alt=media", nil
}

func (g GoogleDrive) Proxy(c *gin.Context, account *model.Account) {
	c.Request.Header.Add("Authorization", "Bearer "+account.AccessToken)
}

func (g GoogleDrive) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, nil
}

var _ Driver = (*GoogleDrive)(nil)

func init() {
	RegisterDriver("GoogleDrive", &GoogleDrive{})
}

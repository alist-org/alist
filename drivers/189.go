package drivers

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	mathRand "math/rand"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Cloud189 struct {
}

var client189Map map[string]*resty.Client

func (c Cloud189) Items() []Item {
	return []Item{
		{
			Name:        "proxy",
			Label:       "proxy",
			Type:        "bool",
			Required:    true,
			Description: "allow proxy",
		},
		{
			Name:        "username",
			Label:       "username",
			Type:        "string",
			Required:    true,
			Description: "account username/phone number",
		},
		{
			Name:        "password",
			Label:       "password",
			Type:        "string",
			Required:    true,
			Description: "account password",
		},
		{
			Name:     "root_folder",
			Label:    "root folder file_id",
			Type:     "string",
			Required: true,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     "select",
			Values:   "name,size,lastOpTime,createdDate",
			Required: true,
		},
		{
			Name:     "order_direction",
			Label:    "desc",
			Type:     "select",
			Values:   "true,false",
			Required: true,
		},
	}
}

func (c Cloud189) Save(account *model.Account, old *model.Account) error {
	if old != nil && old.Name != account.Name {
		delete(client189Map, old.Name)
	}
	if err := c.Login(account); err != nil {
		account.Status = err.Error()
		_ = model.SaveAccount(account)
		return err
	}
	account.Status = "work"
	err := model.SaveAccount(account)
	if err != nil {
		return err
	}
	return nil
}

func (c Cloud189) FormatFile(file *Cloud189File) *model.File {
	f := &model.File{
		Name:      file.Name,
		Size:      file.Size,
		Driver:    "189Cloud",
		UpdatedAt: nil,
		Thumbnail: file.Icon.SmallUrl,
		Url:       file.Url,
	}
	loc, _ := time.LoadLocation("Local")
	lastOpTime, err := time.ParseInLocation("2006-01-02 15:04:05", file.LastOpTime, loc)
	if err == nil {
		f.UpdatedAt = &lastOpTime
	}
	if file.Size == -1 {
		f.Type = conf.FOLDER
		f.Size = 0
	} else {
		f.Type = utils.GetFileType(filepath.Ext(file.Name))
	}
	return f
}

func (c Cloud189) Path(path string, account *model.Account) (*model.File, []*model.File, error) {
	path = utils.ParsePath(path)
	log.Debugf("189 path: %s", path)
	cache, err := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path))
	if err == nil {
		files, _ := cache.([]Cloud189File)
		if len(files) != 0 {
			res := make([]*model.File, 0)
			for _, file := range files {
				res = append(res, c.FormatFile(&file))
			}
			return nil, res, nil
		}
	}
	// no cache or len(files) == 0
	fileId := account.RootFolder
	if path != "/" {
		dir, name := filepath.Split(path)
		dir = utils.ParsePath(dir)
		_, _, err = c.Path(dir, account)
		if err != nil {
			return nil, nil, err
		}
		parentFiles_, _ := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, dir))
		parentFiles, _ := parentFiles_.([]Cloud189File)
		found := false
		for _, file := range parentFiles {
			if file.Name == name {
				found = true
				if file.Size != -1 {
					url, err := c.Link(path, account)
					if err != nil {
						return nil, nil, err
					}
					file.Url = url
					return c.FormatFile(&file), nil, nil
				} else {
					fileId = strconv.FormatInt(file.Id, 10)
					break
				}
			}
		}
		if !found {
			return nil, nil, fmt.Errorf("path not found")
		}
	}
	files, err := c.GetFiles(fileId, account)
	if err != nil {
		return nil, nil, err
	}
	_ = conf.Cache.Set(conf.Ctx, fmt.Sprintf("%s%s", account.Name, path), files, nil)
	res := make([]*model.File, 0)
	for _, file := range files {
		res = append(res, c.FormatFile(&file))
	}
	return nil, res, nil
}

func (c Cloud189) GetFile(path string, account *model.Account) (*Cloud189File, error) {
	dir, name := filepath.Split(path)
	dir = utils.ParsePath(dir)
	_, _, err := c.Path(dir, account)
	if err != nil {
		return nil, err
	}
	parentFiles_, _ := conf.Cache.Get(conf.Ctx, fmt.Sprintf("%s%s", account.Name, dir))
	parentFiles, _ := parentFiles_.([]Cloud189File)
	for _, file := range parentFiles {
		if file.Name == name {
			if file.Size != -1 {
				return &file, err
			} else {
				return nil, fmt.Errorf("not file")
			}
		}
	}
	return nil, fmt.Errorf("path not found")
}

type Cloud189Down struct {
	ResCode         int    `json:"res_code"`
	ResMessage      string `json:"res_message"`
	FileDownloadUrl string `json:"fileDownloadUrl"`
}

func (c Cloud189) Link(path string, account *model.Account) (string, error) {
	file, err := c.GetFile(utils.ParsePath(path), account)
	if err != nil {
		return "", err
	}
	client, ok := client189Map[account.Name]
	if !ok {
		return "", fmt.Errorf("can't find [%s] client", account.Name)
	}
	var e Cloud189Error
	var resp Cloud189Down
	_, err = client.R().SetResult(&resp).SetError(&e).
		SetHeader("Accept", "application/json;charset=UTF-8").
		SetQueryParams(map[string]string{
			"noCache": random(),
			"fileId":  strconv.FormatInt(file.Id, 10),
		}).Get("https://cloud.189.cn/api/open/file/getFileDownloadUrl.action")
	if err != nil {
		return "", err
	}
	if e.ErrorCode != "" {
		if e.ErrorCode == "InvalidSessionKey" {
			err = c.Login(account)
			if err != nil {
				return "", err
			}
			return c.Link(path, account)
		}
	}
	if resp.ResCode != 0 {
		return "", fmt.Errorf(resp.ResMessage)
	}
	res, err := noRedirectClient.R().Get(resp.FileDownloadUrl)
	if err != nil {
		return "", err
	}
	if res.StatusCode() == 302 {
		return res.Header().Get("location"), nil
	}
	return resp.FileDownloadUrl, nil
}

func (c Cloud189) Proxy(ctx *gin.Context, account *model.Account) {
	ctx.Request.Header.Del("Origin")
}

func (c Cloud189) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, nil
}

var _ Driver = (*Cloud189)(nil)

func init() {
	RegisterDriver("189Cloud", &Cloud189{})
	client189Map = make(map[string]*resty.Client, 0)
}

type LoginResp struct {
	Msg    string `json:"msg"`
	Result int    `json:"result"`
	ToUrl  string `json:"toUrl"`
}

// Login refer to PanIndex
func (c Cloud189) Login(account *model.Account) error {
	client, ok := client189Map[account.Name]
	if !ok {
		//cookieJar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
		client = resty.New()
		//client.SetCookieJar(cookieJar)
		client.SetRetryCount(3)
	}
	url := "https://cloud.189.cn/api/portal/loginUrl.action?redirectURL=https%3A%2F%2Fcloud.189.cn%2Fmain.action"
	res, err := client.R().Get(url)
	if err != nil {
		return err
	}
	b := res.String()
	lt := ""
	ltText := regexp.MustCompile(`lt = "(.+?)"`)
	ltTextArr := ltText.FindStringSubmatch(b)
	if len(ltTextArr) > 0 {
		lt = ltTextArr[1]
	} else {
		return fmt.Errorf("ltTextArr = 0")
	}
	captchaToken := regexp.MustCompile(`captchaToken' value='(.+?)'`).FindStringSubmatch(b)[1]
	returnUrl := regexp.MustCompile(`returnUrl = '(.+?)'`).FindStringSubmatch(b)[1]
	paramId := regexp.MustCompile(`paramId = "(.+?)"`).FindStringSubmatch(b)[1]
	//reqId := regexp.MustCompile(`reqId = "(.+?)"`).FindStringSubmatch(b)[1]
	jRsakey := regexp.MustCompile(`j_rsaKey" value="(\S+)"`).FindStringSubmatch(b)[1]
	vCodeID := regexp.MustCompile(`picCaptcha\.do\?token\=([A-Za-z0-9\&\=]+)`).FindStringSubmatch(b)[1]
	vCodeRS := ""
	if vCodeID != "" {
		// need ValidateCode
	}
	userRsa := RsaEncode([]byte(account.Username), jRsakey)
	passwordRsa := RsaEncode([]byte(account.Password), jRsakey)
	url = "https://open.e.189.cn/api/logbox/oauth2/loginSubmit.do"
	var loginResp LoginResp
	res, err = client.R().
		SetHeaders(map[string]string{
			"lt":         lt,
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36",
			"Referer":    "https://open.e.189.cn/",
			"accept":     "application/json;charset=UTF-8",
		}).SetFormData(map[string]string{
		"appKey":       "cloud",
		"accountType":  "01",
		"userName":     "{RSA}" + userRsa,
		"password":     "{RSA}" + passwordRsa,
		"validateCode": vCodeRS,
		"captchaToken": captchaToken,
		"returnUrl":    returnUrl,
		"mailSuffix":   "@pan.cn",
		"paramId":      paramId,
		"clientType":   "10010",
		"dynamicCheck": "FALSE",
		"cb_SaveName":  "1",
		"isOauth2":     "false",
	}).Post(url)
	if err != nil {
		return err
	}
	err = json.Unmarshal(res.Body(), &loginResp)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	if loginResp.Result != 0 {
		return fmt.Errorf(loginResp.Msg)
	}
	_, err = client.R().Get(loginResp.ToUrl)
	if err != nil {
		log.Errorf(err.Error())
		return err
	}
	client189Map[account.Name] = client
	return nil
}

type Cloud189Error struct {
	ErrorCode string `json:"errorCode"`
	ErrorMsg  string `json:"errorMsg"`
}

type Cloud189File struct {
	Id         int64  `json:"id"`
	LastOpTime string `json:"lastOpTime"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Icon       struct {
		SmallUrl string `json:"smallUrl"`
		//LargeUrl string `json:"largeUrl"`
	} `json:"icon"`
	Url string `json:"url"`
}

type Cloud189Folder struct {
	Id         int64  `json:"id"`
	LastOpTime string `json:"lastOpTime"`
	Name       string `json:"name"`
}

type Cloud189Files struct {
	ResCode    int    `json:"res_code"`
	ResMessage string `json:"res_message"`
	FileListAO struct {
		Count      int              `json:"count"`
		FileList   []Cloud189File   `json:"fileList"`
		FolderList []Cloud189Folder `json:"folderList"`
	} `json:"fileListAO"`
}

func (c Cloud189) GetFiles(fileId string, account *model.Account) ([]Cloud189File, error) {
	client, ok := client189Map[account.Name]
	if !ok {
		return nil, fmt.Errorf("can't find [%s] client", account.Name)
	}
	res := make([]Cloud189File, 0)
	pageNum := 1
	for {
		var e Cloud189Error
		var resp Cloud189Files
		_, err := client.R().SetResult(&resp).SetError(&e).
			SetHeader("Accept", "application/json;charset=UTF-8").
			SetQueryParams(map[string]string{
				"noCache":    random(),
				"pageSize":   "60",
				"pageNum":    strconv.Itoa(pageNum),
				"mediaType":  "0",
				"folderId":   fileId,
				"iconOption": "5",
				"orderBy":    account.OrderBy,
				"descending": account.OrderDirection,
			}).Get("https://cloud.189.cn/api/open/file/listFiles.action")
		if err != nil {
			return nil, err
		}
		if e.ErrorCode != "" {
			if e.ErrorCode == "InvalidSessionKey" {
				err = c.Login(account)
				if err != nil {
					return nil, err
				}
				return c.GetFiles(fileId, account)
			}
		}
		if resp.ResCode != 0 {
			return nil, fmt.Errorf(resp.ResMessage)
		}
		if resp.FileListAO.Count == 0 {
			break
		}
		res = append(res, resp.FileListAO.FileList...)
		for _, folder := range resp.FileListAO.FolderList {
			res = append(res, Cloud189File{
				Id:         folder.Id,
				LastOpTime: folder.LastOpTime,
				Name:       folder.Name,
				Size:       -1,
			})
		}
		pageNum++
	}
	return res, nil
}

func random() string {
	return fmt.Sprintf("0.%17v", mathRand.New(mathRand.NewSource(time.Now().UnixNano())).Int63n(100000000000000000))
}

func RsaEncode(origData []byte, j_rsakey string) string {
	publicKey := []byte("-----BEGIN PUBLIC KEY-----\n" + j_rsakey + "\n-----END PUBLIC KEY-----")
	block, _ := pem.Decode(publicKey)
	pubInterface, _ := x509.ParsePKIXPublicKey(block.Bytes)
	pub := pubInterface.(*rsa.PublicKey)
	b, err := rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
	if err != nil {
		log.Errorf("err: %s", err.Error())
	}
	return b64tohex(base64.StdEncoding.EncodeToString(b))
}

var b64map = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

var BI_RM = "0123456789abcdefghijklmnopqrstuvwxyz"

func int2char(a int) string {
	return strings.Split(BI_RM, "")[a]
}

func b64tohex(a string) string {
	d := ""
	e := 0
	c := 0
	for i := 0; i < len(a); i++ {
		m := strings.Split(a, "")[i]
		if m != "=" {
			v := strings.Index(b64map, m)
			if 0 == e {
				e = 1
				d += int2char(v >> 2)
				c = 3 & v
			} else if 1 == e {
				e = 2
				d += int2char(c<<2 | v>>4)
				c = 15 & v
			} else if 2 == e {
				e = 3
				d += int2char(c)
				d += int2char(v >> 2)
				c = 3 & v
			} else {
				e = 0
				d += int2char(c<<2 | v>>4)
				d += int2char(15 & v)
			}
		}
	}
	if e == 1 {
		d += int2char(c << 2)
	}
	return d
}

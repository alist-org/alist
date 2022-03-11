package lanzou

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

type LanZouFile struct {
	Name    string `json:"name"`
	NameAll string `json:"name_all"`
	Id      string `json:"id"`
	FolId   string `json:"fol_id"`
	Size    string `json:"size"`
	Time    string `json:"time"`
	Folder  bool
}

func (driver *Lanzou) FormatFile(file *LanZouFile) *model.File {
	now := time.Now()
	f := &model.File{
		Id:        file.Id,
		Name:      file.Name,
		Driver:    driver.Config().Name,
		SizeStr:   file.Size,
		TimeStr:   file.Time,
		UpdatedAt: &now,
	}
	if file.Folder {
		f.Type = conf.FOLDER
		f.Id = file.FolId
	} else {
		f.Name = file.NameAll
		f.Type = utils.GetFileType(filepath.Ext(file.NameAll))
	}
	return f
}

type LanZouFilesResp struct {
	Zt   int          `json:"zt"`
	Info interface{}  `json:"info"`
	Text []LanZouFile `json:"text"`
}

func (driver *Lanzou) GetFiles(folderId string, account *model.Account) ([]LanZouFile, error) {
	if account.InternalType == "cookie" {
		files := make([]LanZouFile, 0)
		var resp LanZouFilesResp
		// folders
		res, err := base.RestyClient.R().SetResult(&resp).SetHeader("Cookie", account.AccessToken).
			SetFormData(map[string]string{
				"task":      "47",
				"folder_id": folderId,
			}).Post("https://pc.woozooo.com/doupload.php")
		if err != nil {
			return nil, err
		}
		log.Debug(res.String())
		if resp.Zt != 1 && resp.Zt != 2 {
			return nil, fmt.Errorf("%v", resp.Info)
		}
		for _, file := range resp.Text {
			file.Folder = true
			files = append(files, file)
		}
		// files
		pg := 1
		for {
			_, err = base.RestyClient.R().SetResult(&resp).SetHeader("Cookie", account.AccessToken).
				SetFormData(map[string]string{
					"task":      "5",
					"folder_id": folderId,
					"pg":        strconv.Itoa(pg),
				}).Post("https://pc.woozooo.com/doupload.php")
			if err != nil {
				return nil, err
			}
			if resp.Zt != 1 {
				return nil, fmt.Errorf("%v", resp.Info)
			}
			if len(resp.Text) == 0 {
				break
			}
			files = append(files, resp.Text...)
			pg++
		}
		return files, nil
	} else {
		return driver.GetFilesByUrl(account)
	}
}

func (driver *Lanzou) GetFilesByUrl(account *model.Account) ([]LanZouFile, error) {
	files := make([]LanZouFile, 0)
	shareUrl := account.SiteUrl
	u, err := url.Parse(shareUrl)
	if err != nil {
		return nil, err
	}
	res, err := base.RestyClient.R().Get(shareUrl)
	if err != nil {
		return nil, err
	}
	lxArr := regexp.MustCompile(`'lx':(.+?),`).FindStringSubmatch(res.String())
	if len(lxArr) == 0 {
		return nil, fmt.Errorf("get empty page")
	}
	lx := lxArr[1]
	fid := regexp.MustCompile(`'fid':(.+?),`).FindStringSubmatch(res.String())[1]
	uid := regexp.MustCompile(`'uid':'(.+?)',`).FindStringSubmatch(res.String())[1]
	rep := regexp.MustCompile(`'rep':'(.+?)',`).FindStringSubmatch(res.String())[1]
	up := regexp.MustCompile(`'up':(.+?),`).FindStringSubmatch(res.String())[1]
	ls := ""
	if account.Password != "" {
		ls = regexp.MustCompile(`'ls':(.+?),`).FindStringSubmatch(res.String())[1]
	}
	tName := regexp.MustCompile(`'t':(.+?),`).FindStringSubmatch(res.String())[1]
	kName := regexp.MustCompile(`'k':(.+?),`).FindStringSubmatch(res.String())[1]
	t := regexp.MustCompile(`var ` + tName + ` = '(.+?)';`).FindStringSubmatch(res.String())[1]
	k := regexp.MustCompile(`var ` + kName + ` = '(.+?)';`).FindStringSubmatch(res.String())[1]
	pg := 1
	for {
		var resp LanZouFilesResp
		res, err = base.RestyClient.R().SetResult(&resp).SetFormData(map[string]string{
			"lx":  lx,
			"fid": fid,
			"uid": uid,
			"pg":  strconv.Itoa(pg),
			"rep": rep,
			"t":   t,
			"k":   k,
			"up":  up,
			"ls":  ls,
			"pwd": account.Password,
		}).Post(fmt.Sprintf("https://%s/filemoreajax.php", u.Host))
		if err != nil {
			log.Debug(err)
			break
		}
		log.Debug(res.String())
		if resp.Zt != 1 {
			return nil, fmt.Errorf("%v", resp.Info)
		}
		if len(resp.Text) == 0 {
			break
		}
		pg++
		time.Sleep(time.Second)
		files = append(files, resp.Text...)
	}
	return files, nil
}

//type LanzouDownInfo struct {
//	FId    string `json:"f_id"`
//	IsNewd string `json:"is_newd"`
//}

// GetDownPageId 获取下载页面的ID
func (driver *Lanzou) GetDownPageId(fileId string, account *model.Account) (string, string, error) {
	var resp DownPageResp
	res, err := base.RestyClient.R().SetResult(&resp).SetHeader("Cookie", account.AccessToken).
		SetFormData(map[string]string{
			"task":    "22",
			"file_id": fileId,
		}).Post("https://pc.woozooo.com/doupload.php")
	if err != nil {
		return "", "", err
	}
	log.Debug(res.String())
	if resp.Zt != 1 {
		return "", "", fmt.Errorf("%v", resp.Info)
	}
	return resp.Info.FId, resp.Info.Pwd, nil
}

type LanzouLinkResp struct {
	Dom string `json:"dom"`
	Url string `json:"url"`
	Zt  int    `json:"zt"`
}

func (driver *Lanzou) GetLink(downId string, pwd string, account *model.Account) (string, error) {
	shareUrl := account.SiteUrl
	u, err := url.Parse(shareUrl)
	if err != nil {
		return "", err
	}
	log.Debugln(fmt.Sprintf("https://%s/%s", u.Host, downId))
	res, err := base.RestyClient.R().Get(fmt.Sprintf("https://%s/%s", u.Host, downId))
	if err != nil {
		return "", err
	}
	iframe := regexp.MustCompile(`<iframe class="ifr2" name=".{2,20}" src="(.+?)"`).FindStringSubmatch(res.String())
	if len(iframe) == 0 {
		return driver.GetLinkWithPassword(downId, pwd, res.String(), account)
	}
	iframeUrl := fmt.Sprintf("https://%s%s", u.Host, iframe[1])
	res, err = base.RestyClient.R().Get(iframeUrl)
	if err != nil {
		return "", err
	}
	log.Debugln(res.String())
	ajaxdata := regexp.MustCompile(`var ajaxdata = '(.+?)'`).FindStringSubmatch(res.String())
	if len(ajaxdata) == 0 {
		return "", fmt.Errorf("get iframe empty page")
	}
	signs := ajaxdata[1]
	//sign := regexp.MustCompile(`var ispostdowns = '(.+?)';`).FindStringSubmatch(res.String())[1]
	sign := regexp.MustCompile(`'sign':'(.+?)',`).FindStringSubmatch(res.String())[1]
	//websign := regexp.MustCompile(`'websign':'(.+?)'`).FindStringSubmatch(res.String())[1]
	websign := ""
	websignR := regexp.MustCompile(`var websign = '(.+?)'`).FindStringSubmatch(res.String())
	if len(websignR) > 1 {
		websign = websignR[1]
	}
	//websign := ""
	//websignkey := regexp.MustCompile(`'websignkey':'(.+?)'`).FindStringSubmatch(res.String())[1]
	websignkey := regexp.MustCompile(`var websignkey = '(.+?)';`).FindStringSubmatch(res.String())[1]
	var resp LanzouLinkResp
	form := map[string]string{
		"action":     "downprocess",
		"signs":      signs,
		"sign":       sign,
		"ves":        "1",
		"websign":    websign,
		"websignkey": websignkey,
	}
	log.Debugf("form: %+v", form)
	res, err = base.RestyClient.R().SetResult(&resp).
		SetHeader("origin", "https://"+u.Host).
		SetHeader("referer", iframeUrl).
		SetFormData(form).Post(fmt.Sprintf("https://%s/ajaxm.php", u.Host))
	log.Debug(res.String())
	if err != nil {
		return "", err
	}
	if resp.Zt == 1 {
		return resp.Dom + "/file/" + resp.Url, nil
	}
	return "", fmt.Errorf("failed get link")
}

func (driver *Lanzou) GetLinkWithPassword(downId string, pwd string, html string, account *model.Account) (string, error) {
	shareUrl := account.SiteUrl
	u, err := url.Parse(shareUrl)
	if err != nil {
		return "", err
	}
	if html == "" {
		log.Debugln(fmt.Sprintf("https://%s/%s", u.Host, downId))
		res, err := base.RestyClient.R().Get(fmt.Sprintf("https://%s/%s", u.Host, downId))
		if err != nil {
			return "", err
		}
		html = res.String()
	}

	data := regexp.MustCompile(`data : '(.+?)'\+pwd,`).FindStringSubmatch(html)[1] + pwd
	var resp LanzouLinkResp
	_, err = base.RestyClient.R().SetResult(&resp).SetHeaders(map[string]string{
		"Referer":      fmt.Sprintf("https://%s/%s", u.Host, downId),
		"Origin":       "https://" + u.Host,
		"content-type": "application/x-www-form-urlencoded",
	}).SetBody(data).Post(fmt.Sprintf("https://%s/ajaxm.php", u.Host))
	if err != nil {
		return "", err
	}
	if resp.Zt == 1 {
		return resp.Dom + "/file/" + resp.Url, nil
	}
	return "", fmt.Errorf("failed get link with password")
}

func init() {
	base.RegisterDriver(&Lanzou{})
}

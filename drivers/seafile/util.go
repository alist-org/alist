package seafile

import (
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/pkg/utils"
	"net/http"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/go-resty/resty/v2"
)

func (d *Seafile) getToken() error {
	if d.Token != "" {
		d.authorization = fmt.Sprintf("Token %s", d.Token)
		return nil
	}
	var authResp AuthTokenResp
	res, err := base.RestyClient.R().
		SetResult(&authResp).
		SetFormData(map[string]string{
			"username": d.UserName,
			"password": d.Password,
		}).
		Post(d.Address + "/api2/auth-token/")
	if err != nil {
		return err
	}
	if res.StatusCode() >= 400 {
		return fmt.Errorf("get token failed: %s", res.String())
	}
	d.authorization = fmt.Sprintf("Token %s", authResp.Token)
	return nil
}

func (d *Seafile) request(method string, pathname string, callback base.ReqCallback, noRedirect ...bool) ([]byte, error) {
	full := pathname
	if !strings.HasPrefix(pathname, "http") {
		full = d.Address + pathname
	}
	req := base.RestyClient.R()
	if len(noRedirect) > 0 && noRedirect[0] {
		req = base.NoRedirectClient.R()
	}
	req.SetHeader("Authorization", d.authorization)
	callback(req)
	var (
		res *resty.Response
		err error
	)
	for i := 0; i < 2; i++ {
		res, err = req.Execute(method, full)
		if err != nil {
			return nil, err
		}
		if res.StatusCode() != 401 { // Unauthorized
			break
		}
		err = d.getToken()
		if err != nil {
			return nil, err
		}
	}
	if res.StatusCode() >= 400 {
		return nil, fmt.Errorf("request failed: %s", res.String())
	}
	return res.Body(), nil
}

func (d *Seafile) getRepoAndPath(fullPath string) (repo *LibraryInfo, path string, err error) {
	libraryMap := d.libraryMap
	repoId := d.Addition.RepoId
	if repoId != "" {
		if len(repoId) == 36 /* uuid */ {
			for _, library := range libraryMap {
				if library.Id == repoId {
					return library, fullPath, nil
				}
			}
		}
	} else {
		var repoName string
		str := fullPath[1:]
		pos := strings.IndexRune(str, '/')
		if pos == -1 {
			repoName = str
		} else {
			repoName = str[:pos]
		}
		path = utils.FixAndCleanPath(fullPath[1+len(repoName):])
		if library, ok := libraryMap[repoName]; ok {
			return library, path, nil
		}
	}
	return nil, "", errs.ObjectNotFound
}

func (d *Seafile) listLibraries() (resp []LibraryItemResp, err error) {
	repoId := d.Addition.RepoId
	if repoId == "" {
		_, err = d.request(http.MethodGet, "/api2/repos/", func(req *resty.Request) {
			req.SetResult(&resp)
		})
	} else {
		var oneResp LibraryItemResp
		_, err = d.request(http.MethodGet, fmt.Sprintf("/api2/repos/%s/", repoId), func(req *resty.Request) {
			req.SetResult(&oneResp)
		})
		if err == nil {
			resp = append(resp, oneResp)
		}
	}
	if err != nil {
		return nil, err
	}
	libraryMap := make(map[string]*LibraryInfo)
	var putLibraryMap func(library LibraryItemResp, index int)
	putLibraryMap = func(library LibraryItemResp, index int) {
		name := library.Name
		if index > 0 {
			name = fmt.Sprintf("%s (%d)", name, index)
		}
		if _, exist := libraryMap[name]; exist {
			putLibraryMap(library, index+1)
		} else {
			libraryInfo := LibraryInfo{}
			data, _ := utils.Json.Marshal(library)
			_ = utils.Json.Unmarshal(data, &libraryInfo)
			libraryMap[name] = &libraryInfo
		}
	}
	for _, library := range resp {
		putLibraryMap(library, 0)
	}
	d.libraryMap = libraryMap
	return resp, nil
}

var repoPwdNotConfigured = errors.New("library password not configured")
var repoPwdIncorrect = errors.New("library password is incorrect")

func (d *Seafile) decryptLibrary(repo *LibraryInfo) (err error) {
	if !repo.Encrypted {
		return nil
	}
	if d.RepoPwd == "" {
		return repoPwdNotConfigured
	}
	now := time.Now()
	decryptedTime := repo.decryptedTime
	if repo.decryptedSuccess {
		if now.Sub(decryptedTime).Minutes() <= 30 {
			return nil
		}
	} else {
		if now.Sub(decryptedTime).Seconds() <= 10 {
			return repoPwdIncorrect
		}
	}
	var resp string
	_, err = d.request(http.MethodPost, fmt.Sprintf("/api2/repos/%s/", repo.Id), func(req *resty.Request) {
		req.SetResult(&resp).SetFormData(map[string]string{
			"password": d.RepoPwd,
		})
	})
	repo.decryptedTime = time.Now()
	if err != nil || !strings.Contains(resp, "success") {
		repo.decryptedSuccess = false
		return err
	}
	repo.decryptedSuccess = true
	return nil
}



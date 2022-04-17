package baidu

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Baidu struct{}

func (driver Baidu) Config() base.DriverConfig {
	return base.DriverConfig{
		Name: "Baidu.Disk",
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
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     base.TypeString,
			Default:  "/",
			Required: true,
		},
		{
			Name:     "order_by",
			Label:    "order_by",
			Type:     base.TypeSelect,
			Default:  "name",
			Values:   "name,time,size",
			Required: false,
		},
		{
			Name:     "order_direction",
			Label:    "order_direction",
			Type:     base.TypeSelect,
			Values:   "asc,desc",
			Default:  "asc",
			Required: false,
		},
		{
			Name:     "internal_type",
			Label:    "download api",
			Type:     base.TypeSelect,
			Required: true,
			Values:   "official,crack",
			Default:  "official",
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
	dir, name := filepath.Split(path)
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
	cache, err := base.GetCache(path, account)
	if err == nil {
		files, _ := cache.([]model.File)
		return files, nil
	}
	files, err := driver.GetFiles(path, account)
	if err != nil {
		return nil, err
	}
	if len(files) > 0 {
		_ = base.SetCache(path, files, account)
	}
	return files, nil
}

func (driver Baidu) Link(args base.Args, account *model.Account) (*base.Link, error) {
	if account.InternalType == "crack" {
		return driver.LinkCrack(args, account)
	}
	return driver.LinkOfficial(args, account)
}

func (driver Baidu) LinkOfficial(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}
	if file.IsDir() {
		return nil, base.ErrNotFile
	}
	var resp DownloadResp
	params := map[string]string{
		"method": "filemetas",
		"fsids":  fmt.Sprintf("[%s]", file.Id),
		"dlink":  "1",
	}
	_, err = driver.Get("/xpan/multimedia", params, &resp, account)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s&access_token=%s", resp.List[0].Dlink, account.AccessToken)
	res, err := base.NoRedirectClient.R().SetHeader("User-Agent", "pan.baidu.com").Head(u)
	if err != nil {
		return nil, err
	}
	//if res.StatusCode() == 302 {
	u = res.Header().Get("location")
	//}
	return &base.Link{
		Url: u,
		Headers: []base.Header{
			{Name: "User-Agent", Value: "pan.baidu.com"},
		}}, nil
}

func (driver Baidu) LinkCrack(args base.Args, account *model.Account) (*base.Link, error) {
	file, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}
	if file.IsDir() {
		return nil, base.ErrNotFile
	}
	var resp DownloadResp2
	param := map[string]string{
		"target": fmt.Sprintf("[\"%s\"]", utils.Join(account.RootFolder, args.Path)),
		"dlink":  "1",
		"web":    "5",
		"origin": "dlna",
	}
	_, err = driver.Request("https://pan.baidu.com/api/filemetas", base.Get, nil, param, nil, nil, &resp, account)
	if err != nil {
		return nil, err
	}
	return &base.Link{
		Url: resp.Info[0].Dlink,
		Headers: []base.Header{
			{Name: "User-Agent", Value: "pan.baidu.com"},
		}}, nil
}

func (driver Baidu) Path(path string, account *model.Account) (*model.File, []model.File, error) {
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

//func (driver Baidu) Proxy(r *http.Request, account *model.Account) {
//	r.Header.Set("User-Agent", "pan.baidu.com")
//}

func (driver Baidu) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver Baidu) MakeDir(path string, account *model.Account) error {
	_, err := driver.create(utils.Join(account.RootFolder, path), 0, 1, "", "", account)
	return err
}

func (driver Baidu) Move(src string, dst string, account *model.Account) error {
	path := utils.Join(account.RootFolder, src)
	dest, newname := utils.Split(utils.Join(account.RootFolder, dst))
	data := []base.Json{
		{
			"path":    path,
			"dest":    dest,
			"newname": newname,
		},
	}
	_, err := driver.manage("move", data, account)
	return err
}

func (driver Baidu) Rename(src string, dst string, account *model.Account) error {
	path := utils.Join(account.RootFolder, src)
	newname := utils.Base(dst)
	data := []base.Json{
		{
			"path":    path,
			"newname": newname,
		},
	}
	_, err := driver.manage("rename", data, account)
	return err
}

func (driver Baidu) Copy(src string, dst string, account *model.Account) error {
	path := utils.Join(account.RootFolder, src)
	dest, newname := utils.Split(utils.Join(account.RootFolder, dst))
	data := []base.Json{
		{
			"path":    path,
			"dest":    dest,
			"newname": newname,
		},
	}
	_, err := driver.manage("copy", data, account)
	return err
}

func (driver Baidu) Delete(path string, account *model.Account) error {
	path = utils.Join(account.RootFolder, path)
	data := []string{path}
	_, err := driver.manage("delete", data, account)
	return err
}

func (driver Baidu) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	tempFile, err := ioutil.TempFile(conf.Conf.TempDir, "file-*")
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()
	_, err = io.Copy(tempFile, file)
	if err != nil {
		return err
	}
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	var Default uint64 = 4 * 1024 * 1024
	defaultByteData := make([]byte, Default)
	count := int(math.Ceil(float64(file.GetSize()) / float64(Default)))
	var SliceSize uint64 = 256 * 1024
	// cal md5
	h1 := md5.New()
	h2 := md5.New()
	block_list := make([]string, 0)
	content_md5 := ""
	slice_md5 := ""
	left := file.GetSize()
	for i := 0; i < count; i++ {
		byteSize := Default
		var byteData []byte
		if left < Default {
			byteSize = left
			byteData = make([]byte, byteSize)
		} else {
			byteData = defaultByteData
		}
		left -= byteSize
		_, err = io.ReadFull(tempFile, byteData)
		if err != nil {
			return err
		}
		h1.Write(byteData)
		h2.Write(byteData)
		block_list = append(block_list, fmt.Sprintf("\"%s\"", hex.EncodeToString(h2.Sum(nil))))
		h2.Reset()
	}
	content_md5 = hex.EncodeToString(h1.Sum(nil))
	_, err = tempFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}
	if file.GetSize() <= SliceSize {
		slice_md5 = content_md5
	} else {
		sliceData := make([]byte, SliceSize)
		_, err = io.ReadFull(tempFile, sliceData)
		if err != nil {
			return err
		}
		h2.Write(sliceData)
		slice_md5 = hex.EncodeToString(h2.Sum(nil))
		_, err = tempFile.Seek(0, io.SeekStart)
		if err != nil {
			return err
		}
	}
	path := encodeURIComponent(utils.Join(account.RootFolder, file.ParentPath, file.Name))
	block_list_str := fmt.Sprintf("[%s]", strings.Join(block_list, ","))
	data := fmt.Sprintf("path=%s&size=%d&isdir=0&autoinit=1&block_list=%s&content-md5=%s&slice-md5=%s",
		path, file.GetSize(),
		block_list_str,
		content_md5, slice_md5)
	params := map[string]string{
		"method": "precreate",
	}
	var precreateResp PrecreateResp
	_, err = driver.Post("/xpan/file", params, data, &precreateResp, account)
	if err != nil {
		return err
	}
	log.Debugf("%+v", precreateResp)
	if precreateResp.ReturnType == 2 {
		return nil
	}
	params = map[string]string{
		"method":       "upload",
		"access_token": account.AccessToken,
		"type":         "tmpfile",
		"path":         path,
		"uploadid":     precreateResp.Uploadid,
	}
	left = file.GetSize()
	for _, partseq := range precreateResp.BlockList {
		byteSize := Default
		var byteData []byte
		if left < Default {
			byteSize = left
			byteData = make([]byte, byteSize)
		} else {
			byteData = defaultByteData
		}
		left -= byteSize
		_, err = io.ReadFull(tempFile, byteData)
		if err != nil {
			return err
		}
		u := "https://d.pcs.baidu.com/rest/2.0/pcs/superfile2"
		params["partseq"] = strconv.Itoa(partseq)
		res, err := base.RestyClient.R().SetQueryParams(params).SetFileReader("file", file.Name, bytes.NewReader(byteData)).Post(u)
		if err != nil {
			return err
		}
		log.Debugln(res.String())
	}
	_, err = driver.create(path, file.GetSize(), 0, precreateResp.Uploadid, block_list_str, account)
	return err
}

var _ base.Driver = (*Baidu)(nil)

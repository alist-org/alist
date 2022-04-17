package native

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Native struct{}

func (driver Native) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:          "Native",
		OnlyProxy:     true,
		OnlyLocal:     true,
		NoNeedSetLink: true,
		LocalSort:     true,
	}
}

func (driver Native) Items() []base.Item {
	return []base.Item{
		{
			Name:     "root_folder",
			Label:    "root folder path",
			Type:     base.TypeString,
			Required: true,
		},
	}
}

func (driver Native) Save(account *model.Account, old *model.Account) error {
	if account == nil {
		return nil
	}
	log.Debugf("save a account: [%s]", account.Name)
	if !utils.Exists(account.RootFolder) {
		account.Status = fmt.Sprintf("[%s] not exist", account.RootFolder)
		_ = model.SaveAccount(account)
		return fmt.Errorf("[%s] not exist", account.RootFolder)
	}
	account.Status = "work"
	account.Proxy = true
	err := model.SaveAccount(account)
	if err != nil {
		return err
	}
	return nil
}

func (driver Native) File(path string, account *model.Account) (*model.File, error) {
	if utils.IsContain(strings.Split(path, "/"), "..") {
		return nil, base.ErrRelativePath
	}
	fullPath := filepath.Join(account.RootFolder, path)
	if !utils.Exists(fullPath) {
		return nil, base.ErrPathNotFound
	}
	f, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	time := f.ModTime()
	file := &model.File{
		Name:      f.Name(),
		UpdatedAt: &time,
		Driver:    driver.Config().Name,
	}
	if f.IsDir() {
		file.Type = conf.FOLDER
	} else {
		file.Type = utils.GetFileType(filepath.Ext(f.Name()))
		file.Size = f.Size()
	}
	return file, nil
}

func (driver Native) Files(path string, account *model.Account) ([]model.File, error) {
	if utils.IsContain(strings.Split(path, "/"), "..") {
		return nil, base.ErrRelativePath
	}
	fullPath := filepath.Join(account.RootFolder, path)
	if !utils.Exists(fullPath) {
		return nil, base.ErrPathNotFound
	}
	files := make([]model.File, 0)
	rawFiles, err := ioutil.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}
	for _, f := range rawFiles {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		time := f.ModTime()
		file := model.File{
			Name:      f.Name(),
			Type:      0,
			UpdatedAt: &time,
			Driver:    driver.Config().Name,
		}
		if f.IsDir() {
			file.Type = conf.FOLDER
		} else {
			file.Type = utils.GetFileType(filepath.Ext(f.Name()))
			file.Size = f.Size()
		}
		files = append(files, file)
	}
	_, err = base.GetCache(path, account)
	if len(files) != 0 && err != nil {
		_ = base.SetCache(path, files, account)
	}
	return files, nil
}

func (driver Native) Link(args base.Args, account *model.Account) (*base.Link, error) {
	_, err := driver.File(args.Path, account)
	if err != nil {
		return nil, err
	}
	fullPath := filepath.Join(account.RootFolder, args.Path)
	s, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	if s.IsDir() {
		return nil, base.ErrNotFile
	}
	link := base.Link{
		FilePath: fullPath,
	}
	return &link, nil
}

func (driver Native) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	log.Debugf("native path: %s", path)
	file, err := driver.File(path, account)
	if err != nil {
		return nil, nil, err
	}
	if !file.IsDir() {
		//file.Url, _ = driver.Link(path, account)
		return file, nil, nil
	}
	files, err := driver.Files(path, account)
	if err != nil {
		return nil, nil, err
	}
	//model.SortFiles(files, account)
	return nil, files, nil
}

//func (driver Native) Proxy(r *http.Request, account *model.Account) {
//	// unnecessary
//}

func (driver Native) Preview(path string, account *model.Account) (interface{}, error) {
	return nil, base.ErrNotSupport
}

func (driver Native) MakeDir(path string, account *model.Account) error {
	if utils.IsContain(strings.Split(path, "/"), "..") {
		return base.ErrRelativePath
	}
	fullPath := filepath.Join(account.RootFolder, path)
	err := os.MkdirAll(fullPath, 0700)
	return err
}

func (driver Native) Move(src string, dst string, account *model.Account) error {
	if utils.IsContain(strings.Split(src+"/"+dst, "/"), "..") {
		return base.ErrRelativePath
	}
	fullSrc := filepath.Join(account.RootFolder, src)
	fullDst := filepath.Join(account.RootFolder, dst)
	return os.Rename(fullSrc, fullDst)
}

func (driver Native) Rename(src string, dst string, account *model.Account) error {
	return driver.Move(src, dst, account)
}

func (driver Native) Copy(src string, dst string, account *model.Account) error {
	if utils.IsContain(strings.Split(src+"/"+dst, "/"), "..") {
		return base.ErrRelativePath
	}
	fullSrc := filepath.Join(account.RootFolder, src)
	fullDst := filepath.Join(account.RootFolder, dst)
	srcFile, err := driver.File(src, account)
	if err != nil {
		return err
	}
	dstFile, err := driver.File(dst, account)
	if err == nil {
		if !dstFile.IsDir() {
			return base.ErrNotSupport
		}
	}
	if srcFile.IsDir() {
		return driver.CopyDir(fullSrc, fullDst)
	}
	return driver.CopyFile(fullSrc, fullDst)
}

func (driver Native) Delete(path string, account *model.Account) error {
	if utils.IsContain(strings.Split(path, "/"), "..") {
		return base.ErrRelativePath
	}
	fullPath := filepath.Join(account.RootFolder, path)
	file, err := driver.File(path, account)
	if err != nil {
		return err
	}
	if file.IsDir() {
		return os.RemoveAll(fullPath)
	}
	return os.Remove(fullPath)
}

func (driver Native) Upload(file *model.FileStream, account *model.Account) error {
	if file == nil {
		return base.ErrEmptyFile
	}
	if utils.IsContain(strings.Split(file.ParentPath, "/"), "..") {
		return base.ErrRelativePath
	}
	fullPath := filepath.Join(account.RootFolder, file.ParentPath, file.Name)
	_, err := driver.File(filepath.Join(file.ParentPath, file.Name), account)
	if err == nil {
		// TODO overwrite?
	}
	basePath := filepath.Dir(fullPath)
	if !utils.Exists(basePath) {
		err := os.MkdirAll(basePath, 0744)
		if err != nil {
			return err
		}
	}
	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer func() {
		_ = out.Close()
	}()
	//var buf bytes.Buffer
	//reader := io.TeeReader(file, &buf)
	//h := md5.New()
	//_, err = io.Copy(h, reader)
	//if err != nil {
	//	return err
	//}
	//hash := hex.EncodeToString(h.Sum(nil))
	//log.Debugln("md5:", hash)
	//_, err = io.Copy(out, &buf)
	_, err = io.Copy(out, file)
	return err
}

var _ base.Driver = (*Native)(nil)

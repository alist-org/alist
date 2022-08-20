package template

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
)

type ChaoxingDrive struct {
	base.Base
}

func (driver ChaoxingDrive) Config() base.DriverConfig {
	return base.DriverConfig{
		Name:          "ChaoXing",
		OnlyProxy:     true,
		OnlyLocal:     false,
		ApiProxy:      false,
		NoNeedSetLink: false,
		NoCors:        false,
		LocalSort:     false,
	}
}

func (driver ChaoxingDrive) Items() []base.Item {
	// TODO fill need info
	return []base.Item{
		{
			Name:     "username",
			Label:    "手机号/超星号",
			Description: "填写您登录时所用的账号",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "password",
			Label:    "加密后的密码",
			Description: "请在超星网盘页面用控制台执行<br/> encryptByDES(\"你的登录密码\",\"u2oh6Vu^HWe40fj\") <br/>获取",
			Type:     base.TypeString,
			Required: true,
		},
		{
			Name:     "root_folder",
			Label:    "根目录id",
			Description:    "请自行研究，目前不保证本设置有效，建议留空",
			Type:     base.TypeString,
			Default:  "",
			Required: false,
		},
		{
			Name:     "limit",
			Label:    "目录文件上限",
			Description:    "请注意，超出上限的部分不会被加载",
			Type:     base.TypeNumber,
			Default:  "50",
			Required: false,
		},
	}
}

// Save 用户更新账号信息
func (driver ChaoxingDrive) Save(account *model.Account, old *model.Account) error {
	// TODO test available or init
	if old != nil {
		conf.Cron.Remove(cron.EntryID(old.CronId))
	}
	if account == nil {
		return nil
	}
	if account.Limit <= 0 {
		account.Limit = 50
	}

	// 登录网盘并取得 Enc 字符串 （ enc保存在 account.AccessSecret 中； cookie 以键值对的形式保存在 account.AccessToken 中）
	err := driver.GetEnc(account)
	if err != nil {
		return err
	}

	// 每隔一天重新获取一次 Enc(和
	cronId, err := conf.Cron.AddFunc("@every 24h", func() {
		id := account.ID
		log.Debugf("ali account id: %d", id)
		newAccount, err := model.GetAccountById(id)
		log.Debugf("ali account: %+v", newAccount)
		if err != nil {
			return
		}
		err = driver.GetEnc(newAccount)
		_ = model.SaveAccount(newAccount)
	})

	// 记录当前计划任务的id
	account.CronId = int(cronId)
	err = model.SaveAccount(account)

	if err != nil {
		return err
	}
	return nil
}

// File 通过用户路径获取到文件对象(主要是id号)
func (driver ChaoxingDrive) File(path string, account *model.Account) (*model.File, error) {
	path = utils.ParsePath(path)
	if path == "/" {
		return &model.File{
			Id:        "",
			Name:      account.Name,
			Size:      0,
			Type:      conf.FOLDER,
			Driver:    driver.Config().Name,
			UpdatedAt: account.UpdatedAt,
		}, nil
	}
	// 将路径分割成 父文件夹 和 文件名
	dir, name := filepath.Split(path)
	// 等同于访问上级目录
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

// Files 列出所有文件
func (driver ChaoxingDrive) Files(path string, account *model.Account) ([]model.File, error) {
	path = utils.ParsePath(path)
	// 先从缓存中获取结果
	cache, err := base.GetCache(path, account)
	var fileList []model.File
	if err == nil {
		// 缓存命中，将目录信息保存到变量中
		fileList = cache.([]model.File)
	} else {
		// 缓存未命中
		//尝试获取上级目录的id（递归地尝试从上级目录的缓存中读取信息，直到获取到根目录为止）
		file, err := driver.File(path, account)
		if err != nil {
			return nil, err
		}
		//列出上级目录的文件
		fileList, err = driver.ListFile(file.Id, account)
		if err != nil {
			return nil, err
		}
		// 缓存数据
		if len(fileList) > 0 {
			_ = base.SetCache(path, fileList, account)
		}
	}
	files := make([]model.File, 0)
	for _, file := range fileList {
		files = append(files, file)
	}
	return files, nil
}

var api_file_link = "https://pan-yz.chaoxing.com/download/downloadfile?fleid=%s&puid=1"

// Link 返回传入路径对应的文件的直链（本地除外），并包含需要携带的请求头
func (driver ChaoxingDrive) Link(args base.Args, account *model.Account) (*base.Link, error) {
	// TODO get file link
	file, e := driver.File(args.Path, account)
	if e != nil {
		return nil, e
	}
	url := fmt.Sprintf(api_file_link, file.Id[:strings.Index(file.Id, "_")])
	//var resp base.Json
	//var err Resp

	// https://pan-yz.chaoxing.com/download/downloadfile?fleid=582519780768600064&puid=1
	//_, e = chaoxingClient.R().SetResult(&resp).SetError(&err).
	//	SetHeader("Cookie", account.AccessToken).
	//	SetHeader("Referer", "https://pan-yz.chaoxing.com/").
	//	Get(url)
	return &base.Link{
		Headers: []base.Header{
			{
				Name:  "Referer",
				Value: "https://pan-yz.chaoxing.com/",
			},
			//{
			//	Name: "Cookie",
			//	Value: account.AccessToken,
			//},
		},
		Url: url,
	}, nil
}

// Path 通过调用上述的File与Files函数判断是文件还是文件夹，并进行返回，当是文件时附带文件的直链。
func (driver ChaoxingDrive) Path(path string, account *model.Account) (*model.File, []model.File, error) {
	path = utils.ParsePath(path)
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

// Optional function
//func (driver ChaoxingDrive) Preview(path string, account *model.Account) (interface{}, error) {
//	//TODO preview interface if driver support
//	return nil, base.ErrNotImplement
//}

func (driver ChaoxingDrive) splitPath(path string) (string,string){
	// path 展示在前端的路径，假定不以斜杠结尾。 用最后一个 '/' 可以将它分成两部分，根据前半部分查出父文件夹的id（应该已经包含puid），根据后半部分确定要创建的文件夹的名字
	// 规范路径
	path = utils.ParsePath(path)
	// 对路径进行切割
	sp := strings.LastIndex(path, "/")
	// 父文件夹路径
	parentDir := path[:sp]
	// 子文件夹路径
	childDir := path[sp+1:]
	return parentDir,childDir
}

func (driver ChaoxingDrive) MakeDir(path string, account *model.Account) error {
	parentDir, childDir := driver.splitPath(path)
	// 取得父目录的id（能来到这里说明前面 列出文件 的步骤没问题，此时必然可以再次列出文件，且系统中必然有缓存，因此不用考虑报错）
	pDir, _ := driver.File(parentDir, account)
	return driver.Mkdir(pDir.Id, childDir, account)
}

func (driver ChaoxingDrive) Move(src string, dst string, account *model.Account) error {
	// 注意 folderid 是由 {id}_{puid} 组成的(截取前两段即可)
	// https://pan-yz.chaoxing.com/opt/moveres?folderid=762268051373813760_54351295&resids=762263362701209600
	dstDir, _ := driver.splitPath(dst)
	srcFile, _ := driver.File(src, account)
	dstFolder, _ := driver.File(dstDir, account)
	return driver.Mv(srcFile.Id,dstFolder.Id,account)
}

func (driver ChaoxingDrive) Rename(src string, dst string, account *model.Account) error {
	// resid 就是 fileid
	// https://pan-yz.chaoxing.com/opt/rename?resid=762263362701209600&name=test.pdf&puid=54351295
	fmt.Printf(src,dst)
	srcFile, _ := driver.File(src, account)
	_, name := driver.splitPath(dst)
	return driver.Ren(srcFile.Id,name,account)
}

// 超星网盘不支持复制
//func (driver ChaoxingDrive) Copy(src string, dst string, account *model.Account) error {
//	//TODO copy file/dir
//	return base.ErrNotImplement
//}

// Delete 这个函数太危险了，不想实现
func (driver ChaoxingDrive) Delete(path string, account *model.Account) error {
	// 删除单个文件
	// https://pan-yz.chaoxing.com/opt/delres?resids=762268051373813760&resourcetype=0&puids=54351295
	// 删除多个文件
	// https://pan-yz.chaoxing.com/opt/delres?resids=762269933587513344,762269920078848000,&resourcetype=0,0,&puids=54351295,54351295,
	fmt.Printf(path)
	file, _ := driver.File(path, account)
	return driver.Rm(file.Id,account)
}

//func (driver ChaoxingDrive) Upload(file *model.FileStream, account *model.Account) error {
//	//TODO upload file
//	return base.ErrNotImplement
//}

var _ base.Driver = (*ChaoxingDrive)(nil)

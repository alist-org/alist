package model

import (
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/utils"
	log "github.com/sirupsen/logrus"
	"sort"
	"strings"
	"sync"
	"time"
)

type Account struct {
	ID             uint   `json:"id" gorm:"primaryKey"`                  // 唯一ID
	Name           string `json:"name" gorm:"unique" binding:"required"` // 唯一名称
	Index          int    `json:"index"`                                 // 序号 用于排序
	Type           string `json:"type"`                                  // 类型，即driver
	Username       string `json:"username"`
	Password       string `json:"password"`
	RefreshToken   string `json:"refresh_token"`
	AccessToken    string `json:"access_token"`
	RootFolder     string `json:"root_folder"`
	Status         string `json:"status"` // 状态
	CronId         int
	DriveId        string
	Limit          int        `json:"limit"`
	OrderBy        string     `json:"order_by"`
	OrderDirection string     `json:"order_direction"`
	UpdatedAt      *time.Time `json:"updated_at"`
	Search         bool       `json:"search"`
	ClientId       string     `json:"client_id"`
	ClientSecret   string     `json:"client_secret"`
	Zone           string     `json:"zone"`
	RedirectUri    string     `json:"redirect_uri"`
	SiteUrl        string     `json:"site_url"`
	SiteId         string     `json:"site_id"`
	InternalType   string     `json:"internal_type"`
	WebdavProxy    bool       `json:"webdav_proxy"`  // 开启之后只会webdav走中转
	Proxy          bool       `json:"proxy"`         // 是否中转,开启之后web和webdav都会走中转
	WebdavDirect   bool       `json:"webdav_direct"` // webdav 下载不跳转
	//AllowProxy     bool       `json:"allow_proxy"` // 是否允许中转下载
	DownProxyUrl string `json:"down_proxy_url"` // 用于中转下载服务的URL 两处 1. path请求中返回的链接 2. down下载时进行302
	APIProxyUrl  string `json:"api_proxy_url"`  // 用于中转api的地址
	// for s3
	Bucket        string `json:"bucket"`
	Endpoint      string `json:"endpoint"`
	Region        string `json:"region"`
	AccessKey     string `json:"access_key"`
	AccessSecret  string `json:"access_secret"`
	CustomHost    string `json:"custom_host"`
	ExtractFolder string `json:"extract_folder"`
	Bool1         bool   `json:"bool_1"`
	// for xunlei
	Algorithms    string `json:"algorithms"`
	ClientVersion string `json:"client_version"`
	PackageName   string `json:"package_name"`
	UserAgent    string `json:"user_agent"`
	CaptchaToken string `json:"captcha_token"`
	DeviceId     string `json:"device_id"`
}

var accountsMap = make(map[string]Account)

var balance = ".balance"

// SaveAccount save account to database
func SaveAccount(account *Account) error {
	if err := conf.DB.Save(account).Error; err != nil {
		return err
	}
	RegisterAccount(*account)
	return nil
}

func CreateAccount(account *Account) error {
	if err := conf.DB.Create(account).Error; err != nil {
		return err
	}
	RegisterAccount(*account)
	return nil
}

func DeleteAccount(id uint) (*Account, error) {
	var account Account
	account.ID = id
	if err := conf.DB.First(&account).Error; err != nil {
		return nil, err
	}
	name := account.Name
	if err := conf.DB.Delete(&account).Error; err != nil {
		return nil, err
	}
	delete(accountsMap, name)
	return &account, nil
}

func DeleteAccountFromMap(name string) {
	delete(accountsMap, name)
}

func AccountsCount() int {
	return len(accountsMap)
}

func RegisterAccount(account Account) {
	accountsMap[account.Name] = account
}

// GetAccount 根据名称获取账号（不包含负载均衡账号） 用于定时任务更新账号
func GetAccount(name string) (Account, bool) {
	if len(accountsMap) == 1 {
		for _, v := range accountsMap {
			return v, true
		}
	}
	account, ok := accountsMap[name]
	return account, ok
}

// GetAccountsByName 根据名称获取账号（包含负载均衡账号）
//func GetAccountsByName(name string) []Account {
//	accounts := make([]Account, 0)
//	if AccountsCount() == 1 {
//		for _, v := range accountsMap {
//			accounts = append(accounts, v)
//		}
//		return accounts
//	}
//	for _, v := range accountsMap {
//		if v.Name == name || strings.HasPrefix(v.Name, name+balance) {
//			accounts = append(accounts, v)
//		}
//	}
//	return accounts
//}

var balanceMap sync.Map

// GetBalancedAccount 根据名称获取账号，负载均衡之后的
func GetBalancedAccount(name string) (Account, bool) {
	accounts := GetAccountsByPath(name)
	log.Debugf("accounts: %+v", accounts)
	accountNum := len(accounts)
	switch accountNum {
	case 0:
		return Account{}, false
	case 1:
		return accounts[0], true
	default:
		cur, ok := balanceMap.Load(name)
		if ok {
			i := cur.(int)
			i = (i + 1) % accountNum
			balanceMap.Store(name, i)
			log.Debugln("use: ", i)
			return accounts[i], true
		} else {
			balanceMap.Store(name, 0)
			return accounts[0], true
		}
	}
}

// GetAccountById 根据id获取账号，用于更新账号
func GetAccountById(id uint) (*Account, error) {
	var account Account
	account.ID = id
	if err := conf.DB.First(&account).Error; err != nil {
		return nil, err
	}
	return &account, nil
}

// GetAccountFiles 获取账号虚拟文件（去除负载均衡）
//func GetAccountFiles() ([]File, error) {
//	files := make([]File, 0)
//	var accounts []Account
//	if err := conf.DB.Order(columnName("index")).Find(&accounts).Error; err != nil {
//		return nil, err
//	}
//	for _, v := range accounts {
//		if strings.Contains(v.Name, balance) {
//			continue
//		}
//		files = append(files, File{
//			Name:      v.Name,
//			Size:      0,
//			Driver:    v.Type,
//			Type:      conf.FOLDER,
//			UpdatedAt: v.UpdatedAt,
//		})
//	}
//	return files, nil
//}

// GetAccounts 获取所有账号
func GetAccounts() ([]Account, error) {
	var accounts []Account
	if err := conf.DB.Order(columnName("index")).Find(&accounts).Error; err != nil {
		return nil, err
	}
	return accounts, nil
}

// GetAccountsByPath 根据路径获取账号，最长匹配，未负载均衡
// 如有账号： /a/b,/a/c,/a/d/e,/a/d/e.balance
// GetAccountsByPath(/a/d/e/f) => /a/d/e,/a/d/e.balance
func GetAccountsByPath(path string) []Account {
	accounts := make([]Account, 0)
	curSlashCount := 0
	for _, v := range accountsMap {
		name := utils.ParsePath(v.Name)
		bIndex := strings.LastIndex(name, balance)
		if bIndex != -1 {
			name = name[:bIndex]
		}
		if name == "/" {
			name = ""
		}
		// 不是这个账号
		if path != name && !strings.HasPrefix(path, name+"/") {
			continue
		}
		slashCount := strings.Count(name, "/")
		// 不是最长匹配
		if slashCount < curSlashCount {
			continue
		}
		if slashCount > curSlashCount {
			accounts = accounts[:0]
			curSlashCount = slashCount
		}
		accounts = append(accounts, v)
	}
	sort.Slice(accounts, func(i, j int) bool {
		return accounts[i].Name < accounts[j].Name
	})
	return accounts
}

// GetAccountFilesByPath 根据路径获取账号虚拟文件
// 如有账号： /a/b,/a/c,/a/d/e,/a/b.balance1,/av
// GetAccountFilesByPath(/a) => b,c,d
func GetAccountFilesByPath(prefix string) []File {
	files := make([]File, 0)
	accounts := make([]Account, AccountsCount())
	i := 0
	for _, v := range accountsMap {
		accounts[i] = v
		i += 1
	}
	sort.Slice(accounts, func(i, j int) bool {
		if accounts[i].Index == accounts[j].Index {
			return accounts[i].Name < accounts[j].Name
		}
		return accounts[i].Index < accounts[j].Index
	})
	prefix = utils.ParsePath(prefix)
	set := make(map[string]interface{})
	for _, v := range accounts {
		// 负载均衡账号
		if strings.Contains(v.Name, balance) {
			continue
		}
		full := utils.ParsePath(v.Name)
		if len(full) <= len(prefix) {
			continue
		}
		// 不是以prefix为前缀
		if !strings.HasPrefix(full, prefix+"/") && prefix != "/" {
			continue
		}
		name := strings.Split(strings.TrimPrefix(strings.TrimPrefix(full, prefix), "/"), "/")[0]
		if _, ok := set[name]; ok {
			continue
		}
		files = append(files, File{
			Name:      name,
			Size:      0,
			Driver:    v.Type,
			Type:      conf.FOLDER,
			UpdatedAt: v.UpdatedAt,
		})
		set[name] = nil
	}
	return files
}

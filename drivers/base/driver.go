package base

import (
	"github.com/Xhofe/alist/model"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type DriverConfig struct {
	Name        string
	OnlyProxy   bool
	NoLink      bool // 必须本机返回的
	ApiProxy    bool // 使用API中转的
	NeedSetLink bool // 需要设置链接的
}

type Args struct {
	Path string
	IP   string
}

type Driver interface {
	// Config 配置
	Config() DriverConfig
	// Items 账号所需参数
	Items() []Item
	// Save 保存时处理
	Save(account *model.Account, old *model.Account) error
	// File 取文件
	File(path string, account *model.Account) (*model.File, error)
	// Files 取文件夹
	Files(path string, account *model.Account) ([]model.File, error)
	// Link 取链接
	Link(args Args, account *model.Account) (*Link, error)
	// Path 取路径（文件或文件夹）
	Path(path string, account *model.Account) (*model.File, []model.File, error)
	// Proxy 代理处理
	Proxy(c *gin.Context, account *model.Account)
	// Preview 预览
	Preview(path string, account *model.Account) (interface{}, error)
	// MakeDir 创建文件夹
	MakeDir(path string, account *model.Account) error
	// Move 移动/改名
	Move(src string, dst string, account *model.Account) error
	// Copy 拷贝
	Copy(src string, dst string, account *model.Account) error
	// Delete 删除
	Delete(path string, account *model.Account) error
	// Upload 上传
	Upload(file *model.FileStream, account *model.Account) error
	// TODO
	//Search(path string, keyword string, account *model.Account) ([]*model.File, error)
}

type Item struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Values      string `json:"values"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

var driversMap = map[string]Driver{}

func RegisterDriver(driver Driver) {
	log.Infof("register driver: [%s]", driver.Config().Name)
	driversMap[driver.Config().Name] = driver
}

func GetDriver(name string) (driver Driver, ok bool) {
	driver, ok = driversMap[name]
	return
}

func GetDrivers() map[string][]Item {
	res := make(map[string][]Item, 0)
	for k, v := range driversMap {
		if v.Config().OnlyProxy {
			res[k] = v.Items()
		} else {
			res[k] = append([]Item{
				//{
				//	Name:        "allow_proxy",
				//	Label:       "allow_proxy",
				//	Type:        TypeBool,
				//	Required:    true,
				//	Description: "allow proxy",
				//},
				{
					Name:        "proxy",
					Label:       "proxy",
					Type:        TypeBool,
					Required:    true,
					Description: "web proxy",
				},
				{
					Name:        "webdav_proxy",
					Label:       "webdav proxy",
					Type:        TypeBool,
					Required:    true,
					Description: "Transfer the WebDAV of this account through the server",
				},
			}, v.Items()...)
		}
		res[k] = append([]Item{
			{
				Name:  "down_proxy_url",
				Label: "down_proxy_url",
				Type:  TypeString,
			},
		}, res[k]...)
		if v.Config().ApiProxy {
			res[k] = append([]Item{
				{
					Name:  "api_proxy_url",
					Label: "api_proxy_url",
					Type:  TypeString,
				},
			}, res[k]...)
		}
	}
	return res
}

var NoRedirectClient *resty.Client
var RestyClient = resty.New()
var HttpClient = &http.Client{}

func init() {
	NoRedirectClient = resty.New().SetRedirectPolicy(
		resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}),
	)
	userAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"
	NoRedirectClient.SetHeader("user-agent", userAgent)
	RestyClient.SetHeader("user-agent", userAgent)
	RestyClient.SetRetryCount(3)
}

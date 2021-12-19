package base

import (
	"github.com/Xhofe/alist/model"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type DriverConfig struct {
	Name      string
	OnlyProxy bool
	NoLink    bool // 必须本机返回的
}

type Driver interface {
	Config() DriverConfig
	Items() []Item
	Save(account *model.Account, old *model.Account) error
	File(path string, account *model.Account) (*model.File, error)
	Files(path string, account *model.Account) ([]model.File, error)
	Link(path string, account *model.Account) (*Link, error)
	Path(path string, account *model.Account) (*model.File, []model.File, error)
	Proxy(c *gin.Context, account *model.Account)
	Preview(path string, account *model.Account) (interface{}, error)
	// TODO
	//Search(path string, keyword string, account *model.Account) ([]*model.File, error)
	MakeDir(path string, account *model.Account) error
	Move(src string, dst string, account *model.Account) error
	Copy(src string, dst string, account *model.Account) error
	Delete(path string, account *model.Account) error
	Upload(file *model.FileStream, account *model.Account) error
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
			{
				Name:  "api_proxy_url",
				Label: "api_proxy_url",
				Type:  TypeString,
			},
		}, res[k]...)
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

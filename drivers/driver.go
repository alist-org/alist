package drivers

import (
	"github.com/Xhofe/alist/model"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
	"net/http"
)

type Driver interface {
	Items() []Item
	Save(account *model.Account, old *model.Account) error
	File(path string, account *model.Account) (*model.File, error)
	Files(path string, account *model.Account) ([]model.File, error)
	Link(path string, account *model.Account) (string, error)
	Path(path string, account *model.Account) (*model.File, []model.File, error)
	Proxy(c *gin.Context, account *model.Account)
	Preview(path string, account *model.Account) (interface{}, error)
	// TODO
	//Search(path string, keyword string, account *model.Account) ([]*model.File, error)
	//MakeDir(path string, account *model.Account) error
	//Move(src string, des string, account *model.Account) error
	//Delete(path string) error
	//Upload(file *fs.File, path string, account *model.Account) error
}

type Item struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Values      string `json:"values"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type TokenResp struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

var driversMap = map[string]Driver{}

func RegisterDriver(name string, driver Driver) {
	log.Infof("register driver: [%s]", name)
	driversMap[name] = driver
}

func GetDriver(name string) (driver Driver, ok bool) {
	driver, ok = driversMap[name]
	return
}

func GetDrivers() map[string][]Item {
	res := make(map[string][]Item, 0)
	for k, v := range driversMap {
		res[k] = v.Items()
	}
	return res
}

type Json map[string]interface{}

var NoRedirectClient *resty.Client

func init() {
	NoRedirectClient = resty.New().SetRedirectPolicy(
		resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}),
	)
	NoRedirectClient.SetHeader("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
}

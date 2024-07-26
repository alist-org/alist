package lanzou

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Type string `json:"type" type:"select" options:"account,cookie,url" default:"cookie"`

	Account  string `json:"account"`
	Password string `json:"password"`

	Cookie string `json:"cookie" help:"about 15 days valid, ignore if shareUrl is used"`

	driver.RootID
	SharePassword  string `json:"share_password"`
	BaseUrl        string `json:"baseUrl" required:"true" default:"https://pc.woozooo.com" help:"basic URL for file operation"`
	ShareUrl       string `json:"shareUrl" required:"true" default:"https://pan.lanzoui.com" help:"used to get the sharing page"`
	UserAgent      string `json:"user_agent" required:"true" default:"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.39 (KHTML, like Gecko) Chrome/89.0.4389.111 Safari/537.39"`
	RepairFileInfo bool   `json:"repair_file_info" help:"To use webdav, you need to enable it"`
}

func (a *Addition) IsCookie() bool {
	return a.Type == "cookie"
}

func (a *Addition) IsAccount() bool {
	return a.Type == "account"
}

var config = driver.Config{
	Name:        "Lanzou",
	LocalSort:   true,
	DefaultRoot: "-1",
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &LanZou{}
	})
}

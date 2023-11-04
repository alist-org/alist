package _115

import (
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/op"
)

type Addition struct {
	Cookie      string  `json:"cookie" type:"text" help:"one of QR code token and cookie required"`
	QRCodeToken string  `json:"qrcode_token" type:"text" help:"one of QR code token and cookie required"`
	PageSize    int64   `json:"page_size" type:"number" default:"56" help:"list api per page size of 115 driver"`
	LimitRate   float64 `json:"limit_rate" type:"number" default:"2" help:"limit all api request rate (1r/[limit_rate]s)"`
	ShareCode   string  `json:"share_code" type:"text" help:"if the mounted directory is a shared one, a share code is required to support direct link downloads."`
	ReceiveCode string  `json:"receive_code" type:"text" help:"if the mounted directory is a shared one, a receive code is required to support direct link downloads."`
	driver.RootID
}

var config = driver.Config{
	Name:        "115 Cloud",
	DefaultRoot: "0",
	// OnlyProxy:   true,
	// OnlyLocal:         true,
	CheckStatus:       false,
	Alert:             "",
	NoOverwriteUpload: true,
}

func init() {
	op.RegisterDriver(func() driver.Driver {
		return &Pan115{}
	})
}

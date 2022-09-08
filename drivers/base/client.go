package base

import (
	"net/http"
	"time"

	"github.com/go-resty/resty/v2"
)

var NoRedirectClient *resty.Client
var RestyClient = NewRestyClient()
var HttpClient = &http.Client{}
var UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"
var DefaultTimeout = time.Second * 10

func init() {
	NoRedirectClient = resty.New().SetRedirectPolicy(
		resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}),
	)
	NoRedirectClient.SetHeader("user-agent", UserAgent)
}

func NewRestyClient() *resty.Client {
	return resty.New().
		SetHeader("user-agent", UserAgent).
		SetRetryCount(3).
		SetTimeout(DefaultTimeout)
}

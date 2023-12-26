package seafile

import (
	"fmt"
	"strings"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/go-resty/resty/v2"
)

func (d *Seafile) getToken() error {
	var authResp AuthTokenResp
	res, err := base.RestyClient.R().
		SetResult(&authResp).
		SetFormData(map[string]string{
			"username": d.UserName,
			"password": d.Password,
		}).
		Post(d.Address + "/api2/auth-token/")
	if err != nil {
		return err
	}
	if res.StatusCode() >= 400 {
		return fmt.Errorf("get token failed: %s", res.String())
	}
	d.authorization = fmt.Sprintf("Token %s", authResp.Token)
	return nil
}

func (d *Seafile) request(method string, pathname string, callback base.ReqCallback, noRedirect ...bool) ([]byte, error) {
	full := pathname
	if !strings.HasPrefix(pathname, "http") {
		full = d.Address + pathname
	}
	req := base.RestyClient.R()
	if len(noRedirect) > 0 && noRedirect[0] {
		req = base.NoRedirectClient.R()
	}
	req.SetHeader("Authorization", d.authorization)
	callback(req)
	var (
		res *resty.Response
		err error
	)
	for i := 0; i < 2; i++ {
		res, err = req.Execute(method, full)
		if err != nil {
			return nil, err
		}
		if res.StatusCode() != 401 { // Unauthorized
			break
		}
		err = d.getToken()
		if err != nil {
			return nil, err
		}
	}
	if res.StatusCode() >= 400 {
		return nil, fmt.Errorf("request failed: %s", res.String())
	}
	return res.Body(), nil
}

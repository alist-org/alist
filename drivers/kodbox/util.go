package kodbox

import (
	"fmt"
	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	"strings"
)

func (d *KodBox) getToken() error {
	var authResp CommonResp
	res, err := base.RestyClient.R().
		SetResult(&authResp).
		SetQueryParams(map[string]string{
			"name":     d.UserName,
			"password": d.Password,
		}).
		Post(d.Address + "/?user/index/loginSubmit")
	if err != nil {
		return err
	}
	if res.StatusCode() >= 400 {
		return fmt.Errorf("get token failed: %s", res.String())
	}

	if res.StatusCode() == 200 && authResp.Code.(bool) == false {
		return fmt.Errorf("get token failed: %s", res.String())
	}

	d.authorization = fmt.Sprintf("%s", authResp.Info)
	return nil
}

func (d *KodBox) request(method string, pathname string, callback base.ReqCallback, noRedirect ...bool) ([]byte, error) {
	full := pathname
	if !strings.HasPrefix(pathname, "http") {
		full = d.Address + pathname
	}
	req := base.RestyClient.R()
	if len(noRedirect) > 0 && noRedirect[0] {
		req = base.NoRedirectClient.R()
	}
	req.SetFormData(map[string]string{
		"accessToken": d.authorization,
	})
	callback(req)

	var (
		res        *resty.Response
		commonResp *CommonResp
		err        error
		skip       bool
	)
	for i := 0; i < 2; i++ {
		if skip {
			break
		}
		res, err = req.Execute(method, full)
		if err != nil {
			return nil, err
		}

		err := utils.Json.Unmarshal(res.Body(), &commonResp)
		if err != nil {
			return nil, err
		}

		switch commonResp.Code.(type) {
		case bool:
			skip = true
		case string:
			if commonResp.Code.(string) == "10001" {
				err = d.getToken()
				if err != nil {
					return nil, err
				}
				req.SetFormData(map[string]string{"accessToken": d.authorization})
			}
		}
	}
	if commonResp.Code.(bool) == false {
		return nil, fmt.Errorf("request failed: %s", commonResp.Data)
	}
	return res.Body(), nil
}

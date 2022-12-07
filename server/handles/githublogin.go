package handles

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

func Loginredirect(c *gin.Context) {
	method := c.Query("method")
	callbackurl := c.Query("callback_url")
	with_params := c.Query("with_params")
	enabled, err := db.GetSettingItemByKey("github_login_enabled")
	client_id, err := db.GetSettingItemByKey("github_client_id")
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	} else if enabled.Value == "true" {
		urlvalues := url.Values{}
		urlvalues.Add("client_id", client_id.Value)
		if method == "get_github_id" {
			urlvalues.Add("allow_signup", "true")
		} else if method == "github_callback_login" {
			urlvalues.Add("allow_signup", "false")
		}
		if method == "" {
			common.ErrorResp(c, errors.New("No method provided"), 400)
			return
		}
		if with_params != "" {
			urlvalues.Add("redirect_uri", common.GetApiUrl(c.Request)+"/api/auth/github_callback"+"?method="+method+"&callback_url="+callbackurl+"&with_params="+with_params)
		} else {
			urlvalues.Add("redirect_uri", common.GetApiUrl(c.Request)+"/api/auth/github_callback"+"?method="+method+"&callback_url="+callbackurl)
		}

		c.Redirect(302, "https://github.com/login/oauth/authorize?"+urlvalues.Encode())
	} else {
		common.ErrorResp(c, errors.New("Github Signin not enabled"), 403)
	}
	return
}

func GithubCallback(c *gin.Context) {
	argument := c.Query("method")
	callback_url := c.Query("callback_url")
	if argument == "get_github_id" || argument == "github_login" {
		enabled, err := db.GetSettingItemByKey("github_login_enabled")
		client_id, err := db.GetSettingItemByKey("github_client_id")
		client_secret, err := db.GetSettingItemByKey("github_client_secrets")
		if err != nil {
			common.ErrorResp(c, err, 400)
			return
		} else if enabled.Value == "true" {
			callbackcode := c.Query("code")
			if callbackcode == "" {
				common.ErrorResp(c, errors.New("No code provided"), 400)
				return
			}
			urlvalues := url.Values{}
			urlvalues.Add("client_id", client_id.Value)
			urlvalues.Add("client_secret", client_secret.Value)
			urlvalues.Add("code", callbackcode)
			urlvalues.Add("redirect_uri", common.GetApiUrl(c.Request)+"/api/auth/github_callback")
			resp, err := http.Post("https://github.com/login/oauth/access_token?"+urlvalues.Encode(), "application/json", nil)
			if err != nil {
				common.ErrorResp(c, err, 400)
				return
			}
			response, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				common.ErrorResp(c, err, 400)
				return
			}
			decodedValue, err := url.ParseQuery(string(response))
			if err != nil {
				common.ErrorResp(c, err, 400)
				return
			}
			access_token := decodedValue.Get("access_token")
			client := &http.Client{}
			resp, err = client.Get("https://api.github.com/user")
			req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
			req.Header.Add("Authorization", "Bearer "+access_token)
			resp, err = client.Do(req)
			if err != nil {
				common.ErrorResp(c, err, 400)
				return
			} else {
				body, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					common.ErrorResp(c, err, 400)
					return
				}
				user_id := utils.Json.Get(body, "id").ToString()
				if argument == "get_github_id" {
					c.Redirect(302, callback_url+"?callback-id="+user_id)
				}
				if argument == "github_login" {
					useridint, err := strconv.Atoi(user_id)
					if err != nil {
						common.ErrorResp(c, err, 400)
						return
					}
					user, err := db.GetUserByGithubID(useridint)
					if err != nil {
						common.ErrorResp(c, err, 400)
					}
					token, err := common.GenerateToken(user.Username)
					with_params := c.Query("with_params")
					if with_params == "true" {
						c.Redirect(302, callback_url+"&token="+token)
					} else if with_params == "false" {
						c.Redirect(302, callback_url+"?token="+token)
					}
					return
				}
			}
		} else {
			common.ErrorResp(c, errors.New("Invalid Request"), 500)
		}
		return
	}
}

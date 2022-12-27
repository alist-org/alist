package handles

import (
	"errors"
	"net/url"
	"strconv"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

func GithubLoginRedirect(c *gin.Context) {
	method := c.Query("method")
	callbackURL := c.Query("callback_url")
	withParams := c.Query("with_params")
	enabled, err := db.GetSettingItemByKey("github_login_enabled")
	clientId, err := db.GetSettingItemByKey("github_client_id")
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	} else if enabled.Value == "true" {
		urlValues := url.Values{}
		urlValues.Add("client_id", clientId.Value)
		if method == "get_github_id" {
			urlValues.Add("allow_signup", "true")
		} else if method == "github_callback_login" {
			urlValues.Add("allow_signup", "false")
		}
		if method == "" {
			common.ErrorStrResp(c, "no method provided", 400)
			return
		}
		if withParams != "" {
			urlValues.Add("redirect_uri", common.GetApiUrl(c.Request)+"/api/auth/github_callback"+"?method="+method+"&callback_url="+callbackURL+"&with_params="+withParams)
		} else {
			urlValues.Add("redirect_uri", common.GetApiUrl(c.Request)+"/api/auth/github_callback"+"?method="+method+"&callback_url="+callbackURL)
		}
		c.Redirect(302, "https://github.com/login/oauth/authorize?"+urlValues.Encode())
	} else {
		common.ErrorStrResp(c, "github Login not enabled", 403)
	}
}

var githubClient = resty.New().SetRetryCount(3)

func GithubLoginCallback(c *gin.Context) {
	argument := c.Query("method")
	callbackUrl := c.Query("callback_url")
	if argument == "get_github_id" || argument == "github_login" {
		enabled, err := db.GetSettingItemByKey("github_login_enabled")
		clientId, err := db.GetSettingItemByKey("github_client_id")
		clientSecret, err := db.GetSettingItemByKey("github_client_secrets")
		if err != nil {
			common.ErrorResp(c, err, 400)
			return
		} else if enabled.Value == "true" {
			callbackCode := c.Query("code")
			if callbackCode == "" {
				common.ErrorStrResp(c, "No code provided", 400)
				return
			}
			resp, err := githubClient.R().SetHeader("content-type", "application/json").
				SetBody(map[string]string{
					"client_id":     clientId.Value,
					"client_secret": clientSecret.Value,
					"code":          callbackCode,
					"redirect_uri":  common.GetApiUrl(c.Request) + "/api/auth/github_callback",
				}).Post("https://github.com/login/oauth/access_token")
			if err != nil {
				common.ErrorResp(c, err, 400)
				return
			}
			accessToken := utils.Json.Get(resp.Body(), "access_token").ToString()
			resp, err = githubClient.R().SetHeader("Authorization", "Bearer "+accessToken).
				Get("https://api.github.com/user")
			ghUserID := utils.Json.Get(resp.Body(), "id").ToInt()
			if argument == "get_github_id" {
				c.Redirect(302, callbackUrl+"?githubID="+strconv.Itoa(ghUserID))
			}
			if argument == "github_login" {
				user, err := db.GetUserByGithubID(ghUserID)
				if err != nil {
					common.ErrorResp(c, err, 400)
				}
				token, err := common.GenerateToken(user.Username)
				withParams := c.Query("with_params")
				if withParams == "true" {
					c.Redirect(302, callbackUrl+"&token="+token)
				} else if withParams == "false" {
					c.Redirect(302, callbackUrl+"?token="+token)
				}
				return
			}
		} else {
			common.ErrorResp(c, errors.New("invalid request"), 500)
		}
	}
}

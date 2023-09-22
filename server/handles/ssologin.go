package handles

import (
	"encoding/base32"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/pkg/utils/random"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/coreos/go-oidc"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

var opts = totp.ValidateOpts{
	// state verify won't expire in 30 secs, which is quite enough for the callback
	Period: 30,
	Skew:   1,
	// in some OIDC providers(such as Authelia), state parameter must be at least 8 characters
	Digits:    otp.DigitsEight,
	Algorithm: otp.AlgorithmSHA1,
}

func SSOLoginRedirect(c *gin.Context) {
	method := c.Query("method")
	usecompatibility := setting.GetBool(conf.SSOCompatibilityMode)
	enabled := setting.GetBool(conf.SSOLoginEnabled)
	clientId := setting.GetStr(conf.SSOClientId)
	platform := setting.GetStr(conf.SSOLoginPlatform)
	var r_url string
	var redirect_uri string
	if !enabled {
		common.ErrorStrResp(c, "Single sign-on is not enabled", 403)
		return
	}
	urlValues := url.Values{}
	if method == "" {
		common.ErrorStrResp(c, "no method provided", 400)
		return
	}
	if usecompatibility {
		redirect_uri = common.GetApiUrl(c.Request) + "/api/auth/" + method
	} else {
		redirect_uri = common.GetApiUrl(c.Request) + "/api/auth/sso_callback" + "?method=" + method
	}
	urlValues.Add("response_type", "code")
	urlValues.Add("redirect_uri", redirect_uri)
	urlValues.Add("client_id", clientId)
	switch platform {
	case "Github":
		r_url = "https://github.com/login/oauth/authorize?"
		urlValues.Add("scope", "read:user")
	case "Microsoft":
		r_url = "https://login.microsoftonline.com/common/oauth2/v2.0/authorize?"
		urlValues.Add("scope", "user.read")
		urlValues.Add("response_mode", "query")
	case "Google":
		r_url = "https://accounts.google.com/o/oauth2/v2/auth?"
		urlValues.Add("scope", "https://www.googleapis.com/auth/userinfo.profile")
	case "Dingtalk":
		r_url = "https://login.dingtalk.com/oauth2/auth?"
		urlValues.Add("scope", "openid")
		urlValues.Add("prompt", "consent")
		urlValues.Add("response_type", "code")
	case "Casdoor":
		endpoint := strings.TrimSuffix(setting.GetStr(conf.SSOEndpointName), "/")
		r_url = endpoint + "/login/oauth/authorize?"
		urlValues.Add("scope", "profile")
		urlValues.Add("state", endpoint)
	case "OIDC":
		oauth2Config, err := GetOIDCClient(c)
		if err != nil {
			common.ErrorStrResp(c, err.Error(), 400)
			return
		}
		// generate state parameter
		state, err := totp.GenerateCodeCustom(base32.StdEncoding.EncodeToString([]byte(oauth2Config.ClientSecret)), time.Now(), opts)
		if err != nil {
			common.ErrorStrResp(c, err.Error(), 400)
			return
		}
		c.Redirect(http.StatusFound, oauth2Config.AuthCodeURL(state))
		return
	default:
		common.ErrorStrResp(c, "invalid platform", 400)
		return
	}
	c.Redirect(302, r_url+urlValues.Encode())
}

var ssoClient = resty.New().SetRetryCount(3)

func GetOIDCClient(c *gin.Context) (*oauth2.Config, error) {
	var redirect_uri string
	usecompatibility := setting.GetBool(conf.SSOCompatibilityMode)
	argument := c.Query("method")
	if usecompatibility {
		argument = path.Base(c.Request.URL.Path)
	}
	if usecompatibility {
		redirect_uri = common.GetApiUrl(c.Request) + "/api/auth/" + argument
	} else {
		redirect_uri = common.GetApiUrl(c.Request) + "/api/auth/sso_callback" + "?method=" + argument
	}
	endpoint := setting.GetStr(conf.SSOEndpointName)
	provider, err := oidc.NewProvider(c, endpoint)
	if err != nil {
		return nil, err
	}
	clientId := setting.GetStr(conf.SSOClientId)
	clientSecret := setting.GetStr(conf.SSOClientSecret)
	return &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirect_uri,

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		Scopes: []string{oidc.ScopeOpenID, "profile"},
	}, nil
}

func autoRegister(username, userID string, err error) (*model.User, error) {
	if !errors.Is(err, gorm.ErrRecordNotFound) || !setting.GetBool(conf.SSOAutoRegister) {
		return nil, err
	}
	if username == "" {
		return nil, errors.New("cannot get username from SSO provider")
	}
	user := &model.User{
		ID:         0,
		Username:   username,
		Password:   random.String(16),
		Permission: int32(setting.GetInt(conf.SSODefaultPermission, 0)),
		BasePath:   setting.GetStr(conf.SSODefaultDir),
		Role:       0,
		Disabled:   false,
		SsoID:      userID,
	}
	if err = db.CreateUser(user); err != nil {
		if strings.HasPrefix(err.Error(), "UNIQUE constraint failed") && strings.HasSuffix(err.Error(), "username") {
			user.Username = user.Username + "_" + userID
			if err = db.CreateUser(user); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return user, nil
}

func OIDCLoginCallback(c *gin.Context) {
	usecompatibility := setting.GetBool(conf.SSOCompatibilityMode)
	argument := c.Query("method")
	if usecompatibility {
		argument = path.Base(c.Request.URL.Path)
	}
	clientId := setting.GetStr(conf.SSOClientId)
	endpoint := setting.GetStr(conf.SSOEndpointName)
	provider, err := oidc.NewProvider(c, endpoint)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	oauth2Config, err := GetOIDCClient(c)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	// add state verify process
	stateVerification, err := totp.ValidateCustom(c.Query("state"), base32.StdEncoding.EncodeToString([]byte(oauth2Config.ClientSecret)), time.Now(), opts)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if !stateVerification {
		common.ErrorStrResp(c, "incorrect or expired state parameter", 400)
		return
	}

	oauth2Token, err := oauth2Config.Exchange(c, c.Query("code"))
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		common.ErrorStrResp(c, "no id_token found in oauth2 token", 400)
		return
	}
	verifier := provider.Verifier(&oidc.Config{
		ClientID: clientId,
	})
	idToken, err := verifier.Verify(c, rawIDToken)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	type UserInfo struct {
		Name string `json:"name"`
	}
	claims := UserInfo{}
	if err := idToken.Claims(&claims); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	UserID := claims.Name
	if argument == "get_sso_id" {
		if usecompatibility {
			c.Redirect(302, common.GetApiUrl(c.Request)+"/@manage?sso_id="+UserID)
			return
		}
		html := fmt.Sprintf(`<!DOCTYPE html>
				<head></head>
				<body>
				<script>
				window.opener.postMessage({"sso_id": "%s"}, "*")
				window.close()
				</script>
				</body>`, UserID)
		c.Data(200, "text/html; charset=utf-8", []byte(html))
		return
	}
	if argument == "sso_get_token" {
		user, err := db.GetUserBySSOID(UserID)
		if err != nil {
			user, err = autoRegister(UserID, UserID, err)
			if err != nil {
				common.ErrorResp(c, err, 400)
			}
		}
		token, err := common.GenerateToken(user.Username)
		if err != nil {
			common.ErrorResp(c, err, 400)
		}
		if usecompatibility {
			c.Redirect(302, common.GetApiUrl(c.Request)+"/@login?token="+token)
			return
		}
		html := fmt.Sprintf(`<!DOCTYPE html>
				<head></head>
				<body>
				<script>
				window.opener.postMessage({"token":"%s"}, "*")
				window.close()
				</script>
				</body>`, token)
		c.Data(200, "text/html; charset=utf-8", []byte(html))
		return
	}
}

func SSOLoginCallback(c *gin.Context) {
	enabled := setting.GetBool(conf.SSOLoginEnabled)
	usecompatibility := setting.GetBool(conf.SSOCompatibilityMode)
	if !enabled {
		common.ErrorResp(c, errors.New("sso login is disabled"), 500)
		return
	}
	argument := c.Query("method")
	if usecompatibility {
		argument = path.Base(c.Request.URL.Path)
	}
	if !utils.SliceContains([]string{"get_sso_id", "sso_get_token"}, argument) {
		common.ErrorResp(c, errors.New("invalid request"), 500)
		return
	}
	clientId := setting.GetStr(conf.SSOClientId)
	platform := setting.GetStr(conf.SSOLoginPlatform)
	clientSecret := setting.GetStr(conf.SSOClientSecret)
	var tokenUrl, userUrl, scope, authField, idField, usernameField string
	additionalForm := make(map[string]string)
	switch platform {
	case "Github":
		tokenUrl = "https://github.com/login/oauth/access_token"
		userUrl = "https://api.github.com/user"
		authField = "code"
		scope = "read:user"
		idField = "id"
		usernameField = "login"
	case "Microsoft":
		tokenUrl = "https://login.microsoftonline.com/common/oauth2/v2.0/token"
		userUrl = "https://graph.microsoft.com/v1.0/me"
		additionalForm["grant_type"] = "authorization_code"
		scope = "user.read"
		authField = "code"
		idField = "id"
		usernameField = "displayName"
	case "Google":
		tokenUrl = "https://oauth2.googleapis.com/token"
		userUrl = "https://www.googleapis.com/oauth2/v1/userinfo"
		additionalForm["grant_type"] = "authorization_code"
		scope = "https://www.googleapis.com/auth/userinfo.profile"
		authField = "code"
		idField = "id"
		usernameField = "name"
	case "Dingtalk":
		tokenUrl = "https://api.dingtalk.com/v1.0/oauth2/userAccessToken"
		userUrl = "https://api.dingtalk.com/v1.0/contact/users/me"
		authField = "authCode"
		idField = "unionId"
		usernameField = "nick"
	case "Casdoor":
		endpoint := strings.TrimSuffix(setting.GetStr(conf.SSOEndpointName), "/")
		tokenUrl = endpoint + "/api/login/oauth/access_token"
		userUrl = endpoint + "/api/userinfo"
		additionalForm["grant_type"] = "authorization_code"
		scope = "profile"
		authField = "code"
		idField = "sub"
		usernameField = "preferred_username"
	case "OIDC":
		OIDCLoginCallback(c)
		return
	default:
		common.ErrorStrResp(c, "invalid platform", 400)
		return
	}
	callbackCode := c.Query(authField)
	if callbackCode == "" {
		common.ErrorStrResp(c, "No code provided", 400)
		return
	}
	var resp *resty.Response
	var err error
	if platform == "Dingtalk" {
		resp, err = ssoClient.R().SetHeader("content-type", "application/json").SetHeader("Accept", "application/json").
			SetBody(map[string]string{
				"clientId":     clientId,
				"clientSecret": clientSecret,
				"code":         callbackCode,
				"grantType":    "authorization_code",
			}).
			Post(tokenUrl)
	} else {
		var redirect_uri string
		if usecompatibility {
			redirect_uri = common.GetApiUrl(c.Request) + "/api/auth/" + argument
		} else {
			redirect_uri = common.GetApiUrl(c.Request) + "/api/auth/sso_callback" + "?method=" + argument
		}
		resp, err = ssoClient.R().SetHeader("Accept", "application/json").
			SetFormData(map[string]string{
				"client_id":     clientId,
				"client_secret": clientSecret,
				"code":          callbackCode,
				"redirect_uri":  redirect_uri,
				"scope":         scope,
			}).SetFormData(additionalForm).Post(tokenUrl)
	}
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if platform == "Dingtalk" {
		accessToken := utils.Json.Get(resp.Body(), "accessToken").ToString()
		resp, err = ssoClient.R().SetHeader("x-acs-dingtalk-access-token", accessToken).
			Get(userUrl)
	} else {
		accessToken := utils.Json.Get(resp.Body(), "access_token").ToString()
		resp, err = ssoClient.R().SetHeader("Authorization", "Bearer "+accessToken).
			Get(userUrl)
	}
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	userID := utils.Json.Get(resp.Body(), idField).ToString()
	if utils.SliceContains([]string{"", "0"}, userID) {
		common.ErrorResp(c, errors.New("error occured"), 400)
		return
	}
	if argument == "get_sso_id" {
		if usecompatibility {
			c.Redirect(302, common.GetApiUrl(c.Request)+"/@manage?sso_id="+userID)
			return
		}
		html := fmt.Sprintf(`<!DOCTYPE html>
				<head></head>
				<body>
				<script>
				window.opener.postMessage({"sso_id": "%s"}, "*")
				window.close()
				</script>
				</body>`, userID)
		c.Data(200, "text/html; charset=utf-8", []byte(html))
		return
	}
	username := utils.Json.Get(resp.Body(), usernameField).ToString()
	user, err := db.GetUserBySSOID(userID)
	if err != nil {
		user, err = autoRegister(username, userID, err)
		if err != nil {
			common.ErrorResp(c, err, 400)
			return
		}
	}
	token, err := common.GenerateToken(user.Username)
	if err != nil {
		common.ErrorResp(c, err, 400)
	}
	if usecompatibility {
		c.Redirect(302, common.GetApiUrl(c.Request)+"/@login?token="+token)
		return
	}
	html := fmt.Sprintf(`<!DOCTYPE html>
							<head></head>
							<body>
							<script>
							window.opener.postMessage({"token":"%s"}, "*")
							window.close()
							</script>
							</body>`, token)
	c.Data(200, "text/html; charset=utf-8", []byte(html))
}

package controllers

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/server/common"
	"github.com/Xhofe/alist/utils"
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/gin-gonic/gin"
	"net/http"
)

type OAuthReq struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

type OAuthResp struct {
	Token       string `json:"token"`
	AccessToken string `json:"access_token"`
}

func Verify(c *gin.Context) {
	accessToken := c.GetHeader("Authorization")
	if !utils.VerifyAccessToken(accessToken) {
		common.ErrorStrResp(c, "Invalid token", 401)
		return
	}
	common.SuccessResp(c)
}

func GetRedirectUrl(c *gin.Context) {
	redirectUri := generateRedirectUri(c.Request)
	common.SuccessResp(c, utils.GetSignInUrl(redirectUri))
}

func generateRedirectUri(r *http.Request) string {
	protocol := "http"
	host := r.Host
	if r.TLS != nil {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s/@manage", protocol, host)
}

func OAuth(c *gin.Context) {
	var req OAuthReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 401)
		return
	}

	if req.Code == "" || req.State == "" {
		common.ErrorStrResp(c, "Invalid code or state", 400)
		return
	}

	token, err := casdoorsdk.GetOAuthToken(req.Code, req.State)
	if err != nil {
		common.ErrorResp(c, err, 401)
		return
	}

	_, err = casdoorsdk.ParseJwtToken(token.AccessToken)
	if err != nil {
		common.ErrorResp(c, err, 401)
		return
	}

	common.SuccessResp(c, OAuthResp{
		AccessToken: token.AccessToken,
		Token:       conf.Token,
	})
}

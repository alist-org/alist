package handles

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/alist/v3/pkg/authn"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/go-webauthn/webauthn/webauthn"
)

func BeginAuthnLogin(c *gin.Context) {
	enabled := setting.GetBool(conf.WebauthnLoginEnabled)
	if !enabled {
		common.ErrorStrResp(c, "WebAuthn is not enabled", 403)
		return
	}
	username := c.Query("username")
	if username == "" {
		common.ErrorStrResp(c, "empty or no username provided", 400)
		return
	}
	user, err := db.GetUserByName(username)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	authninstance, err := authn.NewAuthnInstance(c.Request)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	options, sessionData, err := authninstance.BeginLogin(user)

	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	val, err := json.Marshal(sessionData)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	common.SuccessResp(c, gin.H{
		"options": options,
		"session": val,
	})
}

func FinishAuthnLogin(c *gin.Context) {
	enabled := setting.GetBool(conf.WebauthnLoginEnabled)
	if !enabled {
		common.ErrorStrResp(c, "WebAuthn is not enabled", 403)
		return
	}
	username := c.Query("username")
	user, err := db.GetUserByName(username)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	sessionDataString := c.GetHeader("session")

	authninstance, err := authn.NewAuthnInstance(c.Request)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	sessionDataBytes, err := base64.StdEncoding.DecodeString(sessionDataString)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal(sessionDataBytes, &sessionData); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	_, err = authninstance.FinishLogin(user, sessionData, c.Request)

	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	token, err := common.GenerateToken(user.Username)
	if err != nil {
		common.ErrorResp(c, err, 400, true)
		return
	}
	common.SuccessResp(c, gin.H{"token": token})
}

func BeginAuthnRegistration(c *gin.Context) {
	enabled := setting.GetBool(conf.WebauthnLoginEnabled)
	if !enabled {
		common.ErrorStrResp(c, "WebAuthn is not enabled", 403)
		return
	}
	user := c.MustGet("user").(*model.User)

	authninstance, err := authn.NewAuthnInstance(c.Request)
	if err != nil {
		common.ErrorResp(c, err, 400)
	}

	options, sessionData, err := authninstance.BeginRegistration(user)

	if err != nil {
		common.ErrorResp(c, err, 400)
	}

	val, err := json.Marshal(sessionData)
	if err != nil {
		common.ErrorResp(c, err, 400)
	}

	common.SuccessResp(c, gin.H{
		"options": options,
		"session": val,
	})
}

func FinishAuthnRegistration(c *gin.Context) {
	enabled := setting.GetBool(conf.WebauthnLoginEnabled)
	if !enabled {
		common.ErrorStrResp(c, "WebAuthn is not enabled", 403)
		return
	}
	user := c.MustGet("user").(*model.User)
	sessionDataString := c.GetHeader("Session")

	authninstance, err := authn.NewAuthnInstance(c.Request)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	sessionDataBytes, err := base64.StdEncoding.DecodeString(sessionDataString)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	var sessionData webauthn.SessionData
	if err := json.Unmarshal(sessionDataBytes, &sessionData); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	credential, err := authninstance.FinishRegistration(user, sessionData, c.Request)

	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	authn := &authn.Authn{}
	err = user.RegisterAuthn(credential, authn)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	err = op.DelUserCache(user.Username)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	common.SuccessResp(c, "Registered Successfully")
}

func DeleteAuthnLogin(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	type DeleteauthnReq struct {
		ID string `json:"id"`
	}
	var req DeleteauthnReq
	err := c.ShouldBind(&req)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	authn := &authn.Authn{}
	user.RemoveAuthn(req.ID, authn)
	err = op.DelUserCache(user.Username)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	common.SuccessResp(c, "Deleted Successfully")
}

func GetAuthnCredentials(c *gin.Context) {
	type WebAuthnCredentials struct {
		ID          []byte `json:"id"`
		FingerPrint string `json:"fingerprint"`
	}
	user := c.MustGet("user").(*model.User)
	credentials := user.WebAuthnCredentials()
	res := make([]WebAuthnCredentials, 0, len(credentials))
	for _, v := range credentials {
		credential := WebAuthnCredentials{
			ID:          v.ID,
			FingerPrint: fmt.Sprintf("% X", v.Authenticator.AAGUID),
		}
		res = append(res, credential)
	}
	common.SuccessResp(c, res)
}

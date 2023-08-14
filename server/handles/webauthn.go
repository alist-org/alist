package handles

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/alist-org/alist/v3/internal/authn"
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/setting"
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
	authnInstance, err := authn.NewAuthnInstance(c.Request)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	options, sessionData, err := authnInstance.BeginLogin(user)

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

	authnInstance, err := authn.NewAuthnInstance(c.Request)
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

	_, err = authnInstance.FinishLogin(user, sessionData, c.Request)

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

	authnInstance, err := authn.NewAuthnInstance(c.Request)
	if err != nil {
		common.ErrorResp(c, err, 400)
	}

	options, sessionData, err := authnInstance.BeginRegistration(user)

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

	authnInstance, err := authn.NewAuthnInstance(c.Request)
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

	credential, err := authnInstance.FinishRegistration(user, sessionData, c.Request)

	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	err = db.RegisterAuthn(user, credential)
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
	type DeleteAuthnReq struct {
		ID string `json:"id"`
	}
	var req DeleteAuthnReq
	err := c.ShouldBind(&req)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	err = db.RemoveAuthn(user, req.ID)
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

package server

import (
	"github.com/Xhofe/alist/model"
	"github.com/gin-gonic/gin"
)

func SaveSettings(c *gin.Context) {
	var req []model.SettingItem
	if err := c.ShouldBind(&req); err != nil {
		ErrorResp(c, err, 400)
		return
	}
	if err := model.SaveSettings(req); err != nil {
		ErrorResp(c, err, 500)
	} else {
		model.LoadSettings()
		SuccessResp(c)
	}
}

func GetSettings(c *gin.Context) {
	settings, err := model.GetSettings()
	if err != nil {
		ErrorResp(c, err, 400)
		return
	}
	SuccessResp(c, settings)
}

func GetSettingsPublic(c *gin.Context) {
	settings, err := model.GetSettingsPublic()
	if err != nil {
		ErrorResp(c, err, 400)
		return
	}
	SuccessResp(c, settings)
}

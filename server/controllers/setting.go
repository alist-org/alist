package controllers

import (
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/gin-gonic/gin"
)

func SaveSettings(c *gin.Context) {
	var req []model.SettingItem
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := model.SaveSettings(req); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		model.LoadSettings()
		common.SuccessResp(c)
	}
}

func GetSettings(c *gin.Context) {
	settings, err := model.GetSettings()
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	common.SuccessResp(c, settings)
}

func GetSettingsPublic(c *gin.Context) {
	settings, err := model.GetSettingsPublic()
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	*settings = append(*settings, model.SettingItem{
		Key:         "no cors",
		Value:       base.GetNoCors(),
		Description: "",
		Type:        "string",
	})
	common.SuccessResp(c, settings)
}

func DeleteSetting(c *gin.Context) {
	key := c.Query("key")
	if err := model.DeleteSetting(key); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

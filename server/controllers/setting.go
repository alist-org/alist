package controllers

import (
	"github.com/Xhofe/alist/drivers"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/gin-gonic/gin"
	"strconv"
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
	groupStr := c.Query("group")
	var settings []model.SettingItem
	var err error
	if groupStr == "" {
		settings, err = model.GetSettings()
	} else {
		group, err := strconv.Atoi(groupStr)
		if err == nil {
			settings, err = model.GetSettingsByGroup(group)
		}
	}
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
	settings = append(settings, []model.SettingItem{{
		Key:         "no cors",
		Value:       drivers.NoCors,
		Description: "",
		Type:        "string",
	}, {
		Key:         "no upload",
		Value:       drivers.NoUpload,
		Description: "",
		Type:        "string",
	}}...)
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

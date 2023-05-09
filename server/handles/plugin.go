package handles

import (
	"strconv"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/plugin"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

func ListPlugin(c *gin.Context) {
	var req model.PageReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Validate()
	storages, total, err := db.GetPlugins(req.Page, req.PerPage)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: storages,
		Total:   total,
	})
}

func DisablePlugin(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	if err := plugin.DisablePluginByID(c, uint(id)); err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.SuccessResp(c)
}

func EnablePlugin(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	if err := plugin.EnablePluginByID(c, uint(id)); err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.SuccessResp(c)
}

func GetPluginRepository(c *gin.Context) {
	plugins := plugin.GetPluginRepository(c)
	common.SuccessResp(c, plugins)
}

func UpdatePluginRepository(c *gin.Context) {
	if err := plugin.UpdatePluginRepository(c); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	plugins := plugin.GetPluginRepository(c)
	common.SuccessResp(c, plugins)
}

type InstallReq struct {
	UUID    string `json:"uuid"`
	Version string `json:"version"`
}

func InstallPlugin(c *gin.Context) {
	var req InstallReq
	if err := c.BindJSON(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	plugin, err := plugin.InstallPlugin(c, req.UUID, req.Version)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, plugin)
}

func UninstallPlugin(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := plugin.UninstallPlugin(c, uint(id)); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

type UpdateReq struct {
	UUID    string `json:"uuid"`
	Version string `json:"version"`
}

func UpdatePlugin(c *gin.Context) {
	var req UpdateReq
	if err := c.BindJSON(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	plugin, err := plugin.UpdatePlugin(c, req.UUID, req.Version)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c, plugin)
}

package handles

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ListMetas(c *gin.Context) {
	var req common.PageReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Validate()
	log.Debugf("%+v", req)
	metas, total, err := db.GetMetas(req.Page, req.PerPage)
	if err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: metas,
		Total:   total,
	})
}

func CreateMeta(c *gin.Context) {
	var req model.Meta
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	r, err := validHide(req.Hide)
	if err != nil {
		common.ErrorStrResp(c, fmt.Sprintf("%s is illegal: %s", r, err.Error()), 400)
		return
	}
	req.Path = utils.StandardizePath(req.Path)
	if err := db.CreateMeta(&req); err != nil {
		common.ErrorResp(c, err, 500, true)
	} else {
		common.SuccessResp(c)
	}
}

func UpdateMeta(c *gin.Context) {
	var req model.Meta
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	r, err := validHide(req.Hide)
	if err != nil {
		common.ErrorStrResp(c, fmt.Sprintf("%s is illegal: %s", r, err.Error()), 400)
		return
	}
	req.Path = utils.StandardizePath(req.Path)
	if err := db.UpdateMeta(&req); err != nil {
		common.ErrorResp(c, err, 500, true)
	} else {
		common.SuccessResp(c)
	}
}

func validHide(hide string) (string, error) {
	rs := strings.Split(hide, "\n")
	for _, r := range rs {
		_, err := regexp.Compile(r)
		if err != nil {
			return r, err
		}
	}
	return "", nil
}

func DeleteMeta(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := db.DeleteMetaById(uint(id)); err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.SuccessResp(c)
}

func GetMeta(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	meta, err := db.GetMetaById(uint(id))
	if err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.SuccessResp(c, meta)
}

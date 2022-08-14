package handles

import (
	"strconv"

	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ListUsers(c *gin.Context) {
	var req common.PageReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	req.Validate()
	log.Debugf("%+v", req)
	users, total, err := db.GetUsers(req.Page, req.PerPage)
	if err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.SuccessResp(c, common.PageResp{
		Content: users,
		Total:   total,
	})
}

func CreateUser(c *gin.Context) {
	var req model.User
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if req.IsAdmin() || req.IsGuest() {
		common.ErrorStrResp(c, "admin or guest user can not be created", 400, true)
		return
	}
	if err := db.CreateUser(&req); err != nil {
		common.ErrorResp(c, err, 500, true)
	} else {
		common.SuccessResp(c)
	}
}

func UpdateUser(c *gin.Context) {
	var req model.User
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user, err := db.GetUserById(req.ID)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if user.Role != req.Role {
		common.ErrorStrResp(c, "role can not be changed", 400)
		return
	}
	if req.Password == "" {
		req.Password = user.Password
	}
	if req.OtpSecret == "" {
		req.OtpSecret = user.OtpSecret
	}
	if err := db.UpdateUser(&req); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func DeleteUser(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := db.DeleteUserById(uint(id)); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

func GetUser(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user, err := db.GetUserById(uint(id))
	if err != nil {
		common.ErrorResp(c, err, 500, true)
		return
	}
	common.SuccessResp(c, user)
}

func Cancel2FAById(c *gin.Context) {
	idStr := c.Query("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := db.Cancel2FAById(uint(id)); err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	common.SuccessResp(c)
}

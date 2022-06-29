package controllers

import (
	"github.com/alist-org/alist/v3/internal/aria2"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"strconv"
)

func UndoneDownTask(c *gin.Context) {
	common.SuccessResp(c, aria2.DownTaskManager.ListUndone())
}

func DoneDownTask(c *gin.Context) {
	common.SuccessResp(c, aria2.DownTaskManager.ListDone())
}

func CancelDownTask(c *gin.Context) {
	tid := c.Query("tid")
	if err := aria2.DownTaskManager.Cancel(tid); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func UndoneTransferTask(c *gin.Context) {
	common.SuccessResp(c, aria2.TransferTaskManager.ListUndone())
}

func DoneTransferTask(c *gin.Context) {
	common.SuccessResp(c, aria2.TransferTaskManager.ListDone())
}

func CancelTransferTask(c *gin.Context) {
	id := c.Query("tid")
	tid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := aria2.TransferTaskManager.Cancel(tid); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func UndoneUploadTask(c *gin.Context) {
	common.SuccessResp(c, fs.UploadTaskManager.ListUndone())
}

func DoneUploadTask(c *gin.Context) {
	common.SuccessResp(c, fs.UploadTaskManager.ListDone())
}

func CancelUploadTask(c *gin.Context) {
	id := c.Query("tid")
	tid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := fs.UploadTaskManager.Cancel(tid); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func UndoneCopyTask(c *gin.Context) {
	common.SuccessResp(c, fs.CopyTaskManager.ListUndone())
}

func DoneCopyTask(c *gin.Context) {
	common.SuccessResp(c, fs.CopyTaskManager.ListDone())
}

func CancelCopyTask(c *gin.Context) {
	id := c.Query("tid")
	tid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := fs.CopyTaskManager.Cancel(tid); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

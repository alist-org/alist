package handles

import (
	"strconv"

	"github.com/alist-org/alist/v3/internal/aria2"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
)

type TaskInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	State    string `json:"state"`
	Status   string `json:"status"`
	Progress int    `json:"progress"`
	Error    string `json:"error"`
}

func getTaskInfoUint(task *task.Task[uint64]) TaskInfo {
	return TaskInfo{
		ID:       strconv.FormatUint(task.ID, 10),
		Name:     task.Name,
		State:    task.GetState(),
		Status:   task.GetStatus(),
		Progress: task.GetProgress(),
		Error:    task.GetErrMsg(),
	}
}

func getTaskInfoStr(task *task.Task[string]) TaskInfo {
	return TaskInfo{
		ID:       task.ID,
		Name:     task.Name,
		State:    task.GetState(),
		Status:   task.GetStatus(),
		Progress: task.GetProgress(),
		Error:    task.GetErrMsg(),
	}
}

func getTaskInfosUint(tasks []*task.Task[uint64]) []TaskInfo {
	var infos []TaskInfo
	for _, t := range tasks {
		infos = append(infos, getTaskInfoUint(t))
	}
	return infos
}

func getTaskInfosStr(tasks []*task.Task[string]) []TaskInfo {
	var infos []TaskInfo
	for _, t := range tasks {
		infos = append(infos, getTaskInfoStr(t))
	}
	return infos
}

func UndoneDownTask(c *gin.Context) {
	common.SuccessResp(c, getTaskInfosStr(aria2.DownTaskManager.ListUndone()))
}

func DoneDownTask(c *gin.Context) {
	common.SuccessResp(c, getTaskInfosStr(aria2.DownTaskManager.ListDone()))
}

func CancelDownTask(c *gin.Context) {
	tid := c.Query("tid")
	if err := aria2.DownTaskManager.Cancel(tid); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func DeleteDownTask(c *gin.Context) {
	tid := c.Query("tid")
	if err := aria2.DownTaskManager.Remove(tid); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func ClearDoneDownTasks(c *gin.Context) {
	aria2.DownTaskManager.ClearDone()
	common.SuccessResp(c)
}

func UndoneTransferTask(c *gin.Context) {
	common.SuccessResp(c, getTaskInfosUint(aria2.TransferTaskManager.ListUndone()))
}

func DoneTransferTask(c *gin.Context) {
	common.SuccessResp(c, getTaskInfosUint(aria2.TransferTaskManager.ListDone()))
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

func DeleteTransferTask(c *gin.Context) {
	id := c.Query("tid")
	tid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := aria2.TransferTaskManager.Remove(tid); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func ClearDoneTransferTasks(c *gin.Context) {
	aria2.TransferTaskManager.ClearDone()
	common.SuccessResp(c)
}

func UndoneUploadTask(c *gin.Context) {
	common.SuccessResp(c, getTaskInfosUint(fs.UploadTaskManager.ListUndone()))
}

func DoneUploadTask(c *gin.Context) {
	common.SuccessResp(c, getTaskInfosUint(fs.UploadTaskManager.ListDone()))
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

func DeleteUploadTask(c *gin.Context) {
	id := c.Query("tid")
	tid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := fs.UploadTaskManager.Remove(tid); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func ClearDoneUploadTasks(c *gin.Context) {
	fs.UploadTaskManager.ClearDone()
	common.SuccessResp(c)
}

func UndoneCopyTask(c *gin.Context) {
	common.SuccessResp(c, getTaskInfosUint(fs.CopyTaskManager.ListUndone()))
}

func DoneCopyTask(c *gin.Context) {
	common.SuccessResp(c, getTaskInfosUint(fs.CopyTaskManager.ListDone()))
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

func DeleteCopyTask(c *gin.Context) {
	id := c.Query("tid")
	tid, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	if err := fs.CopyTaskManager.Remove(tid); err != nil {
		common.ErrorResp(c, err, 500)
	} else {
		common.SuccessResp(c)
	}
}

func ClearDoneCopyTasks(c *gin.Context) {
	fs.CopyTaskManager.ClearDone()
	common.SuccessResp(c)
}

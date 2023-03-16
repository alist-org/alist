package handles

import (
	"strconv"

	"github.com/alist-org/alist/v3/internal/aria2"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/qbittorrent"
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

type K2Str[K comparable] func(k K) string

func uint64K2Str(k uint64) string {
	return strconv.FormatUint(k, 10)
}

func strK2Str(str string) string {
	return str
}

func getTaskInfo[K comparable](task *task.Task[K], k2Str K2Str[K]) TaskInfo {
	return TaskInfo{
		ID:       k2Str(task.ID),
		Name:     task.Name,
		State:    task.GetState(),
		Status:   task.GetStatus(),
		Progress: task.GetProgress(),
		Error:    task.GetErrMsg(),
	}
}

func getTaskInfos[K comparable](tasks []*task.Task[K], k2Str K2Str[K]) []TaskInfo {
	var infos []TaskInfo
	for _, t := range tasks {
		infos = append(infos, getTaskInfo(t, k2Str))
	}
	return infos
}

type Str2K[K comparable] func(str string) (K, error)

func str2Uint64K(str string) (uint64, error) {
	return strconv.ParseUint(str, 10, 64)
}

func str2StrK(str string) (string, error) {
	return str, nil
}

func taskRoute[K comparable](g *gin.RouterGroup, manager *task.Manager[K], k2Str K2Str[K], str2K Str2K[K]) {
	g.GET("/undone", func(c *gin.Context) {
		common.SuccessResp(c, getTaskInfos(manager.ListUndone(), k2Str))
	})
	g.GET("/done", func(c *gin.Context) {
		common.SuccessResp(c, getTaskInfos(manager.ListDone(), k2Str))
	})
	g.POST("/cancel", func(c *gin.Context) {
		tid := c.Query("tid")
		id, err := str2K(tid)
		if err != nil {
			common.ErrorResp(c, err, 400)
			return
		}
		if err := manager.Cancel(id); err != nil {
			common.ErrorResp(c, err, 500)
		} else {
			common.SuccessResp(c)
		}
	})
	g.POST("/delete", func(c *gin.Context) {
		tid := c.Query("tid")
		id, err := str2K(tid)
		if err != nil {
			common.ErrorResp(c, err, 400)
			return
		}
		if err := manager.Remove(id); err != nil {
			common.ErrorResp(c, err, 500)
		} else {
			common.SuccessResp(c)
		}
	})
	g.POST("/retry", func(c *gin.Context) {
		tid := c.Query("tid")
		id, err := str2K(tid)
		if err != nil {
			common.ErrorResp(c, err, 400)
			return
		}
		if err := manager.Retry(id); err != nil {
			common.ErrorResp(c, err, 500)
		} else {
			common.SuccessResp(c)
		}
	})
	g.POST("/clear_done", func(c *gin.Context) {
		manager.ClearDone()
		common.SuccessResp(c)
	})
	g.POST("/clear_succeeded", func(c *gin.Context) {
		manager.ClearSucceeded()
		common.SuccessResp(c)
	})
}

func SetupTaskRoute(g *gin.RouterGroup) {
	taskRoute(g.Group("/aria2_down"), aria2.DownTaskManager, strK2Str, str2StrK)
	taskRoute(g.Group("/aria2_transfer"), aria2.TransferTaskManager, uint64K2Str, str2Uint64K)
	taskRoute(g.Group("/upload"), fs.UploadTaskManager, uint64K2Str, str2Uint64K)
	taskRoute(g.Group("/copy"), fs.CopyTaskManager, uint64K2Str, str2Uint64K)
	taskRoute(g.Group("/qbit_down"), qbittorrent.DownTaskManager, strK2Str, str2StrK)
	taskRoute(g.Group("/qbit_transfer"), qbittorrent.TransferTaskManager, uint64K2Str, str2Uint64K)
}

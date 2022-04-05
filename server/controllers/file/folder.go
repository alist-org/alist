package file

import (
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/server/common"
	"github.com/gin-gonic/gin"
)

type FolderReq struct {
	Path string `json:"path"`
}

func Folder(c *gin.Context) {
	var req FolderReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	var files = make([]model.File, 0)
	_, rawFiles, _, _, _, err := common.Path(req.Path)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}
	if rawFiles == nil {
		common.ErrorStrResp(c, "not a folder", 400)
		return
	}
	for _, file := range rawFiles {
		if file.IsDir() {
			files = append(files, file)
		}
	}
	c.JSON(200, common.Resp{
		Code:    200,
		Message: "success",
		Data:    files,
	})
}

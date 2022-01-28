package file

import (
	"github.com/Xhofe/alist/drivers/operate"
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
	var err error
	if model.AccountsCount() > 1 && (req.Path == "/" || req.Path == "") {
		files, err = model.GetAccountFiles()
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	} else {
		account, path, driver, err := common.ParsePath(req.Path)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		file, rawFiles, err := operate.Path(driver, account, path)
		if err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
		if file != nil {
			common.ErrorStrResp(c, "Not folder", 400)
		}
		for _, file := range rawFiles {
			if file.IsDir() {
				files = append(files, file)
			}
		}
	}
	c.JSON(200, common.Resp{
		Code:    200,
		Message: "success",
		Data:    files,
	})
}

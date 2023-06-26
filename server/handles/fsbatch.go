package handles

import (
	"fmt"
	"regexp"

	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/generic"
	"github.com/alist-org/alist/v3/server/common"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type BatchRenameReq struct {
	SrcDir        string `json:"src_dir"`
	RenameObjects []struct {
		SrcName string `json:"src_name"`
		NewName string `json:"new_name"`
	} `json:"rename_objects"`
}

func FsBatchRename(c *gin.Context) {
	var req BatchRenameReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	if !user.CanRename() {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}

	reqPath, err := user.JoinPath(req.SrcDir)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}

	meta, err := op.GetNearestMeta(reqPath)
	if err != nil {
		if !errors.Is(errors.Cause(err), errs.MetaNotFound) {
			common.ErrorResp(c, err, 500, true)
			return
		}
	}
	c.Set("meta", meta)
	for _, renameObject := range req.RenameObjects {
		if renameObject.SrcName == "" || renameObject.NewName == "" {
			continue
		}
		filePath := fmt.Sprintf("%s/%s", reqPath, renameObject.SrcName)
		if err := fs.Rename(c, filePath, renameObject.NewName); err != nil {
			common.ErrorResp(c, err, 500)
			return
		}
	}
	common.SuccessResp(c)
}

type RecursiveMoveReq struct {
	SrcDir string `json:"src_dir"`
	DstDir string `json:"dst_dir"`
}

func FsRecursiveMove(c *gin.Context) {
	var req RecursiveMoveReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}

	user := c.MustGet("user").(*model.User)
	if !user.CanMove() {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}
	srcDir, err := user.JoinPath(req.SrcDir)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}
	dstDir, err := user.JoinPath(req.DstDir)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}

	meta, err := op.GetNearestMeta(srcDir)
	if err != nil {
		if !errors.Is(errors.Cause(err), errs.MetaNotFound) {
			common.ErrorResp(c, err, 500, true)
			return
		}
	}
	c.Set("meta", meta)

	rootFiles, err := fs.List(c, srcDir, &fs.ListArgs{})
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	// record the file path
	filePathMap := make(map[model.Obj]string)
	movingFiles := generic.NewQueue[model.Obj]()
	for _, file := range rootFiles {
		movingFiles.Push(file)
		filePathMap[file] = srcDir
	}

	for !movingFiles.IsEmpty() {

		movingFile := movingFiles.Pop()
		movingFilePath := filePathMap[movingFile]
		movingFileName := fmt.Sprintf("%s/%s", movingFilePath, movingFile.GetName())
		if movingFile.IsDir() {
			// directory, recursive move
			subFilePath := movingFileName
			subFiles, err := fs.List(c, movingFileName, &fs.ListArgs{Refresh: true})
			if err != nil {
				common.ErrorResp(c, err, 500)
				return
			}
			for _, subFile := range subFiles {
				movingFiles.Push(subFile)
				filePathMap[subFile] = subFilePath
			}
		} else {

			if movingFilePath == dstDir {
				// same directory, don't move
				continue
			}

			// move
			err := fs.Move(c, movingFileName, dstDir, movingFiles.IsEmpty())
			if err != nil {
				common.ErrorResp(c, err, 500)
				return
			}
		}

	}

	common.SuccessResp(c)
}

type RegexRenameReq struct {
	SrcDir       string `json:"src_dir"`
	SrcNameRegex string `json:"src_name_regex"`
	NewNameRegex string `json:"new_name_regex"`
}

func FsRegexRename(c *gin.Context) {
	var req RegexRenameReq
	if err := c.ShouldBind(&req); err != nil {
		common.ErrorResp(c, err, 400)
		return
	}
	user := c.MustGet("user").(*model.User)
	if !user.CanRename() {
		common.ErrorResp(c, errs.PermissionDenied, 403)
		return
	}

	reqPath, err := user.JoinPath(req.SrcDir)
	if err != nil {
		common.ErrorResp(c, err, 403)
		return
	}

	meta, err := op.GetNearestMeta(reqPath)
	if err != nil {
		if !errors.Is(errors.Cause(err), errs.MetaNotFound) {
			common.ErrorResp(c, err, 500, true)
			return
		}
	}
	c.Set("meta", meta)

	srcRegexp, err := regexp.Compile(req.SrcNameRegex)
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	files, err := fs.List(c, reqPath, &fs.ListArgs{})
	if err != nil {
		common.ErrorResp(c, err, 500)
		return
	}

	for _, file := range files {

		if srcRegexp.MatchString(file.GetName()) {
			filePath := fmt.Sprintf("%s/%s", reqPath, file.GetName())
			newFileName := srcRegexp.ReplaceAllString(file.GetName(), req.NewNameRegex)
			if err := fs.Rename(c, filePath, newFileName); err != nil {
				common.ErrorResp(c, err, 500)
				return
			}
		}

	}

	common.SuccessResp(c)
}

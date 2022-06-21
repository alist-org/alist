package aria2

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/conf"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/operations"
	"github.com/alist-org/alist/v3/pkg/task"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"
)

func AddURI(ctx context.Context, uri string, dstPath string, parentPath string) error {
	// check account
	account, actualParentPath, err := operations.GetAccountAndActualPath(parentPath)
	if err != nil {
		return errors.WithMessage(err, "failed get account")
	}
	// check is it could upload
	if account.Config().NoUpload {
		return errors.WithStack(fs.ErrUploadNotSupported)
	}
	// check path is valid
	obj, err := operations.Get(ctx, account, actualParentPath)
	if err != nil {
		if !errors.Is(errors.Cause(err), driver.ErrorObjectNotFound) {
			return errors.WithMessage(err, "failed get object")
		}
	} else {
		if !obj.IsDir() {
			// can't add to a file
			return errors.WithStack(fs.ErrNotFolder)
		}
	}
	// call aria2 rpc
	tempDir := filepath.Join(conf.Conf.TempDir, "aria2", uuid.NewString())
	options := map[string]interface{}{
		"dir": tempDir,
	}
	gid, err := client.AddURI([]string{uri}, options)
	if err != nil {
		return errors.Wrapf(err, "failed to add uri %s", uri)
	}
	// TODO add to task manager
	Aria2TaskManager.Submit(task.WithCancelCtx(&task.Task[string, OfflineDownload]{
		ID:   gid,
		Name: fmt.Sprintf("download %s to [%s](%s)", uri, account.GetAccount().VirtualPath, actualParentPath),
		Func: func(tsk *task.Task[string, OfflineDownload]) error {
			defer func() {
				notify.Signals.Delete(gid)
				// clear temp dir
				_ = os.RemoveAll(tempDir)
			}()
			c := make(chan int)
			notify.Signals.Store(gid, c)
			retried := 0
			for {
				select {
				case <-tsk.Ctx.Done():
					_, err := client.Remove(gid)
					if err != nil {
						return err
					}
				case status := <-c:
					switch status {
					case Completed:
						return nil
					default:
						info, err := client.TellStatus(gid)
						if err != nil {
							retried++
						}
						if retried > 5 {
							return errors.Errorf("failed to get status of %s, retried %d times", gid, retried)
						}
						retried = 0
						if len(info.FollowedBy) != 0 {
							gid = info.FollowedBy[0]

						}
						// update download status
						total, err := strconv.ParseUint(info.TotalLength, 10, 64)
						if err != nil {
							total = 0
						}
						downloaded, err := strconv.ParseUint(info.CompletedLength, 10, 64)
						if err != nil {
							downloaded = 0
						}
						tsk.SetProgress(int(float64(downloaded) / float64(total)))
						switch info.Status {
						case "complete":
							// get files
							files, err := client.GetFiles(gid)
							if err != nil {
								return errors.Wrapf(err, "failed to get files of %s", gid)
							}
							// upload files
							for _, file := range files {
								size, _ := strconv.ParseUint(file.Length, 10, 64)
								f, err := os.Open(file.Path)
								mimetype := mime.TypeByExtension(path.Ext(file.Path))
								if mimetype == "" {
									mimetype = "application/octet-stream"
								}
								if err != nil {
									return errors.Wrapf(err, "failed to open file %s", file.Path)
								}
								stream := model.FileStream{
									Obj: model.Object{
										Name:     path.Base(file.Path),
										Size:     size,
										Modified: time.Now(),
										IsFolder: false,
									},
									ReadCloser: f,
									Mimetype:   "",
								}
								return operations.Put(tsk.Ctx, account, actualParentPath, stream, tsk.SetProgress)
							}
						case "error":
							return errors.Errorf("failed to download %s, error: %s", gid, info.ErrorMessage)
						case "active", "waiting", "paused":
							// do nothing
						case "removed":
							return errors.Errorf("failed to download %s, removed", gid)
						default:
							return errors.Errorf("failed to download %s, unknown status %s", gid, info.Status)
						}
					}
				}
			}
		},
		Data: OfflineDownload{
			Gid:     gid,
			URI:     uri,
			DstPath: dstPath,
		},
	}))
	return nil
}

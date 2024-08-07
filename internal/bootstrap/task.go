package bootstrap

import (
	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/offline_download/tool"
	"github.com/xhofe/tache"
)

func InitTaskManager() {
	fs.UploadTaskManager = tache.NewManager[*fs.UploadTask](tache.WithWorks(conf.Conf.Tasks.Upload.Workers), tache.WithMaxRetry(conf.Conf.Tasks.Upload.MaxRetry)) //upload will not support persist
	fs.CopyTaskManager = tache.NewManager[*fs.CopyTask](tache.WithWorks(conf.Conf.Tasks.Copy.Workers), tache.WithPersistFunction(db.GetTaskDataFunc("copy", conf.Conf.Tasks.Copy.TaskPersistant), db.UpdateTaskDataFunc("copy", conf.Conf.Tasks.Copy.TaskPersistant)), tache.WithMaxRetry(conf.Conf.Tasks.Copy.MaxRetry))
	tool.DownloadTaskManager = tache.NewManager[*tool.DownloadTask](tache.WithWorks(conf.Conf.Tasks.Download.Workers), tache.WithPersistFunction(db.GetTaskDataFunc("download", conf.Conf.Tasks.Download.TaskPersistant), db.UpdateTaskDataFunc("download", conf.Conf.Tasks.Download.TaskPersistant)), tache.WithMaxRetry(conf.Conf.Tasks.Download.MaxRetry))
	tool.TransferTaskManager = tache.NewManager[*tool.TransferTask](tache.WithWorks(conf.Conf.Tasks.Transfer.Workers), tache.WithPersistFunction(db.GetTaskDataFunc("transfer", conf.Conf.Tasks.Transfer.TaskPersistant), db.UpdateTaskDataFunc("transfer", conf.Conf.Tasks.Transfer.TaskPersistant)), tache.WithMaxRetry(conf.Conf.Tasks.Transfer.MaxRetry))
	if len(tool.TransferTaskManager.GetAll()) == 0 { //prevent offline downloaded files from being deleted
		CleanTempDir()
	}
}

package bootstrap

import (
	"context"
	"time"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/db"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
)

func LoadStorages() {
	storages, err := db.GetEnabledStorages()
	if err != nil {
		utils.Log.Fatalf("failed get enabled storages: %+v", err)
	}
	go func(storages []model.Storage) {
		for i := range storages {
			err := op.LoadStorage(context.Background(), storages[i])
			if err != nil {
				// 挂载失败，重试一次
				time.Sleep(1*time.Second)
				err = op.LoadStorage(context.Background(), storages[i])
				if err != nil {

					utils.Log.Errorf("failed get enabled storages: %+v", err)
				} else {
					utils.Log.Infof("success load storage: [%s], driver: [%s]",
						storages[i].MountPath, storages[i].Driver)
				}
			} else {
				utils.Log.Infof("success load storage: [%s], driver: [%s]",
					storages[i].MountPath, storages[i].Driver)
			}
		}
		conf.StoragesLoaded = true
	}(storages)
}

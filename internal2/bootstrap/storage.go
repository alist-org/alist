package bootstrap

import (
	"context"

	"github.com/alist-org/alist/v3/internal2/conf"
	"github.com/alist-org/alist/v3/internal2/db"
	"github.com/alist-org/alist/v3/internal2/model"
	"github.com/alist-org/alist/v3/internal2/op"
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
				utils.Log.Errorf("failed get enabled storages: %+v", err)
			} else {
				utils.Log.Infof("success load storage: [%s], driver: [%s]",
					storages[i].MountPath, storages[i].Driver)
			}
		}
		conf.StoragesLoaded = true
	}(storages)
}

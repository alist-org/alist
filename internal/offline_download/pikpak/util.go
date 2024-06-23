package pikpak

import (
	// "github.com/Xhofe/go-cache"

	"context"

	"github.com/alist-org/alist/v3/drivers/pikpak"
	"github.com/alist-org/alist/v3/internal/op"

	"github.com/alist-org/alist/v3/pkg/singleflight"
)

// var taskCache = cache.NewMemCache(cache.WithShards[[]pikpak.OfflineTask](16))
var taskG singleflight.Group[[]pikpak.OfflineTask]

func GetTasks(pikpakDriver *pikpak.PikPak) ([]pikpak.OfflineTask, error) {
	key := op.Key(pikpakDriver, "/drive/v1/task")
	tasks, err, _ := taskG.Do(key, func() ([]pikpak.OfflineTask, error) {
		ctx := context.Background()
		phase := []string{"PHASE_TYPE_RUNNING", "PHASE_TYPE_ERROR", "PHASE_TYPE_PENDING", "PHASE_TYPE_COMPLETE"}
		tasks, err := pikpakDriver.OfflineList(ctx, "", phase)
		if err != nil {
			return nil, err
		}
		return tasks, nil
	})
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

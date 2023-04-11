package offline_download

import (
	"fmt"

	"github.com/alist-org/alist/v3/pkg/task"
)

var (
	Tools           = make(ToolsManager)
	DownTaskManager = task.NewTaskManager[string](3)
)

type ToolsManager map[string]OfflineDownload

func (t ToolsManager) Get(name string) (OfflineDownload, error) {
	if tool, ok := t[name]; ok {
		return tool, nil
	}
	return nil, fmt.Errorf("tool %s not found", name)
}

func (t ToolsManager) Add(name string, tool OfflineDownload) {
	t[name] = tool
}

func (t ToolsManager) Names() []string {
	names := make([]string, 0, len(t))
	for name := range t {
		names = append(names, name)
	}
	return names
}

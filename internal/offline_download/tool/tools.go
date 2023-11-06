package tool

import (
	"fmt"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/task"
)

var (
	Tools           = make(ToolsManager)
	DownTaskManager = task.NewTaskManager[string](3)
)

type ToolsManager map[string]Tool

func (t ToolsManager) Get(name string) (Tool, error) {
	if tool, ok := t[name]; ok {
		return tool, nil
	}
	return nil, fmt.Errorf("tool %s not found", name)
}

func (t ToolsManager) Add(name string, tool Tool) {
	t[name] = tool
}

func (t ToolsManager) Names() []string {
	names := make([]string, 0, len(t))
	for name := range t {
		names = append(names, name)
	}
	return names
}

func (t ToolsManager) Items() []model.SettingItem {
	var items []model.SettingItem
	for _, tool := range t {
		items = append(items, tool.Items()...)
	}
	return items
}

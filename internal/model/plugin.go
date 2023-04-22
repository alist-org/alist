package model

const (
	PLUGIN_MODE_WASI  = "wasi"
	PLUGIN_MODE_DRPC  = "drpc"
	PLUGIN_MODE_YAEGI = "yaegi"
)

const (
	PLUGIN_TYPE_STORAGE = "storage"
)

type Plugin struct {
	ID   uint   `json:"id"  gorm:"primaryKey"`
	Type string `json:"type"` // 插件类型

	Name       string `json:"name"`        // 插件名称
	UUID       string `json:"uuid"`        // 插件唯一识别标识
	Mode       string `json:"mode"`        // 插件加载模式
	Path       string `json:"path"`        // 插件启动路径（或ip:port）(或包名)
	Version    string `json:"version"`     // 插件版本
	ApiVersion string `json:"api_version"` // 插件使用的API版本

	Disabled bool `json:"disabled"`
}

// 通用部分
type PluginControl interface {
	// 卸载插件功能
	Unload() error
	// 加载插件功能
	Load() error

	// 获取插件信息
	Config() (PluginConfig, error)

	// 释放插件
	Release() error
}

type PluginConfig interface {
	GetUUID() string
	GetType() []string
	GetName() string
	GetVersion() string
	GetApiVersion() []string
}

type PluginConfigHelper struct {
	UUID        string
	Types       []string
	Name        string
	Version     string
	ApiVersions []string
}

func (p *PluginConfigHelper) GetUUID() string {
	return p.UUID
}
func (p *PluginConfigHelper) GetType() []string {
	return p.Types
}
func (p *PluginConfigHelper) GetName() string {
	return p.Name
}
func (p *PluginConfigHelper) GetVersion() string {
	return p.Version
}
func (p *PluginConfigHelper) GetApiVersion() []string {
	return p.ApiVersions
}

var _ PluginConfig = (*PluginConfigHelper)(nil)

type PluginControlHelper struct {
	UnloadFunc  func() error
	LoadFunc    func() error
	ConfigFunc  func() (PluginConfig, error)
	ReleaseFunc func() error
}

func (w *PluginControlHelper) Unload() error {
	return w.UnloadFunc()
}

func (w *PluginControlHelper) Load() error {
	return w.LoadFunc()
}

func (w *PluginControlHelper) Config() (PluginConfig, error) {
	return w.ConfigFunc()
}

func (w *PluginControlHelper) Release() error {
	return w.ReleaseFunc()
}

var _ PluginControl = (*PluginControlHelper)(nil)

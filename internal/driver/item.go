package driver

type Additional interface{}

type Select string

type Item struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Default  string `json:"default"`
	Options  string `json:"options"`
	Required bool   `json:"required"`
	Help     string `json:"help"`
}

type Info struct {
	Common     []Item `json:"common"`
	Additional []Item `json:"additional"`
	Config     Config `json:"config"`
}

type IRootFolderPath interface {
	GetRootFolderPath() string
}

type IRootFolderId interface {
	GetRootFolderId() string
}

type RootFolderPath struct {
	RootFolder string `json:"root_folder" required:"true" help:"root folder path"`
}

type RootFolderId struct {
	RootFolder string `json:"root_folder" required:"true" help:"root folder id"`
}

func (r RootFolderPath) GetRootFolderPath() string {
	return r.RootFolder
}

func (r RootFolderId) GetRootFolderId() string {
	return r.RootFolder
}

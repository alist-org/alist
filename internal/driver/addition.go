package driver

type Additional interface{}

type Select string

const (
	TypeString = "string"
	TypeSelect = "select"
	TypeBool   = "bool"
	TypeText   = "text"
)

type Item struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Default  string `json:"default"`
	Values   string `json:"values"`
	Required bool   `json:"required"`
	Help     string `json:"help"`
}

type Items struct {
	Main       []Item `json:"main"`
	Additional []Item `json:"additional"`
}

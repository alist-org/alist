package driver

type Additional interface {
}

type Item struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Default  string `json:"default"`
	Values   string `json:"values"`
	Required bool   `json:"required"`
	Desc     string `json:"desc"`
}

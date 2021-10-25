package model

type ConfigItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  int    `json:"type"`
}

package lanzou

type DownPageResp struct {
	Zt   int `json:"zt"`
	Info struct {
		Pwd    string `json:"pwd"`
		Onof   string `json:"onof"`
		FId    string `json:"f_id"`
		Taoc   string `json:"taoc"`
		IsNewd string `json:"is_newd"`
	} `json:"info"`
	Text interface{} `json:"text"`
}

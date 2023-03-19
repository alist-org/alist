package baidu_share

import "encoding/json"

type jsonResp struct {
	Errno int64 `json:"errno"`
	Data  struct {
		More bool `json:"has_more"`
		List []struct {
			ID    json.Number `json:"fs_id"`
			Dir   json.Number `json:"isdir"`
			Path  string      `json:"path"`
			Name  string      `json:"server_filename"`
			Time  json.Number `json:"server_mtime"`
			Size  json.Number `json:"size"`
			Dlink string      `json:"dlink"`
		} `json:"list"`
	} `json:"data"`
}

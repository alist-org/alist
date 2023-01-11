package seafile

type AuthTokenResp struct {
	Token string `json:"token"`
}

type RepoDirItemResp struct {
	Id         string `json:"id"`
	Type       string `json:"type"` // dir, file
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Modified   int64  `json:"mtime"`
	Permission string `json:"permission"`
}

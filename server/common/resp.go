package common

type Resp[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type PageResp struct {
	Content interface{} `json:"content"`
	Total   int64       `json:"total"`
}

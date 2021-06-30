package controllers

type Response struct {
	Code int `json:"code"`
	Data interface{} `json:"data"`
	Message string `json:"message"`
}

// MetaResponse common meta response
func MetaResponse(code int, msg string) Response {
	return Response{
		Code:    code,
		Data:    nil,
		Message: msg,
	}
}

// DataResponse common data response
func DataResponse(data interface{})  Response {
	return Response{
		Code:    200,
		Data:    data,
		Message: "ok",
	}
}

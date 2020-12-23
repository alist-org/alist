package alidrive

type ListReq struct {
	DriveId               string	`json:"drive_id"`
	Fields                string	`json:"fields"`
	ImageThumbnailProcess string	`json:"image_thumbnail_process"`
	ImageUrlProcess       string	`json:"image_url_process"`
	Limit                 int		`json:"limit"`
	Marker                string	`json:"marker"`
	OrderBy               string	`json:"order_by"`
	OrderDirection        string	`json:"order_direction"`
	ParentFileId          string	`json:"parent_file_id"`
	VideoThumbnailProcess string	`json:"video_thumbnail_process"`
}

type GetReq struct {
	DriveId               string	`json:"drive_id"`
	FileId                string	`json:"file_id"`
	ImageThumbnailProcess string	`json:"image_thumbnail_process"`
	VideoThumbnailProcess string	`json:"video_thumbnail_process"`
}

type SearchReq struct {
	DriveId               string	`json:"drive_id"`
	ImageThumbnailProcess string	`json:"image_thumbnail_process"`
	ImageUrlProcess       string	`json:"image_url_process"`
	Limit                 int		`json:"limit"`
	Marker                string	`json:"marker"`
	OrderBy               string	`json:"order_by"`//"type ASC,updated_at DESC"

	Query				  string	`json:"query"`// "name match '测试文件'"

	VideoThumbnailProcess string	`json:"video_thumbnail_process"`
}

type TokenLoginReq struct {
	Token 		string	`json:"token"`
}

type GetTokenReq struct {
	Code	string	`json:"code"`
}

type RefreshTokenReq struct {
	RefreshToken string	`json:"refresh_token"`
}
package teambition

import "time"

type Collection struct {
	ID      string    `json:"_id"`
	Title   string    `json:"title"`
	Updated time.Time `json:"updated"`
}

type Work struct {
	ID           string    `json:"_id"`
	FileName     string    `json:"fileName"`
	FileSize     int64     `json:"fileSize"`
	FileKey      string    `json:"fileKey"`
	FileCategory string    `json:"fileCategory"`
	DownloadURL  string    `json:"downloadUrl"`
	ThumbnailURL string    `json:"thumbnailUrl"`
	Thumbnail    string    `json:"thumbnail"`
	Updated      time.Time `json:"updated"`
	PreviewURL   string    `json:"previewUrl"`
}

type FileUpload struct {
	FileKey        string        `json:"fileKey"`
	FileName       string        `json:"fileName"`
	FileType       string        `json:"fileType"`
	FileSize       int           `json:"fileSize"`
	FileCategory   string        `json:"fileCategory"`
	ImageWidth     int           `json:"imageWidth"`
	ImageHeight    int           `json:"imageHeight"`
	InvolveMembers []interface{} `json:"involveMembers"`
	Source         string        `json:"source"`
	Visible        string        `json:"visible"`
	ParentId       string        `json:"_parentId"`
}

type ChunkUpload struct {
	FileUpload
	Storage        string        `json:"storage"`
	MimeType       string        `json:"mimeType"`
	Chunks         int           `json:"chunks"`
	ChunkSize      int           `json:"chunkSize"`
	Created        time.Time     `json:"created"`
	FileMD5        string        `json:"fileMD5"`
	LastUpdated    time.Time     `json:"lastUpdated"`
	UploadedChunks []interface{} `json:"uploadedChunks"`
	Token          struct {
		AppID          string    `json:"AppID"`
		OrganizationID string    `json:"OrganizationID"`
		UserID         string    `json:"UserID"`
		Exp            time.Time `json:"Exp"`
		Storage        string    `json:"Storage"`
		Resource       string    `json:"Resource"`
		Speed          int       `json:"Speed"`
	} `json:"token"`
	DownloadUrl    string      `json:"downloadUrl"`
	ThumbnailUrl   string      `json:"thumbnailUrl"`
	PreviewUrl     string      `json:"previewUrl"`
	ImmPreviewUrl  string      `json:"immPreviewUrl"`
	PreviewExt     string      `json:"previewExt"`
	LastUploadTime interface{} `json:"lastUploadTime"`
}

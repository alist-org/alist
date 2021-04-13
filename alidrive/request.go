package alidrive

import (
	"fmt"
	"github.com/Xhofe/alist/conf"
)

// get file
func GetFile(fileId string, drive *conf.Drive) (*File, error) {
	url := conf.Conf.AliDrive.ApiUrl + "/file/get"
	req := GetReq{
		DriveId:               drive.DefaultDriveId,
		FileId:                fileId,
		ImageThumbnailProcess: conf.ImageThumbnailProcess,
		VideoThumbnailProcess: conf.VideoThumbnailProcess,
	}
	var resp File
	if err := BodyToJson(url, req, &resp, drive); err != nil {
		return nil, err
	}
	return &resp, nil
}

// get download_url
func GetDownLoadUrl(fileId string, drive *conf.Drive) (*DownloadResp, error) {
	url := conf.Conf.AliDrive.ApiUrl + "/file/get_download_url"
	req := DownloadReq{
		DriveId:   drive.DefaultDriveId,
		FileId:    fileId,
		ExpireSec: 14400,
	}
	var resp DownloadResp
	if err := BodyToJson(url, req, &resp, drive); err != nil {
		return nil, err
	}
	return &resp, nil
}

// search by keyword
func Search(key string, limit int, marker string, drive *conf.Drive) (*Files, error) {
	url := conf.Conf.AliDrive.ApiUrl + "/file/search"
	req := SearchReq{
		DriveId:               drive.DefaultDriveId,
		ImageThumbnailProcess: conf.ImageThumbnailProcess,
		ImageUrlProcess:       conf.ImageUrlProcess,
		Limit:                 limit,
		Marker:                marker,
		OrderBy:               conf.OrderSearch,
		Query:                 fmt.Sprintf("name match '%s'", key),
		VideoThumbnailProcess: conf.VideoThumbnailProcess,
	}
	var resp Files
	if err := BodyToJson(url, req, &resp, drive); err != nil {
		return nil, err
	}
	return &resp, nil
}

// get root folder
func GetRoot(limit int, marker string, orderBy string, orderDirection string, drive *conf.Drive) (*Files, error) {
	return GetList(drive.RootFolder, limit, marker, orderBy, orderDirection, drive)
}

// get folder list by file_id
func GetList(parent string, limit int, marker string, orderBy string, orderDirection string, drive *conf.Drive) (*Files, error) {
	url := conf.Conf.AliDrive.ApiUrl + "/file/list"
	req := ListReq{
		DriveId:               drive.DefaultDriveId,
		Fields:                "*",
		ImageThumbnailProcess: conf.ImageThumbnailProcess,
		ImageUrlProcess:       conf.ImageUrlProcess,
		Limit:                 limit,
		Marker:                marker,
		OrderBy:               orderBy,
		OrderDirection:        orderDirection,
		ParentFileId:          parent,
		VideoThumbnailProcess: conf.VideoThumbnailProcess,
	}
	var resp Files
	if err := BodyToJson(url, req, &resp, drive); err != nil {
		return nil, err
	}
	return &resp, nil
}

// get user info
func GetUserInfo(drive *conf.Drive) (*UserInfo, error) {
	url := conf.Conf.AliDrive.ApiUrl + "/user/get"
	var resp UserInfo
	if err := BodyToJson(url, map[string]interface{}{}, &resp, drive); err != nil {
		return nil, err
	}
	return &resp, nil
}

// get office preview url and token
func GetOfficePreviewUrl(fileId string, drive *conf.Drive) (*OfficePreviewUrlResp, error) {
	url := conf.Conf.AliDrive.ApiUrl + "/file/get_office_preview_url"
	req := OfficePreviewUrlReq{
		AccessToken: drive.AccessToken,
		DriveId:     drive.DefaultDriveId,
		FileId:      fileId,
	}
	var resp OfficePreviewUrlResp
	if err := BodyToJson(url, req, &resp, drive); err != nil {
		return nil, err
	}
	return &resp, nil
}

// get video preview url
func GetVideoPreviewUrl(fileId string, drive *conf.Drive) (*VideoPreviewUrlResp, error) {
	url := conf.Conf.AliDrive.ApiUrl + "/databox/get_video_play_info"
	req := VideoPreviewUrlReq{
		DriveId:     drive.DefaultDriveId,
		FileId:      fileId,
		ExpireSec: 14400,
	}
	var resp VideoPreviewUrlResp
	if err := BodyToJson(url, req, &resp, drive); err != nil {
		return nil, err
	}
	return &resp, nil
}

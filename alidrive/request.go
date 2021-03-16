package alidrive

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Xhofe/alist/conf"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
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

// convert body to json
func BodyToJson(url string, req interface{}, resp RespHandle, drive *conf.Drive) error {
	if body, err := DoPost(url, req, drive.AccessToken); err != nil {
		log.Errorf("doPost出错:%s", err.Error())
		return err
	} else {
		if err = json.Unmarshal(body, &resp); err != nil {
			log.Errorf("解析json[%s]出错:%s", string(body), err.Error())
			return err
		}
	}
	if resp.IsAvailable() {
		return nil
	}
	if resp.GetCode() == conf.AccessTokenInvalid {
		resp.SetCode("")
		if RefreshToken(drive) {
			return BodyToJson(url, req, resp, drive)
		}
	}
	return fmt.Errorf(resp.GetMessage())
}

// do post request
func DoPost(url string, request interface{}, auth string) (body []byte, err error) {
	var (
		resp *http.Response
	)
	requestBody := new(bytes.Buffer)
	err = json.NewEncoder(requestBody).Encode(request)
	if err != nil {
		log.Errorf("创建requestBody出错:%s", err.Error())
	}
	req, err := http.NewRequest("POST", url, requestBody)
	log.Debugf("do_post_req:%+v", req)
	if err != nil {
		log.Errorf("创建request出错:%s", err.Error())
		return
	}
	if auth != "" {
		req.Header.Set("authorization", conf.Bearer + auth)
	}
	req.Header.Add("content-type", "application/json")
	req.Header.Add("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
	req.Header.Add("origin", "https://aliyundrive.com")
	req.Header.Add("accept", "*/*")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,en-US;q=0.5,en;q=0.3")
	req.Header.Add("Connection", "keep-alive")

	for retryCount := 3; retryCount >= 0; retryCount-- {
		if resp, err = conf.Client.Do(req); err != nil && strings.Contains(err.Error(), "timeout") {
			<-time.After(time.Second)
		} else {
			break
		}
	}
	if err != nil {
		log.Errorf("请求阿里云盘api时出错:%s", err.Error())
		return
	}
	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Errorf("读取api返回内容失败")
	}
	log.Debugf("请求返回信息:%s", string(body))
	return
}

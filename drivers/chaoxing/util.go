package template

import (
	"encoding/json"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/bitly/go-simplejson"
	"github.com/go-resty/resty/v2"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// write util func here, such as cal sign

var chaoxingClient = resty.New()

var form_login_fmt = "fid=-1&uname=%s&password=%s&t=true&forbidotherlogin=0&validate=&doubleFactorLogin=0"

var api_list_root = "https://pan-yz.chaoxing.com/opt/listres?page=1&size=%d&enc=%s"
var api_list_file = "https://pan-yz.chaoxing.com/opt/listres?puid=%s&shareid=%s&parentId=%s&page=1&size=%d&enc=%s"
var api_list_shared_root = "https://pan-yz.chaoxing.com/opt/listres?puid=0&shareid=-1&parentId=0&page=1&size=%d&enc=%s"
var api_new_folder = "https://pan-yz.chaoxing.com/opt/newfolder?parentId=%s&name=%s&puid=%s"
var api_move_file = "https://pan-yz.chaoxing.com/opt/moveres?folderid=%s_%s&resids=%s"
var api_rename = "https://pan-yz.chaoxing.com/opt/rename?resid=%s&name=%s&puid=%s"
var api_delete_file = "https://pan-yz.chaoxing.com/opt/delres?resids=%s&resourcetype=0&puids=%s"

var reg_enc_fmt = regexp.MustCompile("enc[ ]*=\"(.*)\"")

func (driver ChaoxingDrive) Login(account *model.Account) error {
	url := "https://passport2.chaoxing.com/fanyalogin"
	var resp base.Json
	var err Resp

	req_body := fmt.Sprintf(form_login_fmt, account.Username, account.Password)

	loginReq, e := chaoxingClient.R().SetBody(req_body).
		SetResult(&resp).SetError(&err).
		SetHeader("Content-Length", strconv.FormatInt(int64(len(req_body)), 10)).
		SetHeader("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8").
		SetHeader("Host", "passport2.chaoxing.com").
		SetHeader("User-Agent", " Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/95.0.4638.69 Safari/537.36").
		SetHeader("X-Requested-With", "XMLHttpRequest").
		Post(url)

	if e != nil {
		return e
	}

	account.AccessToken = ""
	for _, cookie := range loginReq.Cookies() {
		//route=9d169c0aea4b7c89fa0d073417b5645f;
		account.AccessToken += fmt.Sprintf("%s=%s; ", cookie.Name, cookie.Value)
	}

	return nil
}

func (driver ChaoxingDrive) GetEnc(account *model.Account) error {
	url := "https://pan-yz.chaoxing.com/"

	encReq, e := chaoxingClient.R().SetHeader("Cookie", account.AccessToken).Get(url)
	if e != nil {
		return e
	}

	//直接读取响应
	sresp := string(encReq.Body())
	submatch := reg_enc_fmt.FindAllStringSubmatch(sresp, 1)
	//第一次获取失败，可能是未登录
	if len(submatch) == 0 {
		e = driver.Login(account)
		if e != nil {
			return e
		}
		encReq, e = chaoxingClient.R().SetHeader("Cookie", account.AccessToken).Get(url)
		if e != nil {
			return e
		}
		sresp = string(encReq.Body())
		submatch = reg_enc_fmt.FindAllStringSubmatch(sresp, 1)
		if len(submatch) == 0 {
			account.Status = "failed"
			return fmt.Errorf("登录失败，服务器返回信息：%s", sresp)
		}
	}
	enc := submatch[0][1]
	account.AccessSecret = enc
	account.Status = "work"
	return nil
}

func parseFileId(fileId string)(string,string,string){
	//按规则解析 id 号
	fileIdInfo := strings.Split(fileId, "_")
	if len(fileIdInfo) == 3 {
		fileId = fileIdInfo[0]
		filePuid := fileIdInfo[1]
		fileShareid := fileIdInfo[2]
		return fileId,filePuid,fileShareid
	}
	return fileId,"",""
}

func (driver ChaoxingDrive) ListFile(folder_id string, account *model.Account) ([]model.File, error) {
	var url string
	folder_id, folder_puid, folder_shareid := parseFileId(folder_id)
	if folder_puid != "" {
		if folder_id == "0" {
			//访问“共享给我的文件夹”
			url = fmt.Sprintf(api_list_shared_root, account.Limit, account.AccessSecret)
		} else {
			//访问其他目录
			url = fmt.Sprintf(api_list_file, folder_puid, folder_shareid, folder_id, account.Limit, account.AccessSecret)
		}
	}else {
		//id无法解析为三段，应当是访问根目录（此时为 ""）
		url = fmt.Sprintf(api_list_root, account.Limit, account.AccessSecret)
	}

	listFileReq, e := chaoxingClient.R().SetHeader("Cookie", account.AccessToken).Post(url)
	resp, e := simplejson.NewJson(listFileReq.Body())
	if e != nil || resp == nil {
		return nil, e
	}

	files := make([]model.File, 0)
	array, _ := resp.Get("list").Array()

	for _, file := range array {
		var f = model.File{}
		file_ := file.(map[string]interface{})
		//f.Id = file_["id"].(string)
		f.Id = fmt.Sprintf("%s_%s_%s", file_["id"].(string), file_["puid"].(json.Number).String(), file_["shareid"].(json.Number).String())
		f.Name = file_["name"].(string)
		f_server_type, _ := file_["type"].(json.Number).Int64()
		if f_server_type != TYPE_CX_SHARED_ROOT {
			f.Size, _ = file_["filesize"].(json.Number).Int64()
		}
		// 为文件分配类型
		switch f_server_type {
		case TYPE_CX_FILE:
			{
				f.Type = utils.GetFileType(file_["suffix"].(string))
			}
		case TYPE_CX_FOLDER:
			f.Type = conf.FOLDER
		case TYPE_CX_SHARED_ROOT:
			f.Type = conf.FOLDER
		}
		modifyDate, e := time.Parse("2006-01-02 15:04:05", file_["modifyDate"].(string))
		if e == nil {
			f.UpdatedAt = &modifyDate
		}
		f.Thumbnail = file_["thumbnail"].(string)
		files = append(files, f)
	}
	return files, nil
}

func (driver ChaoxingDrive) Mkdir(parentFolderId string,newFolderName string, account *model.Account) error {
	// file.Id = "581429863022592000_142134055_922191"
	fileId, puid, _ := parseFileId(parentFolderId)
	// https://pan-yz.chaoxing.com/opt/newfolder?parentId=205255741446029312&name=test&puid=54351295
	url := fmt.Sprintf(api_new_folder, fileId, newFolderName, puid)
	_, err := chaoxingClient.R().SetHeader("Cookie", account.AccessToken).Post(url)
	return err
}

func (driver ChaoxingDrive) Mv(srcFileId string, dstFolderId string, account *model.Account) error {
	// https://pan-yz.chaoxing.com/opt/moveres?folderid=502966447562248192_142134055&resids=534433141663821824
	srcFid, _, _ := parseFileId(srcFileId)
	dstFid, dstPuid, _ := parseFileId(dstFolderId)
	url := fmt.Sprintf(api_move_file,dstFid,dstPuid,srcFid)
	_, err  := chaoxingClient.R().SetHeader("Cookie", account.AccessToken).Post(url)
	return err
}

func (driver ChaoxingDrive) Ren(srcFileId string, fileName string, account *model.Account) error {
	// https://pan-yz.chaoxing.com/opt/rename?resid=762263362701209600&name=test.pdf&puid=54351295
	srcFid, srcPuid, _ := parseFileId(srcFileId)
	url := fmt.Sprintf(api_rename,srcFid,fileName,srcPuid)
	_, err  := chaoxingClient.R().SetHeader("Cookie", account.AccessToken).Post(url)
	return err
}

func (driver ChaoxingDrive) Rm(srcFileId string, account *model.Account) error {
	// https://pan-yz.chaoxing.com/opt/delres?resids=762268051373813760&resourcetype=0&puids=54351295
	srcFid, srcPuid, _ := parseFileId(srcFileId)
	url := fmt.Sprintf(api_delete_file,srcFid,srcPuid)
	_, err  := chaoxingClient.R().SetHeader("Cookie", account.AccessToken).Post(url)
	return err
}

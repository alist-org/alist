package chaoxing

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type Resp struct {
	Result int `json:"result"`
}

type UserAuth struct {
	GroupAuth struct {
		AddData                 int    `json:"addData"`
		AddDataFolder           int    `json:"addDataFolder"`
		AddLebel                int    `json:"addLebel"`
		AddManager              int    `json:"addManager"`
		AddMem                  int    `json:"addMem"`
		AddTopicFolder          int    `json:"addTopicFolder"`
		AnonymousAddReply       int    `json:"anonymousAddReply"`
		AnonymousAddTopic       int    `json:"anonymousAddTopic"`
		BatchOperation          int    `json:"batchOperation"`
		DelData                 int    `json:"delData"`
		DelDataFolder           int    `json:"delDataFolder"`
		DelMem                  int    `json:"delMem"`
		DelTopicFolder          int    `json:"delTopicFolder"`
		Dismiss                 int    `json:"dismiss"`
		ExamEnc                 string `json:"examEnc"`
		GroupChat               int    `json:"groupChat"`
		IsShowCircleChatButton  int    `json:"isShowCircleChatButton"`
		IsShowCircleCloudButton int    `json:"isShowCircleCloudButton"`
		IsShowCompanyButton     int    `json:"isShowCompanyButton"`
		Join                    int    `json:"join"`
		MemberShowRankSet       int    `json:"memberShowRankSet"`
		ModifyDataFolder        int    `json:"modifyDataFolder"`
		ModifyExpose            int    `json:"modifyExpose"`
		ModifyName              int    `json:"modifyName"`
		ModifyShowPic           int    `json:"modifyShowPic"`
		ModifyTopicFolder       int    `json:"modifyTopicFolder"`
		ModifyVisibleState      int    `json:"modifyVisibleState"`
		OnlyMgrScoreSet         int    `json:"onlyMgrScoreSet"`
		Quit                    int    `json:"quit"`
		SendNotice              int    `json:"sendNotice"`
		ShowActivityManage      int    `json:"showActivityManage"`
		ShowActivitySet         int    `json:"showActivitySet"`
		ShowAttentionSet        int    `json:"showAttentionSet"`
		ShowAutoClearStatus     int    `json:"showAutoClearStatus"`
		ShowBarcode             int    `json:"showBarcode"`
		ShowChatRoomSet         int    `json:"showChatRoomSet"`
		ShowCircleActivitySet   int    `json:"showCircleActivitySet"`
		ShowCircleSet           int    `json:"showCircleSet"`
		ShowCmem                int    `json:"showCmem"`
		ShowDataFolder          int    `json:"showDataFolder"`
		ShowDelReason           int    `json:"showDelReason"`
		ShowForward             int    `json:"showForward"`
		ShowGroupChat           int    `json:"showGroupChat"`
		ShowGroupChatSet        int    `json:"showGroupChatSet"`
		ShowGroupSquareSet      int    `json:"showGroupSquareSet"`
		ShowLockAddSet          int    `json:"showLockAddSet"`
		ShowManager             int    `json:"showManager"`
		ShowManagerIdentitySet  int    `json:"showManagerIdentitySet"`
		ShowNeedDelReasonSet    int    `json:"showNeedDelReasonSet"`
		ShowNotice              int    `json:"showNotice"`
		ShowOnlyManagerReplySet int    `json:"showOnlyManagerReplySet"`
		ShowRank                int    `json:"showRank"`
		ShowRank2               int    `json:"showRank2"`
		ShowRecycleBin          int    `json:"showRecycleBin"`
		ShowReplyByClass        int    `json:"showReplyByClass"`
		ShowReplyNeedCheck      int    `json:"showReplyNeedCheck"`
		ShowSignbanSet          int    `json:"showSignbanSet"`
		ShowSpeechSet           int    `json:"showSpeechSet"`
		ShowTopicCheck          int    `json:"showTopicCheck"`
		ShowTopicNeedCheck      int    `json:"showTopicNeedCheck"`
		ShowTransferSet         int    `json:"showTransferSet"`
	} `json:"groupAuth"`
	OperationAuth struct {
		Add                int `json:"add"`
		AddTopicToFolder   int `json:"addTopicToFolder"`
		ChoiceSet          int `json:"choiceSet"`
		DelTopicFromFolder int `json:"delTopicFromFolder"`
		Delete             int `json:"delete"`
		Reply              int `json:"reply"`
		ScoreSet           int `json:"scoreSet"`
		TopSet             int `json:"topSet"`
		Update             int `json:"update"`
	} `json:"operationAuth"`
}

// 手机端学习通上传的文件的json内容(content字段)与网页端上传的有所不同
// 网页端json `"puid": 54321, "size": 12345`
// 手机端json `"puid": "54321". "size": "12345"`
type int_str int

// json 字符串数字和纯数字解析
func (ios *int_str) UnmarshalJSON(data []byte) error {
	intValue, err := strconv.Atoi(string(bytes.Trim(data, "\"")))
	if err != nil {
		return err
	}
	*ios = int_str(intValue)
	return nil
}

type File struct {
	Cataid  int `json:"cataid"`
	Cfid    int `json:"cfid"`
	Content struct {
		Cfid             int     `json:"cfid"`
		Pid              int     `json:"pid"`
		FolderName       string  `json:"folderName"`
		ShareType        int     `json:"shareType"`
		Preview          string  `json:"preview"`
		Filetype         string  `json:"filetype"`
		PreviewURL       string  `json:"previewUrl"`
		IsImg            bool    `json:"isImg"`
		ParentPath       string  `json:"parentPath"`
		Icon             string  `json:"icon"`
		Suffix           string  `json:"suffix"`
		Duration         int     `json:"duration"`
		Pantype          string  `json:"pantype"`
		Puid             int_str `json:"puid"`
		Filepath         string  `json:"filepath"`
		Crc              string  `json:"crc"`
		Isfile           bool    `json:"isfile"`
		Residstr         string  `json:"residstr"`
		ObjectID         string  `json:"objectId"`
		Extinfo          string  `json:"extinfo"`
		Thumbnail        string  `json:"thumbnail"`
		Creator          int     `json:"creator"`
		ResTypeValue     int     `json:"resTypeValue"`
		UploadDateFormat string  `json:"uploadDateFormat"`
		DisableOpt       bool    `json:"disableOpt"`
		DownPath         string  `json:"downPath"`
		Sort             int     `json:"sort"`
		Topsort          int     `json:"topsort"`
		Restype          string  `json:"restype"`
		Size             int_str `json:"size"`
		UploadDate       int64   `json:"uploadDate"`
		FileSize         string  `json:"fileSize"`
		Name             string  `json:"name"`
		FileID           string  `json:"fileId"`
	} `json:"content"`
	CreatorID  int    `json:"creatorId"`
	DesID      string `json:"des_id"`
	ID         int    `json:"id"`
	Inserttime int64  `json:"inserttime"`
	Key        string `json:"key"`
	Norder     int    `json:"norder"`
	OwnerID    int    `json:"ownerId"`
	OwnerType  int    `json:"ownerType"`
	Path       string `json:"path"`
	Rid        int    `json:"rid"`
	Status     int    `json:"status"`
	Topsign    int    `json:"topsign"`
}

type ListFileResp struct {
	Msg      string   `json:"msg"`
	Result   int      `json:"result"`
	Status   bool     `json:"status"`
	UserAuth UserAuth `json:"userAuth"`
	List     []File   `json:"list"`
}

type DownResp struct {
	Msg        string `json:"msg"`
	Duration   int    `json:"duration"`
	Download   string `json:"download"`
	FileStatus string `json:"fileStatus"`
	URL        string `json:"url"`
	Status     bool   `json:"status"`
}

type UploadDataRsp struct {
	Result int `json:"result"`
	Msg    struct {
		Puid  int    `json:"puid"`
		Token string `json:"token"`
	} `json:"msg"`
}

type UploadFileDataRsp struct {
	Result   bool   `json:"result"`
	Msg      string `json:"msg"`
	Crc      string `json:"crc"`
	ObjectID string `json:"objectId"`
	Resid    int64  `json:"resid"`
	Puid     int    `json:"puid"`
	Data     struct {
		DisableOpt       bool   `json:"disableOpt"`
		Resid            int64  `json:"resid"`
		Crc              string `json:"crc"`
		Puid             int    `json:"puid"`
		Isfile           bool   `json:"isfile"`
		Pantype          string `json:"pantype"`
		Size             int    `json:"size"`
		Name             string `json:"name"`
		ObjectID         string `json:"objectId"`
		Restype          string `json:"restype"`
		UploadDate       int64  `json:"uploadDate"`
		ModifyDate       int64  `json:"modifyDate"`
		UploadDateFormat string `json:"uploadDateFormat"`
		Residstr         string `json:"residstr"`
		Suffix           string `json:"suffix"`
		Preview          string `json:"preview"`
		Thumbnail        string `json:"thumbnail"`
		Creator          int    `json:"creator"`
		Duration         int    `json:"duration"`
		IsImg            bool   `json:"isImg"`
		PreviewURL       string `json:"previewUrl"`
		Filetype         string `json:"filetype"`
		Filepath         string `json:"filepath"`
		Sort             int    `json:"sort"`
		Topsort          int    `json:"topsort"`
		ResTypeValue     int    `json:"resTypeValue"`
		Extinfo          string `json:"extinfo"`
	} `json:"data"`
}

type UploadDoneParam struct {
	Cataid string `json:"cataid"`
	Key    string `json:"key"`
	Param  struct {
		DisableOpt       bool   `json:"disableOpt"`
		Resid            int64  `json:"resid"`
		Crc              string `json:"crc"`
		Puid             int    `json:"puid"`
		Isfile           bool   `json:"isfile"`
		Pantype          string `json:"pantype"`
		Size             int    `json:"size"`
		Name             string `json:"name"`
		ObjectID         string `json:"objectId"`
		Restype          string `json:"restype"`
		UploadDate       int64  `json:"uploadDate"`
		ModifyDate       int64  `json:"modifyDate"`
		UploadDateFormat string `json:"uploadDateFormat"`
		Residstr         string `json:"residstr"`
		Suffix           string `json:"suffix"`
		Preview          string `json:"preview"`
		Thumbnail        string `json:"thumbnail"`
		Creator          int    `json:"creator"`
		Duration         int    `json:"duration"`
		IsImg            bool   `json:"isImg"`
		PreviewURL       string `json:"previewUrl"`
		Filetype         string `json:"filetype"`
		Filepath         string `json:"filepath"`
		Sort             int    `json:"sort"`
		Topsort          int    `json:"topsort"`
		ResTypeValue     int    `json:"resTypeValue"`
		Extinfo          string `json:"extinfo"`
	} `json:"param"`
}

func fileToObj(f File) *model.Object {
	if len(f.Content.FolderName) > 0 {
		return &model.Object{
			ID:       fmt.Sprintf("%d", f.ID),
			Name:     f.Content.FolderName,
			Size:     0,
			Modified: time.UnixMilli(f.Inserttime),
			IsFolder: true,
		}
	}
	paserTime := time.UnixMilli(f.Content.UploadDate)
	return &model.Object{
		ID:       fmt.Sprintf("%d$%s", f.ID, f.Content.FileID),
		Name:     f.Content.Name,
		Size:     int64(f.Content.Size),
		Modified: paserTime,
		IsFolder: false,
	}
}

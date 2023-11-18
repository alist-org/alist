package vtencent

import (
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type RespErr struct {
	Code    string `json:"Code"`
	Message string `json:"Message"`
}

type Reqfiles struct {
	ScrollToken string `json:"ScrollToken"`
	Text        string `json:"Text"`
	Offset      int    `json:"Offset"`
	Limit       int    `json:"Limit"`
	Sort        struct {
		Field string `json:"Field"`
		Order string `json:"Order"`
	} `json:"Sort"`
	CreateTimeRanges []any `json:"CreateTimeRanges"`
	MaterialTypes    []any `json:"MaterialTypes"`
	ReviewStatuses   []any `json:"ReviewStatuses"`
	Tags             []any `json:"Tags"`
	SearchScopes     []struct {
		Owner struct {
			Type string `json:"Type"`
			ID   string `json:"Id"`
		} `json:"Owner"`
		ClassID        int  `json:"ClassId"`
		SearchOneDepth bool `json:"SearchOneDepth"`
	} `json:"SearchScopes"`
}

type File struct {
	Type      string `json:"Type"`
	ClassInfo struct {
		ClassID     int       `json:"ClassId"`
		Name        string    `json:"Name"`
		UpdateTime  time.Time `json:"UpdateTime"`
		CreateTime  time.Time `json:"CreateTime"`
		FileInboxID string    `json:"FileInboxId"`
		Owner       struct {
			Type string `json:"Type"`
			ID   string `json:"Id"`
		} `json:"Owner"`
		ClassPath      string `json:"ClassPath"`
		ParentClassID  int    `json:"ParentClassId"`
		AttachmentInfo struct {
			SubClassCount int   `json:"SubClassCount"`
			MaterialCount int   `json:"MaterialCount"`
			Size          int64 `json:"Size"`
		} `json:"AttachmentInfo"`
		ClassPreviewURLSet []string `json:"ClassPreviewUrlSet"`
	} `json:"ClassInfo"`
	MaterialInfo struct {
		BasicInfo struct {
			MaterialID             string    `json:"MaterialId"`
			MaterialType           string    `json:"MaterialType"`
			Name                   string    `json:"Name"`
			CreateTime             time.Time `json:"CreateTime"`
			UpdateTime             time.Time `json:"UpdateTime"`
			ClassPath              string    `json:"ClassPath"`
			ClassID                int       `json:"ClassId"`
			TagInfoSet             []any     `json:"TagInfoSet"`
			TagSet                 []any     `json:"TagSet"`
			PreviewURL             string    `json:"PreviewUrl"`
			MediaURL               string    `json:"MediaUrl"`
			UnifiedMediaPreviewURL string    `json:"UnifiedMediaPreviewUrl"`
			Owner                  struct {
				Type string `json:"Type"`
				ID   string `json:"Id"`
			} `json:"Owner"`
			PermissionSet        any    `json:"PermissionSet"`
			PermissionInfoSet    []any  `json:"PermissionInfoSet"`
			TfUID                string `json:"TfUid"`
			GroupID              string `json:"GroupId"`
			VersionMaterialIDSet []any  `json:"VersionMaterialIdSet"`
			FileType             string `json:"FileType"`
			CmeMaterialPlayList  []any  `json:"CmeMaterialPlayList"`
			Status               string `json:"Status"`
			DownloadSwitch       string `json:"DownloadSwitch"`
		} `json:"BasicInfo"`
		MediaInfo struct {
			Width          int     `json:"Width"`
			Height         int     `json:"Height"`
			Size           int     `json:"Size"`
			Duration       float64 `json:"Duration"`
			Fps            int     `json:"Fps"`
			BitRate        int     `json:"BitRate"`
			Codec          string  `json:"Codec"`
			MediaType      string  `json:"MediaType"`
			FavoriteStatus string  `json:"FavoriteStatus"`
		} `json:"MediaInfo"`
		MaterialStatus struct {
			ContentReviewStatus          string `json:"ContentReviewStatus"`
			EditorUsableStatus           string `json:"EditorUsableStatus"`
			UnifiedPreviewStatus         string `json:"UnifiedPreviewStatus"`
			EditPreviewImageSpiritStatus string `json:"EditPreviewImageSpiritStatus"`
			TranscodeStatus              string `json:"TranscodeStatus"`
			AdaptiveStreamingStatus      string `json:"AdaptiveStreamingStatus"`
			StreamConnectable            string `json:"StreamConnectable"`
			AiAnalysisStatus             string `json:"AiAnalysisStatus"`
			AiRecognitionStatus          string `json:"AiRecognitionStatus"`
		} `json:"MaterialStatus"`
		ImageMaterial struct {
			Height      int    `json:"Height"`
			Width       int    `json:"Width"`
			Size        int    `json:"Size"`
			MaterialURL string `json:"MaterialUrl"`
			Resolution  string `json:"Resolution"`
			VodFileID   string `json:"VodFileId"`
			OriginalURL string `json:"OriginalUrl"`
		} `json:"ImageMaterial"`
		VideoMaterial struct {
			MetaData struct {
				Size               int     `json:"Size"`
				Container          string  `json:"Container"`
				Bitrate            int     `json:"Bitrate"`
				Height             int     `json:"Height"`
				Width              int     `json:"Width"`
				Duration           float64 `json:"Duration"`
				Rotate             int     `json:"Rotate"`
				VideoStreamInfoSet []struct {
					Bitrate int    `json:"Bitrate"`
					Height  int    `json:"Height"`
					Width   int    `json:"Width"`
					Codec   string `json:"Codec"`
					Fps     int    `json:"Fps"`
				} `json:"VideoStreamInfoSet"`
				AudioStreamInfoSet []struct {
					Bitrate      int    `json:"Bitrate"`
					SamplingRate int    `json:"SamplingRate"`
					Codec        string `json:"Codec"`
				} `json:"AudioStreamInfoSet"`
			} `json:"MetaData"`
			ImageSpriteInfo    any    `json:"ImageSpriteInfo"`
			MaterialURL        string `json:"MaterialUrl"`
			CoverURL           string `json:"CoverUrl"`
			Resolution         string `json:"Resolution"`
			VodFileID          string `json:"VodFileId"`
			OriginalURL        string `json:"OriginalUrl"`
			AudioWaveformURL   string `json:"AudioWaveformUrl"`
			SubtitleURL        string `json:"SubtitleUrl"`
			TranscodeInfoSet   []any  `json:"TranscodeInfoSet"`
			ImageSpriteInfoSet []any  `json:"ImageSpriteInfoSet"`
		} `json:"VideoMaterial"`
	} `json:"MaterialInfo"`
}

type RspFiles struct {
	Code           string `json:"Code"`
	Message        string `json:"Message"`
	EnglishMessage string `json:"EnglishMessage"`
	Data           struct {
		TotalCount      int    `json:"TotalCount"`
		ResourceInfoSet []File `json:"ResourceInfoSet"`
		ScrollToken     string `json:"ScrollToken"`
	} `json:"Data"`
}

type RspDown struct {
	Code           string `json:"Code"`
	Message        string `json:"Message"`
	EnglishMessage string `json:"EnglishMessage"`
	Data           struct {
		DownloadURLInfoSet []struct {
			MaterialID  string `json:"MaterialId"`
			DownloadURL string `json:"DownloadUrl"`
		} `json:"DownloadUrlInfoSet"`
	} `json:"Data"`
}

type RspCreatrMaterial struct {
	Code           string `json:"Code"`
	Message        string `json:"Message"`
	EnglishMessage string `json:"EnglishMessage"`
	Data           struct {
		UploadContext string `json:"UploadContext"`
		VodUploadSign string `json:"VodUploadSign"`
		QuickUpload   bool   `json:"QuickUpload"`
	} `json:"Data"`
}

type RspApplyUploadUGC struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Video struct {
			StorageSignature string `json:"storageSignature"`
			StoragePath      string `json:"storagePath"`
		} `json:"video"`
		StorageAppID    int    `json:"storageAppId"`
		StorageBucket   string `json:"storageBucket"`
		StorageRegion   string `json:"storageRegion"`
		StorageRegionV5 string `json:"storageRegionV5"`
		Domain          string `json:"domain"`
		VodSessionKey   string `json:"vodSessionKey"`
		TempCertificate struct {
			SecretID    string `json:"secretId"`
			SecretKey   string `json:"secretKey"`
			Token       string `json:"token"`
			ExpiredTime int    `json:"expiredTime"`
		} `json:"tempCertificate"`
		AppID                     int    `json:"appId"`
		Timestamp                 int    `json:"timestamp"`
		StorageRegionV50          string `json:"StorageRegionV5"`
		MiniProgramAccelerateHost string `json:"MiniProgramAccelerateHost"`
	} `json:"data"`
}

type RspCommitUploadUGC struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Video struct {
			URL           string `json:"url"`
			VerifyContent string `json:"verify_content"`
		} `json:"video"`
		FileID string `json:"fileId"`
	} `json:"data"`
}

type RspFinishUpload struct {
	Code           string `json:"Code"`
	Message        string `json:"Message"`
	EnglishMessage string `json:"EnglishMessage"`
	Data           struct {
		MaterialID string `json:"MaterialId"`
	} `json:"Data"`
}

func fileToObj(f File) *model.Object {
	obj := &model.Object{}
	if f.Type == "CLASS" {
		obj.Name = f.ClassInfo.Name
		obj.ID = strconv.Itoa(f.ClassInfo.ClassID)
		obj.IsFolder = true
		obj.Modified = f.ClassInfo.CreateTime
		obj.Size = 0
	} else if f.Type == "MATERIAL" {
		obj.Name = f.MaterialInfo.BasicInfo.Name
		obj.ID = f.MaterialInfo.BasicInfo.MaterialID
		obj.IsFolder = false
		obj.Modified = f.MaterialInfo.BasicInfo.CreateTime
		obj.Size = int64(f.MaterialInfo.MediaInfo.Size)
	}
	return obj
}

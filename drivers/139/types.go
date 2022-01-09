package _39

type BaseResp struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Catalog struct {
	CatalogID       string      `json:"catalogID"`
	CatalogName     string      `json:"catalogName"`
	CatalogType     int         `json:"catalogType"`
	CreateTime      string      `json:"createTime"`
	UpdateTime      string      `json:"updateTime"`
	IsShared        bool        `json:"isShared"`
	CatalogLevel    int         `json:"catalogLevel"`
	ShareDoneeCount int         `json:"shareDoneeCount"`
	OpenType        int         `json:"openType"`
	ParentCatalogID string      `json:"parentCatalogId"`
	DirEtag         int         `json:"dirEtag"`
	Tombstoned      int         `json:"tombstoned"`
	ProxyID         interface{} `json:"proxyID"`
	Moved           int         `json:"moved"`
	IsFixedDir      int         `json:"isFixedDir"`
	IsSynced        interface{} `json:"isSynced"`
	Owner           string      `json:"owner"`
	Modifier        interface{} `json:"modifier"`
	Path            string      `json:"path"`
	ShareType       int         `json:"shareType"`
	SoftLink        interface{} `json:"softLink"`
	ExtProp1        interface{} `json:"extProp1"`
	ExtProp2        interface{} `json:"extProp2"`
	ExtProp3        interface{} `json:"extProp3"`
	ExtProp4        interface{} `json:"extProp4"`
	ExtProp5        interface{} `json:"extProp5"`
	ETagOprType     int         `json:"ETagOprType"`
}

type Content struct {
	ContentID       string      `json:"contentID"`
	ContentName     string      `json:"contentName"`
	ContentSuffix   string      `json:"contentSuffix"`
	ContentSize     int         `json:"contentSize"`
	ContentDesc     string      `json:"contentDesc"`
	ContentType     int         `json:"contentType"`
	ContentOrigin   int         `json:"contentOrigin"`
	UpdateTime      string      `json:"updateTime"`
	CommentCount    int         `json:"commentCount"`
	ThumbnailURL    string      `json:"thumbnailURL"`
	BigthumbnailURL string      `json:"bigthumbnailURL"`
	PresentURL      string      `json:"presentURL"`
	PresentLURL     string      `json:"presentLURL"`
	PresentHURL     string      `json:"presentHURL"`
	ContentTAGList  interface{} `json:"contentTAGList"`
	ShareDoneeCount int         `json:"shareDoneeCount"`
	Safestate       int         `json:"safestate"`
	Transferstate   int         `json:"transferstate"`
	IsFocusContent  int         `json:"isFocusContent"`
	UpdateShareTime interface{} `json:"updateShareTime"`
	UploadTime      string      `json:"uploadTime"`
	OpenType        int         `json:"openType"`
	AuditResult     int         `json:"auditResult"`
	ParentCatalogID string      `json:"parentCatalogId"`
	Channel         string      `json:"channel"`
	GeoLocFlag      string      `json:"geoLocFlag"`
	Digest          string      `json:"digest"`
	Version         string      `json:"version"`
	FileEtag        string      `json:"fileEtag"`
	FileVersion     string      `json:"fileVersion"`
	Tombstoned      int         `json:"tombstoned"`
	ProxyID         string      `json:"proxyID"`
	Moved           int         `json:"moved"`
	MidthumbnailURL string      `json:"midthumbnailURL"`
	Owner           string      `json:"owner"`
	Modifier        string      `json:"modifier"`
	ShareType       int         `json:"shareType"`
	ExtInfo         struct {
		Uploader string `json:"uploader"`
		Address  string `json:"address"`
	} `json:"extInfo"`
	Exif struct {
		CreateTime    string      `json:"createTime"`
		Longitude     interface{} `json:"longitude"`
		Latitude      interface{} `json:"latitude"`
		LocalSaveTime interface{} `json:"localSaveTime"`
	} `json:"exif"`
	CollectionFlag interface{} `json:"collectionFlag"`
	TreeInfo       interface{} `json:"treeInfo"`
	IsShared       bool        `json:"isShared"`
	ETagOprType    int         `json:"ETagOprType"`
}

type GetDiskResp struct {
	BaseResp
	Data struct {
		Result struct {
			ResultCode string      `json:"resultCode"`
			ResultDesc interface{} `json:"resultDesc"`
		} `json:"result"`
		GetDiskResult struct {
			ParentCatalogID string    `json:"parentCatalogID"`
			NodeCount       int       `json:"nodeCount"`
			CatalogList     []Catalog `json:"catalogList"`
			ContentList     []Content `json:"contentList"`
			IsCompleted     int       `json:"isCompleted"`
		} `json:"getDiskResult"`
	} `json:"data"`
}

type UploadResp struct {
	BaseResp
	Data struct {
		Result struct {
			ResultCode string      `json:"resultCode"`
			ResultDesc interface{} `json:"resultDesc"`
		} `json:"result"`
		UploadResult struct {
			UploadTaskID     string `json:"uploadTaskID"`
			RedirectionURL   string `json:"redirectionUrl"`
			NewContentIDList []struct {
				ContentID     string `json:"contentID"`
				ContentName   string `json:"contentName"`
				IsNeedUpload  string `json:"isNeedUpload"`
				FileEtag      int64  `json:"fileEtag"`
				FileVersion   int64  `json:"fileVersion"`
				OverridenFlag int    `json:"overridenFlag"`
			} `json:"newContentIDList"`
			CatalogIDList interface{} `json:"catalogIDList"`
			IsSlice       interface{} `json:"isSlice"`
		} `json:"uploadResult"`
	} `json:"data"`
}

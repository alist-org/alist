package qbittorrent

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/alist-org/alist/v3/pkg/utils"
)

type Client interface {
	AddFromLink(link string, savePath string, id string) error
	GetInfo(id string) (TorrentInfo, error)
	GetFiles(id string) ([]FileInfo, error)
	Delete(id string, deleteFiles bool) error
}

type client struct {
	url    *url.URL
	client http.Client
	Client
}

func New(webuiUrl string) (Client, error) {
	u, err := url.Parse(webuiUrl)
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	var c = &client{
		url:    u,
		client: http.Client{Jar: jar},
	}

	err = c.checkAuthorization()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *client) checkAuthorization() error {
	// check authorization
	if c.authorized() {
		return nil
	}

	// check authorization after logging in
	err := c.login()
	if err != nil {
		return err
	}
	if c.authorized() {
		return nil
	}
	return errors.New("unauthorized qbittorrent url")
}

func (c *client) authorized() bool {
	resp, err := c.post("/api/v2/app/version", nil)
	if err != nil {
		return false
	}
	return resp.StatusCode == 200 // the status code will be 403 if not authorized
}

func (c *client) login() error {
	// prepare HTTP request
	v := url.Values{}
	v.Set("username", c.url.User.Username())
	passwd, _ := c.url.User.Password()
	v.Set("password", passwd)
	resp, err := c.post("/api/v2/auth/login", v)
	if err != nil {
		return err
	}

	// check result
	body := make([]byte, 2)
	_, err = resp.Body.Read(body)
	if err != nil {
		return err
	}
	if string(body) != "Ok" {
		return errors.New("failed to login into qBittorrent webui with url: " + c.url.String())
	}
	return nil
}

func (c *client) post(path string, data url.Values) (*http.Response, error) {
	u := c.url.JoinPath(path)
	u.User = nil // remove userinfo for requests

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader([]byte(data.Encode())))
	if err != nil {
		return nil, err
	}
	if data != nil {
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.Cookies() != nil {
		c.client.Jar.SetCookies(u, resp.Cookies())
	}
	return resp, nil
}

func (c *client) AddFromLink(link string, savePath string, id string) error {
	err := c.checkAuthorization()
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)

	addField := func(name string, value string) {
		if err != nil {
			return
		}
		err = writer.WriteField(name, value)
	}
	addField("urls", link)
	addField("savepath", savePath)
	addField("tags", "alist-"+id)
	addField("autoTMM", "false")
	if err != nil {
		return err
	}

	err = writer.Close()
	if err != nil {
		return err
	}

	u := c.url.JoinPath("/api/v2/torrents/add")
	u.User = nil // remove userinfo for requests
	req, err := http.NewRequest("POST", u.String(), buf)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}

	// check result
	body := make([]byte, 2)
	_, err = resp.Body.Read(body)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 || string(body) != "Ok" {
		return errors.New("failed to add qBittorrent task: " + link)
	}
	return nil
}

type TorrentStatus string

const (
	ERROR              TorrentStatus = "error"
	MISSINGFILES       TorrentStatus = "missingFiles"
	UPLOADING          TorrentStatus = "uploading"
	PAUSEDUP           TorrentStatus = "pausedUP"
	QUEUEDUP           TorrentStatus = "queuedUP"
	STALLEDUP          TorrentStatus = "stalledUP"
	CHECKINGUP         TorrentStatus = "checkingUP"
	FORCEDUP           TorrentStatus = "forcedUP"
	ALLOCATING         TorrentStatus = "allocating"
	DOWNLOADING        TorrentStatus = "downloading"
	METADL             TorrentStatus = "metaDL"
	PAUSEDDL           TorrentStatus = "pausedDL"
	QUEUEDDL           TorrentStatus = "queuedDL"
	STALLEDDL          TorrentStatus = "stalledDL"
	CHECKINGDL         TorrentStatus = "checkingDL"
	FORCEDDL           TorrentStatus = "forcedDL"
	CHECKINGRESUMEDATA TorrentStatus = "checkingResumeData"
	MOVING             TorrentStatus = "moving"
	UNKNOWN            TorrentStatus = "unknown"
)

// https://github.com/DGuang21/PTGo/blob/main/app/client/client_distributer.go
type TorrentInfo struct {
	AddedOn           int           `json:"added_on"`           // 将 torrent 添加到客户端的时间（Unix Epoch）
	AmountLeft        int64         `json:"amount_left"`        // 剩余大小（字节）
	AutoTmm           bool          `json:"auto_tmm"`           // 此 torrent 是否由 Automatic Torrent Management 管理
	Availability      float64       `json:"availability"`       // 当前百分比
	Category          string        `json:"category"`           //
	Completed         int64         `json:"completed"`          // 完成的传输数据量（字节）
	CompletionOn      int           `json:"completion_on"`      // Torrent 完成的时间（Unix Epoch）
	ContentPath       string        `json:"content_path"`       // torrent 内容的绝对路径（多文件 torrent 的根路径，单文件 torrent 的绝对文件路径）
	DlLimit           int           `json:"dl_limit"`           // Torrent 下载速度限制（字节/秒）
	Dlspeed           int           `json:"dlspeed"`            // Torrent 下载速度（字节/秒）
	Downloaded        int64         `json:"downloaded"`         // 已经下载大小
	DownloadedSession int64         `json:"downloaded_session"` // 此会话下载的数据量
	Eta               int           `json:"eta"`                //
	FLPiecePrio       bool          `json:"f_l_piece_prio"`     // 如果第一个最后一块被优先考虑，则为true
	ForceStart        bool          `json:"force_start"`        // 如果为此 torrent 启用了强制启动，则为true
	Hash              string        `json:"hash"`               //
	LastActivity      int           `json:"last_activity"`      // 上次活跃的时间（Unix Epoch）
	MagnetURI         string        `json:"magnet_uri"`         // 与此 torrent 对应的 Magnet URI
	MaxRatio          float64       `json:"max_ratio"`          // 种子/上传停止种子前的最大共享比率
	MaxSeedingTime    int           `json:"max_seeding_time"`   // 停止种子种子前的最长种子时间（秒）
	Name              string        `json:"name"`               //
	NumComplete       int           `json:"num_complete"`       //
	NumIncomplete     int           `json:"num_incomplete"`     //
	NumLeechs         int           `json:"num_leechs"`         // 连接到的 leechers 的数量
	NumSeeds          int           `json:"num_seeds"`          // 连接到的种子数
	Priority          int           `json:"priority"`           // 速度优先。如果队列被禁用或 torrent 处于种子模式，则返回 -1
	Progress          float64       `json:"progress"`           // 进度
	Ratio             float64       `json:"ratio"`              // Torrent 共享比率
	RatioLimit        int           `json:"ratio_limit"`        //
	SavePath          string        `json:"save_path"`
	SeedingTime       int           `json:"seeding_time"`       // Torrent 完成用时（秒）
	SeedingTimeLimit  int           `json:"seeding_time_limit"` // max_seeding_time
	SeenComplete      int           `json:"seen_complete"`      // 上次 torrent 完成的时间
	SeqDl             bool          `json:"seq_dl"`             // 如果启用顺序下载，则为true
	Size              int64         `json:"size"`               //
	State             TorrentStatus `json:"state"`              // 参见https://github.com/qbittorrent/qBittorrent/wiki/WebUI-API-(qBittorrent-4.1)#get-torrent-list
	SuperSeeding      bool          `json:"super_seeding"`      // 如果启用超级播种，则为true
	Tags              string        `json:"tags"`               // Torrent 的逗号连接标签列表
	TimeActive        int           `json:"time_active"`        // 总活动时间（秒）
	TotalSize         int64         `json:"total_size"`         // 此 torrent 中所有文件的总大小（字节）（包括未选择的文件）
	Tracker           string        `json:"tracker"`            // 第一个具有工作状态的tracker。如果没有tracker在工作，则返回空字符串。
	TrackersCount     int           `json:"trackers_count"`     //
	UpLimit           int           `json:"up_limit"`           // 上传限制
	Uploaded          int64         `json:"uploaded"`           // 累计上传
	UploadedSession   int64         `json:"uploaded_session"`   // 当前session累计上传
	Upspeed           int           `json:"upspeed"`            // 上传速度（字节/秒）
}

type InfoNotFoundError struct {
	Id  string
	Err error
}

func (i InfoNotFoundError) Error() string {
	return "there should be exactly one task with tag \"alist-" + i.Id + "\""
}

func NewInfoNotFoundError(id string) InfoNotFoundError {
	return InfoNotFoundError{Id: id}
}

func (c *client) GetInfo(id string) (TorrentInfo, error) {
	var infos []TorrentInfo

	err := c.checkAuthorization()
	if err != nil {
		return TorrentInfo{}, err
	}

	v := url.Values{}
	v.Set("tag", "alist-"+id)
	response, err := c.post("/api/v2/torrents/info", v)
	if err != nil {
		return TorrentInfo{}, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return TorrentInfo{}, err
	}
	err = utils.Json.Unmarshal(body, &infos)
	if err != nil {
		return TorrentInfo{}, err
	}
	if len(infos) != 1 {
		return TorrentInfo{}, NewInfoNotFoundError(id)
	}
	return infos[0], nil
}

type FileInfo struct {
	Index        int     `json:"index"`
	Name         string  `json:"name"`
	Size         int64   `json:"size"`
	Progress     float32 `json:"progress"`
	Priority     int     `json:"priority"`
	IsSeed       bool    `json:"is_seed"`
	PieceRange   []int   `json:"piece_range"`
	Availability float32 `json:"availability"`
}

func (c *client) GetFiles(id string) ([]FileInfo, error) {
	var infos []FileInfo

	err := c.checkAuthorization()
	if err != nil {
		return []FileInfo{}, err
	}

	tInfo, err := c.GetInfo(id)
	if err != nil {
		return []FileInfo{}, err
	}

	v := url.Values{}
	v.Set("hash", tInfo.Hash)
	response, err := c.post("/api/v2/torrents/files", v)
	if err != nil {
		return []FileInfo{}, err
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []FileInfo{}, err
	}
	err = utils.Json.Unmarshal(body, &infos)
	if err != nil {
		return []FileInfo{}, err
	}
	return infos, nil
}

func (c *client) Delete(id string, deleteFiles bool) error {
	err := c.checkAuthorization()
	if err != nil {
		return err
	}

	info, err := c.GetInfo(id)
	if err != nil {
		return err
	}
	v := url.Values{}
	v.Set("hashes", info.Hash)
	if deleteFiles {
		v.Set("deleteFiles", "true")
	} else {
		v.Set("deleteFiles", "false")
	}
	response, err := c.post("/api/v2/torrents/delete", v)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New("failed to delete qbittorrent task")
	}

	v = url.Values{}
	v.Set("tags", "alist-"+id)
	response, err = c.post("/api/v2/torrents/deleteTags", v)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return errors.New("failed to delete qbittorrent tag")
	}
	return nil
}

package fastwebdav

import (
	"encoding/base64"
	"encoding/json"
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	hash_extend "github.com/alist-org/alist/v3/pkg/utils/hash"
)

type Resp struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type Policy struct {
	Id       string   `json:"id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	MaxSize  int      `json:"max_size"`
	FileType []string `json:"file_type"`
}

type UploadInfo struct {
	SessionID string `json:"sessionID"`
	ChunkSize int    `json:"chunkSize"`
	Expires   int    `json:"expires"`
}

type DirectoryResp struct {
	Parent  string   `json:"parent"`
	Objects []Object `json:"objects"`
	Policy  Policy   `json:"policy"`
}

type Object struct {
	Id            string    `json:"id"`
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	Pic           string    `json:"pic"`
	Size          int       `json:"size"`
	Type          string    `json:"type"`
	Date          time.Time `json:"date"`
	CreateDate    time.Time `json:"create_date"`
	SourceEnabled bool      `json:"source_enabled"`
}

type DirectoryProp struct {
	Size int `json:"size"`
}

func objectToObj(f Object, t model.Thumbnail) *model.ObjThumb {
	return &model.ObjThumb{
		Object: model.Object{
			ID:       f.Id,
			Name:     f.Name,
			Size:     int64(f.Size),
			Modified: f.Date,
			IsFolder: f.Type == "dir",
		},
		Thumbnail: t,
	}
}

type Config struct {
	LoginCaptcha bool   `json:"loginCaptcha"`
	CaptchaType  string `json:"captcha_type"`
}

type File struct {
	Id          string `json:"id"`
	Kind        int    `json:"kind"`
	Provider    string `json:"provider"`
	Name        string `json:"name"`
	CreateTime  string `json:"create_time"`
	Sha1        string `json:"sha1"`
	Size        string `json:"size"`
	ParentId    string `json:"parent_id"`
	Oriname     string `json:"oriname"`
	DownloadUrl string `json:"download_url"`
	PlayHeaders string `json:"play_headers"`
	Password    string `json:"password"`
}

func fileToObj(f File) *model.ObjThumb {
	size, _ := strconv.ParseInt(f.Size, 10, 64)
	create_time, _ := time.Parse("2006-01-02 15:04:05", f.CreateTime)
	b, _ := json.Marshal(f)
	file_id := base64.StdEncoding.EncodeToString(b)
	return &model.ObjThumb{
		Object: model.Object{
			ID:       file_id,
			Name:     f.Name,
			Size:     size,
			Ctime:    create_time,
			Modified: create_time,
			IsFolder: f.Kind == 0,
			HashInfo: utils.NewHashInfo(hash_extend.GCID, f.Sha1),
		},
	}
}

// Node is a node in the folder tree
type Node struct {
	Url      string
	Name     string
	Level    int
	Modified int64
	Size     int64
	Children []*Node
}

func (node *Node) getByPath(paths []string) *Node {
	if len(paths) == 0 || node == nil {
		return nil
	}
	if node.Name != paths[0] {
		return nil
	}
	if len(paths) == 1 {
		return node
	}
	for _, child := range node.Children {
		tmp := child.getByPath(paths[1:])
		if tmp != nil {
			return tmp
		}
	}
	return nil
}

func (node *Node) isFile() bool {
	return node.Url != ""
}

func (node *Node) calSize() int64 {
	if node.isFile() {
		return node.Size
	}
	var size int64 = 0
	for _, child := range node.Children {
		size += child.calSize()
	}
	node.Size = size
	return size
}

func nodeToObj(node *Node, path string) (model.Obj, error) {
	if node == nil {
		return nil, errs.ObjectNotFound
	}
	return &model.Object{
		Name:     node.Name,
		Size:     node.Size,
		Modified: time.Unix(node.Modified, 0),
		IsFolder: !node.isFile(),
		Path:     path,
	}, nil
}

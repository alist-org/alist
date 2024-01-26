package dropbox

import (
	"github.com/alist-org/alist/v3/internal/model"
	"time"
)

type TokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type ErrorResp struct {
	Error struct {
		Tag string `json:".tag"`
	} `json:"error"`
	ErrorSummary string `json:"error_summary"`
}

type RefreshTokenErrorResp struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type CurrentAccountResp struct {
	RootInfo struct {
		RootNamespaceId string `json:"root_namespace_id"`
		HomeNamespaceId string `json:"home_namespace_id"`
	} `json:"root_info"`
}

type File struct {
	Tag            string    `json:".tag"`
	Name           string    `json:"name"`
	PathLower      string    `json:"path_lower"`
	PathDisplay    string    `json:"path_display"`
	ID             string    `json:"id"`
	ClientModified time.Time `json:"client_modified"`
	ServerModified time.Time `json:"server_modified"`
	Rev            string    `json:"rev"`
	Size           int       `json:"size"`
	IsDownloadable bool      `json:"is_downloadable"`
	ContentHash    string    `json:"content_hash"`
}

type ListResp struct {
	Entries []File `json:"entries"`
	Cursor  string `json:"cursor"`
	HasMore bool   `json:"has_more"`
}

type UploadCursor struct {
	Offset    int64  `json:"offset"`
	SessionID string `json:"session_id"`
}

type UploadAppendArgs struct {
	Close  bool         `json:"close"`
	Cursor UploadCursor `json:"cursor"`
}

type UploadFinishArgs struct {
	Commit struct {
		Autorename     bool   `json:"autorename"`
		Mode           string `json:"mode"`
		Mute           bool   `json:"mute"`
		Path           string `json:"path"`
		StrictConflict bool   `json:"strict_conflict"`
	} `json:"commit"`
	Cursor UploadCursor `json:"cursor"`
}

func fileToObj(f File) *model.ObjThumb {
	return &model.ObjThumb{
		Object: model.Object{
			ID:       f.ID,
			Path:     f.PathDisplay,
			Name:     f.Name,
			Size:     int64(f.Size),
			Modified: f.ServerModified,
			IsFolder: f.Tag == "folder",
		},
		Thumbnail: model.Thumbnail{},
	}
}

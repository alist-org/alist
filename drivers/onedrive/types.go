package onedrive

import (
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

type Host struct {
	Oauth string
	Api   string
}

type TokenErr struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type RespErr struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type File struct {
	Id             string               `json:"id"`
	Name           string               `json:"name"`
	Size           int64                `json:"size"`
	FileSystemInfo *FileSystemInfoFacet `json:"fileSystemInfo"`
	Url            string               `json:"@microsoft.graph.downloadUrl"`
	File           *struct {
		MimeType string `json:"mimeType"`
	} `json:"file"`
	Thumbnails []struct {
		Medium struct {
			Url string `json:"url"`
		} `json:"medium"`
	} `json:"thumbnails"`
	ParentReference struct {
		DriveId string `json:"driveId"`
	} `json:"parentReference"`
}

type Object struct {
	model.ObjThumb
	ParentID string
}

func fileToObj(f File, parentID string) *Object {
	thumb := ""
	if len(f.Thumbnails) > 0 {
		thumb = f.Thumbnails[0].Medium.Url
	}
	return &Object{
		ObjThumb: model.ObjThumb{
			Object: model.Object{
				ID:       f.Id,
				Name:     f.Name,
				Size:     f.Size,
				Modified: f.FileSystemInfo.LastModifiedDateTime,
				IsFolder: f.File == nil,
			},
			Thumbnail: model.Thumbnail{Thumbnail: thumb},
			//Url:       model.Url{Url: f.Url},
		},
		ParentID: parentID,
	}
}

type Files struct {
	Value    []File `json:"value"`
	NextLink string `json:"@odata.nextLink"`
}

// Metadata represents a request to update Metadata.
// It includes only the writeable properties.
// omitempty is intentionally included for all, per https://learn.microsoft.com/en-us/onedrive/developer/rest-api/api/driveitem_update?view=odsp-graph-online#request-body
type Metadata struct {
	Description    string               `json:"description,omitempty"`    // Provides a user-visible description of the item. Read-write. Only on OneDrive Personal. Undocumented limit of 1024 characters.
	FileSystemInfo *FileSystemInfoFacet `json:"fileSystemInfo,omitempty"` // File system information on client. Read-write.
}

// FileSystemInfoFacet contains properties that are reported by the
// device's local file system for the local version of an item. This
// facet can be used to specify the last modified date or created date
// of the item as it was on the local device.
type FileSystemInfoFacet struct {
	CreatedDateTime      time.Time `json:"createdDateTime,omitempty"`      // The UTC date and time the file was created on a client.
	LastModifiedDateTime time.Time `json:"lastModifiedDateTime,omitempty"` // The UTC date and time the file was last modified on a client.
}

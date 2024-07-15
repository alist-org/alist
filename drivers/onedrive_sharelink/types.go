package onedrive_sharelink

import (
	"strconv"
	"time"

	"github.com/alist-org/alist/v3/internal/model"
)

// FolderResp represents the structure of the folder response from the OneDrive API.
type FolderResp struct {
	// Data holds the nested structure of the response.
	Data struct {
		Legacy struct {
			RenderListData struct {
				ListData struct {
					Items []Item `json:"Row"` // Items contains the list of items in the folder.
				} `json:"ListData"`
			} `json:"renderListDataAsStream"`
		} `json:"legacy"`
	} `json:"data"`
}

// Item represents an individual item in the folder.
type Item struct {
	ObjType      string    `json:"FSObjType"`       // ObjType indicates if the item is a file or folder.
	Name         string    `json:"FileLeafRef"`     // Name is the name of the item.
	ModifiedTime time.Time `json:"Modified."`       // ModifiedTime is the last modified time of the item.
	Size         string    `json:"File_x0020_Size"` // Size is the size of the item in string format.
	Id           string    `json:"UniqueId"`        // Id is the unique identifier of the item.
}

// fileToObj converts an Item to an ObjThumb.
func fileToObj(f Item) *model.ObjThumb {
	// Convert Size from string to int64.
	size, _ := strconv.ParseInt(f.Size, 10, 64)
	// Convert ObjType from string to int.
	objtype, _ := strconv.Atoi(f.ObjType)

	// Create a new ObjThumb with the converted values.
	file := &model.ObjThumb{
		Object: model.Object{
			Name:     f.Name,
			Modified: f.ModifiedTime,
			Size:     size,
			IsFolder: objtype == 1, // Check if the item is a folder.
			ID:       f.Id,
		},
		Thumbnail: model.Thumbnail{},
	}
	return file
}

// GraphQLNEWRequest represents the structure of a new GraphQL request.
type GraphQLNEWRequest struct {
	ListData struct {
		NextHref string `json:"NextHref"` // NextHref is the link to the next set of data.
		Row      []Item `json:"Row"`      // Row contains the list of items.
	} `json:"ListData"`
}

// GraphQLRequest represents the structure of a GraphQL request.
type GraphQLRequest struct {
	Data struct {
		Legacy struct {
			RenderListDataAsStream struct {
				ListData struct {
					NextHref string `json:"NextHref"` // NextHref is the link to the next set of data.
					Row      []Item `json:"Row"`      // Row contains the list of items.
				} `json:"ListData"`
				ViewMetadata struct {
					ListViewXml string `json:"ListViewXml"` // ListViewXml contains the XML of the list view.
				} `json:"ViewMetadata"`
			} `json:"renderListDataAsStream"`
		} `json:"legacy"`
	} `json:"data"`
}

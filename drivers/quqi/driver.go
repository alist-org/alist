package quqi

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/go-resty/resty/v2"
)

type Quqi struct {
	model.Storage
	Addition
	GroupID string
}

var header = http.Header{}

func (d *Quqi) Config() driver.Config {
	return config
}

func (d *Quqi) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Quqi) Init(ctx context.Context) error {
	// 登录
	if err := d.login(); err != nil {
		return err
	}

	// (暂时仅获取私人云) 获取私人云ID
	groupResp := &GroupRes{}
	if _, err := d.request("group.quqi.com", "/v1/group/list", resty.MethodGet, nil, groupResp); err != nil {
		return err
	}
	for _, groupInfo := range groupResp.Data {
		if groupInfo == nil {
			continue
		}
		if groupInfo.Type == 2 {
			d.GroupID = strconv.Itoa(groupInfo.ID)
			break
		}
	}
	if d.GroupID == "" {
		return errs.StorageNotFound
	}

	// 设置header
	header.Set("Origin", "https://quqi.com")
	header.Set("Cookie", d.Cookie)

	return nil
}

func (d *Quqi) Drop(ctx context.Context) error {
	return nil
}

func (d *Quqi) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	var (
		listResp = &ListRes{}
		files    []model.Obj
	)

	if _, err := d.request("", "/api/dir/ls", resty.MethodPost, func(req *resty.Request) {
		req.SetFormData(map[string]string{
			"quqi_id": d.GroupID,
			"node_id": dir.GetID(),
		})
	}, listResp); err != nil {
		return nil, err
	}

	if listResp.Data == nil {
		return nil, nil
	}

	// dirs
	for _, dirInfo := range listResp.Data.Dir {
		if dirInfo == nil {
			continue
		}
		files = append(files, &model.Object{
			ID:       strconv.FormatInt(dirInfo.NodeID, 10),
			Name:     dirInfo.Name,
			Modified: time.Unix(dirInfo.UpdateTime, 0),
			Ctime:    time.Unix(dirInfo.AddTime, 0),
			IsFolder: true,
		})
	}

	// files
	for _, fileInfo := range listResp.Data.File {
		if fileInfo == nil {
			continue
		}
		if fileInfo.EXT != "" {
			fileInfo.Name = strings.Join([]string{fileInfo.Name, fileInfo.EXT}, ".")
		}

		files = append(files, &model.Object{
			ID:       strconv.FormatInt(fileInfo.NodeID, 10),
			Name:     fileInfo.Name,
			Size:     fileInfo.Size,
			Modified: time.Unix(fileInfo.UpdateTime, 0),
			Ctime:    time.Unix(fileInfo.AddTime, 0),
		})
	}

	return files, nil
}

func (d *Quqi) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	var getDocResp = &GetDocRes{}

	if _, err := d.request("", "/api/doc/getDoc", resty.MethodPost, func(req *resty.Request) {
		req.SetFormData(map[string]string{
			"quqi_id": d.GroupID,
			"node_id": file.GetID(),
		})
	}, getDocResp); err != nil {
		return nil, err
	}

	return &model.Link{
		URL:    getDocResp.Data.OriginPath,
		Header: header,
	}, nil
}

func (d *Quqi) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) (model.Obj, error) {
	var (
		makeDirRes = &MakeDirRes{}
		timeNow    = time.Now()
	)

	if _, err := d.request("", "/api/dir/mkDir", resty.MethodPost, func(req *resty.Request) {
		req.SetFormData(map[string]string{
			"quqi_id":   d.GroupID,
			"parent_id": parentDir.GetID(),
			"name":      dirName,
		})
	}, makeDirRes); err != nil {
		return nil, err
	}

	return &model.Object{
		ID:       strconv.FormatInt(makeDirRes.Data.NodeID, 10),
		Name:     dirName,
		Modified: timeNow,
		Ctime:    timeNow,
		IsFolder: true,
	}, nil
}

func (d *Quqi) Move(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	var moveRes = &MoveRes{}

	if _, err := d.request("", "/api/dir/mvDir", resty.MethodPost, func(req *resty.Request) {
		req.SetFormData(map[string]string{
			"quqi_id":        d.GroupID,
			"node_id":        dstDir.GetID(),
			"source_quqi_id": d.GroupID,
			"source_node_id": srcObj.GetID(),
		})
	}, moveRes); err != nil {
		return nil, err
	}

	return &model.Object{
		ID:       strconv.FormatInt(moveRes.Data.NodeID, 10),
		Name:     moveRes.Data.NodeName,
		Size:     srcObj.GetSize(),
		Modified: time.Now(),
		Ctime:    srcObj.CreateTime(),
		IsFolder: srcObj.IsDir(),
	}, nil
}

func (d *Quqi) Rename(ctx context.Context, srcObj model.Obj, newName string) (model.Obj, error) {
	var renameRes = &RenameRes{}

	if _, err := d.request("", "/api/dir/renameDir", resty.MethodPost, func(req *resty.Request) {
		req.SetFormData(map[string]string{
			"quqi_id": d.GroupID,
			"node_id": srcObj.GetID(),
			"rename":  newName,
		})
	}, renameRes); err != nil {
		return nil, err
	}

	return &model.Object{
		ID:       strconv.FormatInt(renameRes.Data.NodeID, 10),
		Name:     renameRes.Data.Rename,
		Size:     srcObj.GetSize(),
		Modified: time.Unix(renameRes.Data.UpdateTime, 0),
		Ctime:    srcObj.CreateTime(),
		IsFolder: srcObj.IsDir(),
	}, nil
}

func (d *Quqi) Copy(ctx context.Context, srcObj, dstDir model.Obj) (model.Obj, error) {
	// 无法从曲奇接口响应中直接获取复制后的文件信息
	if _, err := d.request("", "/api/node/copy", resty.MethodPost, nil, map[string]string{
		"quqi_id":        d.GroupID,
		"node_id":        dstDir.GetID(),
		"source_quqi_id": d.GroupID,
		"source_node_id": srcObj.GetID(),
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (d *Quqi) Remove(ctx context.Context, obj model.Obj) error {
	// 暂时不做直接删除，默认都放到回收站。直接删除方法：先调用删除接口放入回收站，在通过回收站接口删除文件
	if _, err := d.request("", "/api/node/del", resty.MethodPost, nil, map[string]string{
		"quqi_id": d.GroupID,
		"node_id": obj.GetID(),
	}); err != nil {
		return err
	}

	return nil
}

func (d *Quqi) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) (model.Obj, error) {
	// TODO upload file, optional
	//
	return nil, errs.NotImplement
}

//func (d *Template) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*Quqi)(nil)

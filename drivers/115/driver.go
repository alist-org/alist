package _115

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io"
	"os"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/deadblue/elevengo"
)

type Pan115 struct {
	model.Storage
	Addition
	agent *elevengo.Agent
}

func (d *Pan115) Config() driver.Config {
	return config
}

func (d *Pan115) GetAddition() driver.Additional {
	return d.Addition
}

func (d *Pan115) Init(ctx context.Context, storage model.Storage) error {
	d.Storage = storage
	err := utils.Json.UnmarshalFromString(d.Storage.Addition, &d.Addition)
	if err != nil {
		return err
	}
	return d.login()
}

func (d *Pan115) Drop(ctx context.Context) error {
	return nil
}

// List files in the path
// if identify files by path, need to set ID with path,like path.Join(dir.GetID(), obj.GetName())
// if identify files by id, need to set ID with corresponding id
func (d *Pan115) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	files, err := d.getFiles(dir.GetID())
	if err != nil {
		return nil, err
	}
	return utils.SliceConvert(files, func(src File) (model.Obj, error) {
		return src, nil
	})
}

// Get optional get file obj
func (d *Pan115) Get(ctx context.Context, fileId string) (model.Obj, error) {
	file := elevengo.File{}
	if err := d.agent.FileGet(fileId, &file); err != nil {
		return nil, err
	}
	return File(file), nil
}

// Link get url/filepath/reader of file
func (d *Pan115) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	pickCode := file.(File).PickCode
	ticket := elevengo.DownloadTicket{}
	link := &model.Link{}
	if file.GetSize() < int64(UploadSimplyMaxSize) {
		if err := d.agent.DownloaCreateTicketdSimply(pickCode, &ticket); err != nil {
			return nil, err
		}

		data, err := d.agent.Get(ticket.Url)
		if err != nil {
			return nil, err
		}
		link.Data = data
	} else {
		return nil, fmt.Errorf("file size is bigger than 200MB, failed to download from 115 cloud")
	}

	return link, nil
}

// MakeDir make a folder named `dirName` in `parentDir`
func (d *Pan115) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	if _, err := d.agent.DirMake(parentDir.GetID(), dirName); err != nil {
		return err
	}
	return nil
}

// Move `srcObject` to `dstDir`
func (d *Pan115) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	return d.agent.FileMove(dstDir.GetID(), srcObj.GetID())
}

// Rename rename `srcObject` to `newName`
func (d *Pan115) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	return d.agent.FileRename(srcObj.GetID(), newName)
}

// Copy `srcObject` to `dstDir`
func (d *Pan115) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return d.agent.FileCopy(dstDir.GetID(), srcObj.GetID())
}

// Remove remove `object`
func (d *Pan115) Remove(ctx context.Context, obj model.Obj) error {
	return d.agent.FileDelete(obj.GetID())
}

// Put upload `stream` to `parentDir`
func (d *Pan115) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	tempFile, err := utils.CreateTempFile(stream.GetReadCloser())
	if err != nil {
		return err
	}
	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
	}()

	if _, err = io.Copy(sha1.New(), tempFile); err != nil {
		return err
	}

	if _, err = tempFile.Seek(0, io.SeekStart); err != nil {
		return err
	}

	ts, err := tempFile.Stat()
	if err != nil {
		return err
	}

	size := ts.Size()
	if size < int64(UploadSimplyMaxSize) {
		if stream.GetSize() == 0 { // if create a new blank file, size is 0, the 115 lib will upload failed.
			size = 1
		}
		if _, err := d.agent.UploadSimply(
			dstDir.GetID(),
			stream.GetName(),
			size,
			tempFile,
		); err != nil {
			return err
		}
	} else {
		// ticket := &elevengo.UploadTicket{}
		// if err := d.agent.UploadCreateTicket(
		// 	dstDir.GetID(),
		// 	stream.GetName(),
		// 	tempFile,
		// 	ticket,
		// ); err != nil {
		// 	return err
		// }
		// if ticket.Exist {
		// 	return fmt.Errorf("file already exists")
		// }
		return fmt.Errorf("file size is bigger than 200MB, failed to upload to 115 cloud")
	}

	return nil
}

var _ driver.Driver = (*Pan115)(nil)

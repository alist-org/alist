package crypt

import (
	"context"
	"fmt"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/net"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/http_range"
	"github.com/alist-org/alist/v3/pkg/utils"
	rcCrypt "github.com/rclone/rclone/backend/crypt"
	"github.com/rclone/rclone/fs/config/configmap"
	"github.com/rclone/rclone/fs/config/obscure"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	stdpath "path"
	"path/filepath"
	"regexp"
	"strings"
)

type Crypt struct {
	model.Storage
	Addition
	cipher        *rcCrypt.Cipher
	remoteStorage driver.Driver
}

const obfuscatedPrefix = "___Obfuscated___"

func (d *Crypt) Config() driver.Config {
	return config
}

func (d *Crypt) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Crypt) Init(ctx context.Context) error {
	//obfuscate credentials if it's updated or just created
	err := d.updateObfusParm(&d.Password)
	if err != nil {
		return fmt.Errorf("failed to obfuscate password: %w", err)
	}
	err = d.updateObfusParm(&d.Salt)
	if err != nil {
		return fmt.Errorf("failed to obfuscate salt: %w", err)
	}

	isSafeExt := regexp.MustCompile(`^[.][A-Za-z0-9-_]{2,}$`).MatchString
	if !isSafeExt(d.EncryptedSuffix) {
		return fmt.Errorf("EncryptedSuffix is Illegal")
	}

	op.MustSaveDriverStorage(d)

	//need remote storage exist
	storage, err := fs.GetStorage(d.RemotePath, &fs.GetStoragesArgs{})
	if err != nil {
		return fmt.Errorf("can't find remote storage: %w", err)
	}
	d.remoteStorage = storage

	p, _ := strings.CutPrefix(d.Password, obfuscatedPrefix)
	p2, _ := strings.CutPrefix(d.Salt, obfuscatedPrefix)
	config := configmap.Simple{
		"password":                  p,
		"password2":                 p2,
		"filename_encryption":       d.FileNameEnc,
		"directory_name_encryption": d.DirNameEnc,
		"filename_encoding":         "base64",
		"suffix":                    d.EncryptedSuffix,
		"pass_bad_blocks":           "",
	}
	c, err := rcCrypt.NewCipher(config)
	if err != nil {
		return fmt.Errorf("failed to create Cipher: %w", err)
	}
	d.cipher = c

	//c, err := rcCrypt.newCipher(rcCrypt.NameEncryptionStandard, "", "", true, nil)
	return nil
}

func (d *Crypt) updateObfusParm(str *string) error {
	temp := *str
	if !strings.HasPrefix(temp, obfuscatedPrefix) {
		temp, err := obscure.Obscure(temp)
		if err != nil {
			return err
		}
		temp = obfuscatedPrefix + temp
		*str = temp
	}
	return nil
}

func (d *Crypt) Drop(ctx context.Context) error {
	return nil
}

func (d *Crypt) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	path := dir.GetPath()
	//return d.list(ctx, d.RemotePath, path)
	//remoteFull

	objs, err := fs.List(ctx, d.getPathForRemote(path, true), &fs.ListArgs{NoLog: true})
	// the obj must implement the model.SetPath interface
	// return objs, err
	if err != nil {
		return nil, err
	}

	var result []model.Obj
	for _, obj := range objs {
		if obj.IsDir() {
			name, err := d.cipher.DecryptDirName(obj.GetName())
			if err != nil {
				//filter illegal files
				continue
			}
			objRes := model.Object{
				Name:     name,
				Size:     0,
				Modified: obj.ModTime(),
				IsFolder: obj.IsDir(),
			}
			result = append(result, &objRes)
		} else {
			thumb, ok := model.GetThumb(obj)
			size, err := d.cipher.DecryptedSize(obj.GetSize())
			if err != nil {
				//filter illegal files
				continue
			}
			name, err := d.cipher.DecryptFileName(obj.GetName())
			if err != nil {
				//filter illegal files
				continue
			}
			objRes := model.Object{
				Name:     name,
				Size:     size,
				Modified: obj.ModTime(),
				IsFolder: obj.IsDir(),
			}
			if !ok {
				result = append(result, &objRes)
			} else {
				objWithThumb := model.ObjThumb{
					Object: objRes,
					Thumbnail: model.Thumbnail{
						Thumbnail: thumb,
					},
				}
				result = append(result, &objWithThumb)
			}

		}
	}

	return result, nil
}

func (d *Crypt) getPathForRemote(path string, isFolder bool) (remoteFullPath string) {
	if isFolder && !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	dir, fileName := filepath.Split(path)

	remoteDir := d.cipher.EncryptDirName(dir)
	remoteFileName := ""
	if len(strings.TrimSpace(fileName)) > 0 {
		remoteFileName = d.cipher.EncryptFileName(fileName)
	}
	return stdpath.Join(d.RemotePath, remoteDir, remoteFileName)

}

// actual path is used for internal only. any link for user should come from remoteFullPath
func (d *Crypt) getActualPathForRemote(path string, isFolder bool) (string, error) {
	_, remoteActualPath, err := op.GetStorageAndActualPath(d.getPathForRemote(path, isFolder))
	return remoteActualPath, err
}

func (d *Crypt) Get(ctx context.Context, path string) (model.Obj, error) {
	if utils.PathEqual(path, "/") {
		return &model.Object{
			Name:     "Root",
			IsFolder: true,
			Path:     "/",
		}, nil
	}
	remoteFullPath := ""
	var remoteObj model.Obj
	var err, err2 error
	if !strings.HasSuffix(path, d.EncryptedSuffix) {
		//most likely it's a folder
		remoteFullPath = d.getPathForRemote(path, true)
		remoteObj, err = fs.Get(ctx, remoteFullPath, &fs.GetArgs{NoLog: true})
		if err != nil {
			if errs.IsObjectNotFound(err) {
				//treat path as a file and try again
				remoteFullPath = d.getPathForRemote(path, false)
				remoteObj, err2 = fs.Get(ctx, remoteFullPath, &fs.GetArgs{NoLog: true})
				if err2 != nil {
					return nil, err2
				}
			} else {
				return nil, err
			}
		}

	} else {
		//most likely it's a file
		remoteFullPath = d.getPathForRemote(path, false)
		remoteObj, err = fs.Get(ctx, remoteFullPath, &fs.GetArgs{NoLog: true})
		if err != nil {
			if errs.IsObjectNotFound(err) {
				//treat path as a folder and try again
				remoteFullPath = d.getPathForRemote(path, true)
				remoteObj, err2 = fs.Get(ctx, remoteFullPath, &fs.GetArgs{NoLog: true})
				if err2 != nil {
					return nil, err2
				}
			} else {
				return nil, err
			}
		}
	}
	var remoteSize int64 = 0
	name := ""
	if !remoteObj.IsDir() {
		remoteSize, err = d.cipher.DecryptedSize(remoteObj.GetSize())
		if err != nil {
			log.Warnf("DecryptedSize failed for %s ,will use original size, err:%s", path, err)
			remoteSize = remoteObj.GetSize()
		}
		name, err = d.cipher.DecryptFileName(remoteObj.GetName())
		if err != nil {
			log.Warnf("DecryptFileName failed for %s ,will use original name, err:%s", path, err)
			name = remoteObj.GetName()
		}
	} else {
		name, err = d.cipher.DecryptDirName(remoteObj.GetName())
		if err != nil {
			log.Warnf("DecryptDirName failed for %s ,will use original name, err:%s", path, err)
			name = remoteObj.GetName()
		}
	}
	obj := &model.Object{
		Path:     path,
		Name:     name,
		Size:     remoteSize,
		Modified: remoteObj.ModTime(),
		IsFolder: remoteObj.IsDir(),
	}
	return obj, nil
	//return nil, errs.ObjectNotFound
}

func (d *Crypt) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	dstDirActualPath, err := d.getActualPathForRemote(file.GetPath(), false)
	if err != nil {
		return nil, fmt.Errorf("failed to convert path to remote path: %w", err)
	}
	remoteLink, remoteFile, err := op.Link(ctx, d.remoteStorage, dstDirActualPath, args)
	if err != nil {
		return nil, err
	}

	/*_, err = fs.Get(ctx, dstDirActualPath, &fs.GetArgs{NoLog: true})
	if err != nil {
		return nil, err
	}
	if common.ShouldProxy(d.remoteStorage, stdpath.Base(file.GetPath())) {
		return &model.Link{
			URL: fmt.Sprintf("%s/p%s?sign=%s",
				common.GetApiUrl(args.HttpReq),
				utils.EncodePath(dstDirActualPath, true),
				sign.Sign(dstDirActualPath)),
		}, nil
	}
	link, _, err := fs.Link(ctx, dstDirActualPath, args)*/

	/*	link, err := d.link(ctx, d.RemotePath, file.GetPath(), args)
		if err != nil {
			return nil, err
		}*/

	if remoteLink.RangeReader == nil && remoteLink.ReadSeeker == nil && len(remoteLink.URL) == 0 && remoteLink.Data == nil {
		return nil, fmt.Errorf("the remote storage driver need to be enhanced to support encrytion")
	}
	remoteFileSize := remoteFile.GetSize()
	rangeReaderFunc := func(ctx context.Context, underlyingOffset, underlyingLength int64) (io.ReadCloser, error) {
		length := underlyingLength
		if underlyingLength >= 0 && underlyingOffset+underlyingLength > remoteFileSize {
			length = -1
		}
		if remoteLink.RangeReader != nil {
			//remoteRangeReader, err :=
			remoteReader, err := remoteLink.RangeReader(http_range.Range{Start: underlyingOffset, Length: length})
			if err != nil {
				return nil, err
			}
			return io.NopCloser(remoteReader), nil
		}
		if remoteLink.ReadSeeker != nil {
			_, err := remoteLink.ReadSeeker.Seek(underlyingOffset, io.SeekStart)
			if err != nil {
				return nil, err
			}
			return io.NopCloser(remoteLink.ReadSeeker), nil
		}
		if len(remoteLink.URL) > 0 {
			rangedRemoteLink := &model.Link{
				URL:    remoteLink.URL,
				Header: remoteLink.Header,
			}
			response, err := net.RequestRangedHttp(args.HttpReq, rangedRemoteLink, underlyingOffset, underlyingLength)
			if err != nil {
				_ = response.Body.Close()
				return nil, fmt.Errorf("remote storage http request failure,status: %d err:%s", response.StatusCode, err)
			}
			if response.StatusCode == http.StatusPartialContent {
				// server supported partial content, return as-is.
				return response.Body, nil
			}

			return response.Body, nil
		}
		if remoteLink.Data != nil {
			readCloser, err := net.GetRangedHttpReader(remoteLink.Data, underlyingOffset, underlyingLength)
			if err != nil {
				remoteLink.Data.Close()
				return nil, err
			}
			return readCloser, nil
		}
		return nil, errs.NotSupport

	}
	resultRangeReader := func(httpRange http_range.Range) (io.ReadCloser, error) {
		readSeeker, err := d.cipher.DecryptDataSeek(ctx, rangeReaderFunc, httpRange.Start, httpRange.Length)
		if err != nil {
			return nil, err
		}
		return readSeeker, nil
	}
	resultLink := &model.Link{
		Header:      remoteLink.Header,
		RangeReader: resultRangeReader,
		Expiration:  remoteLink.Expiration,
	}

	return resultLink, nil

}

func (d *Crypt) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	dstDirActualPath, err := d.getActualPathForRemote(parentDir.GetPath(), true)
	if err != nil {
		return fmt.Errorf("failed to convert path to remote path: %w", err)
	}
	dir := d.cipher.EncryptDirName(dirName)
	return op.MakeDir(ctx, d.remoteStorage, stdpath.Join(dstDirActualPath, dir))
}

func (d *Crypt) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO move obj, optional
	return errs.NotImplement
}

func (d *Crypt) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	dstDirActualPath, err := d.getActualPathForRemote(srcObj.GetPath(), srcObj.IsDir())
	if err != nil {
		return fmt.Errorf("failed to convert path to remote path: %w", err)
	}
	var newEncryptedName string
	if srcObj.IsDir() {
		newEncryptedName = d.cipher.EncryptDirName(newName)
	} else {
		newEncryptedName = d.cipher.EncryptFileName(newName)
	}
	return op.Rename(ctx, d.remoteStorage, dstDirActualPath, newEncryptedName)
}

func (d *Crypt) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	// TODO copy obj, optional
	return errs.NotImplement
}

func (d *Crypt) Remove(ctx context.Context, obj model.Obj) error {
	dstDirActualPath, err := d.getActualPathForRemote(obj.GetPath(), obj.IsDir())
	if err != nil {
		return fmt.Errorf("failed to convert path to remote path: %w", err)
	}
	return op.Remove(ctx, d.remoteStorage, dstDirActualPath)
}

func (d *Crypt) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	dstDirActualPath, err := d.getActualPathForRemote(dstDir.GetPath(), false)
	if err != nil {
		return fmt.Errorf("failed to convert path to remote path: %w", err)
	}

	in := stream.GetReadCloser()
	// Encrypt the data into wrappedIn
	wrappedIn, err := d.cipher.EncryptData(in)
	if err != nil {
		return fmt.Errorf("failed to EncryptData: %w", err)
	}

	streamOut := &model.FileStream{
		Obj: &model.Object{
			ID:       stream.GetID(),
			Path:     stream.GetPath(),
			Name:     d.cipher.EncryptFileName(stream.GetName()),
			Size:     d.cipher.EncryptedSize(stream.GetSize()),
			Modified: stream.ModTime(),
			IsFolder: stream.IsDir(),
		},
		ReadCloser:   io.NopCloser(wrappedIn),
		Mimetype:     "application/octet-stream",
		WebPutAsTask: stream.NeedStore(),
		Old:          stream.GetOld(),
	}
	err = op.Put(ctx, d.remoteStorage, dstDirActualPath, streamOut, up)
	if err != nil {
		return err
	}
	return nil
}

//func (d *Crypt) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*Crypt)(nil)

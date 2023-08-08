package aliyundrive_open

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/alist-org/alist/v3/drivers/base"
	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/go-resty/resty/v2"
	log "github.com/sirupsen/logrus"
)

func makePartInfos(size int) []base.Json {
	partInfoList := make([]base.Json, size)
	for i := 0; i < size; i++ {
		partInfoList[i] = base.Json{"part_number": 1 + i}
	}
	return partInfoList
}

func calPartSize(fileSize int64) int64 {
	var partSize int64 = 20 * 1024 * 1024
	if fileSize > partSize {
		if fileSize > 1*1024*1024*1024*1024 { // file Size over 1TB
			partSize = 5 * 1024 * 1024 * 1024 // file part size 5GB
		} else if fileSize > 768*1024*1024*1024 { // over 768GB
			partSize = 109951163 // ≈ 104.8576MB, split 1TB into 10,000 part
		} else if fileSize > 512*1024*1024*1024 { // over 512GB
			partSize = 82463373 // ≈ 78.6432MB
		} else if fileSize > 384*1024*1024*1024 { // over 384GB
			partSize = 54975582 // ≈ 52.4288MB
		} else if fileSize > 256*1024*1024*1024 { // over 256GB
			partSize = 41231687 // ≈ 39.3216MB
		} else if fileSize > 128*1024*1024*1024 { // over 128GB
			partSize = 27487791 // ≈ 26.2144MB
		}
	}
	return partSize
}

func (d *AliyundriveOpen) getUploadUrl(count int, fileId, uploadId string) ([]PartInfo, error) {
	partInfoList := makePartInfos(count)
	var resp CreateResp
	_, err := d.request("/adrive/v1.0/openFile/getUploadUrl", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id":       d.DriveId,
			"file_id":        fileId,
			"part_info_list": partInfoList,
			"upload_id":      uploadId,
		}).SetResult(&resp)
	})
	return resp.PartInfoList, err
}

func (d *AliyundriveOpen) uploadPart(ctx context.Context, i, count int, reader *utils.MultiReadable, resp *CreateResp, retry bool) error {
	partInfo := resp.PartInfoList[i-1]
	uploadUrl := partInfo.UploadUrl
	if d.InternalUpload {
		uploadUrl = strings.ReplaceAll(uploadUrl, "https://cn-beijing-data.aliyundrive.net/", "http://ccp-bj29-bj-1592982087.oss-cn-beijing-internal.aliyuncs.com/")
	}
	req, err := http.NewRequest("PUT", uploadUrl, reader)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	res, err := base.HttpClient.Do(req)
	if err != nil {
		if retry {
			reader.Reset()
			return d.uploadPart(ctx, i, count, reader, resp, false)
		}
		return err
	}
	res.Body.Close()
	if retry && res.StatusCode == http.StatusForbidden {
		resp.PartInfoList, err = d.getUploadUrl(count, resp.FileId, resp.UploadId)
		if err != nil {
			return err
		}
		reader.Reset()
		return d.uploadPart(ctx, i, count, reader, resp, false)
	}
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusConflict {
		return fmt.Errorf("upload status: %d", res.StatusCode)
	}
	return nil
}

func (d *AliyundriveOpen) normalUpload(ctx context.Context, stream model.FileStreamer, up driver.UpdateProgress, createResp CreateResp, count int, partSize int64) error {
	log.Debugf("[aliyundive_open] normal upload")
	// 2. upload
	preTime := time.Now()
	for i := 1; i <= len(createResp.PartInfoList); i++ {
		if utils.IsCanceled(ctx) {
			return ctx.Err()
		}
		err := d.uploadPart(ctx, i, count, utils.NewMultiReadable(io.LimitReader(stream, partSize)), &createResp, true)
		if err != nil {
			return err
		}
		if count > 0 {
			up(i * 100 / count)
		}
		// refresh upload url if 50 minutes passed
		if time.Since(preTime) > 50*time.Minute {
			createResp.PartInfoList, err = d.getUploadUrl(count, createResp.FileId, createResp.UploadId)
			if err != nil {
				return err
			}
			preTime = time.Now()
		}
	}
	// 3. complete
	_, err := d.request("/adrive/v1.0/openFile/complete", http.MethodPost, func(req *resty.Request) {
		req.SetBody(base.Json{
			"drive_id":  d.DriveId,
			"file_id":   createResp.FileId,
			"upload_id": createResp.UploadId,
		})
	})
	return err
}

type ProofRange struct {
	Start int64
	End   int64
}

func getProofRange(input string, size int64) (*ProofRange, error) {
	if size == 0 {
		return &ProofRange{}, nil
	}
	tmpStr := utils.GetMD5EncodeStr(input)[0:16]
	tmpInt, err := strconv.ParseUint(tmpStr, 16, 64)
	if err != nil {
		return nil, err
	}
	index := tmpInt % uint64(size)
	pr := &ProofRange{
		Start: int64(index),
		End:   int64(index) + 8,
	}
	if pr.End >= size {
		pr.End = size
	}
	return pr, nil
}

// io.ReadSeekCloser
func (d *AliyundriveOpen) calProofCode(readSeekCloser io.ReadSeekCloser, fileSize int64) (string, error) {
	proofRange, err := getProofRange(d.AccessToken, fileSize)
	if err != nil {
		return "", err
	}
	length := proofRange.End - proofRange.Start
	buf := bytes.NewBuffer(make([]byte, length))
	_, err = readSeekCloser.Seek(proofRange.Start, io.SeekStart)
	io.CopyN(buf, readSeekCloser, length)

	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func (d *AliyundriveOpen) upload(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	// 1. create
	// Part Size Unit: Bytes, Default: 20MB,
	// Maximum number of slices 10,000, ≈195.3125GB
	var partSize = calPartSize(stream.GetSize())
	const dateFormat = "2019-08-20T06:51:27.292Z"
	mtime := stream.ModTime()
	mtimeStr := mtime.UTC().Format(dateFormat)
	ctimeStr := mtimeStr
	ct, ok := stream.GetObj().(*model.Object)
	if ok {
		ctimeStr = ct.CreateTime().UTC().Format(dateFormat)
	}

	createData := base.Json{
		"drive_id":          d.DriveId,
		"parent_file_id":    dstDir.GetID(),
		"name":              stream.GetName(),
		"type":              "file",
		"check_name_mode":   "ignore",
		"local_modified_at": mtimeStr,
		"local_created_at":  ctimeStr,
	}
	count := int(math.Ceil(float64(stream.GetSize()) / float64(partSize)))
	createData["part_info_list"] = makePartInfos(count)
	// rapid upload
	rapidUpload := stream.GetSize() > 100*1024 && d.RapidUpload
	if rapidUpload {
		log.Debugf("[aliyundrive_open] start cal pre_hash")
		// read 1024 bytes to calculate pre hash
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		_, err := io.CopyN(buf, stream, 1024)
		if err != nil {
			return err
		}
		createData["size"] = stream.GetSize()
		createData["pre_hash"] = utils.GetSHA1Encode(buf.Bytes())
		// if support seek, seek to start
		if localFile, ok := stream.(io.Seeker); ok {
			if _, err := localFile.Seek(0, io.SeekStart); err != nil {
				return err
			}
		} else {
			// Put spliced head back to stream
			stream.SetReadCloser(struct {
				io.Reader
				io.Closer
			}{
				Reader: io.MultiReader(buf, stream.GetReadCloser()),
				Closer: stream.GetReadCloser(),
			})
		}
	}
	var createResp CreateResp
	_, err, e := d.requestReturnErrResp("/adrive/v1.0/openFile/create", http.MethodPost, func(req *resty.Request) {
		req.SetBody(createData).SetResult(&createResp)
	})
	if err != nil {
		if e.Code != "PreHashMatched" || !rapidUpload {
			return err
		}
		log.Debugf("[aliyundrive_open] pre_hash matched, start rapid upload")
		// convert to local file
		file, err := utils.NewBufferedReadSeekCloser(stream, stream.GetSize())
		if err != nil {
			return err
		}
		_ = stream.GetReadCloser().Close()
		stream.SetReadCloser(file)
		// calculate full hash
		h := sha1.New()
		_, err = io.Copy(h, file)
		if err != nil {
			return err
		}
		delete(createData, "pre_hash")
		createData["proof_version"] = "v1"
		createData["content_hash_name"] = "sha1"
		createData["content_hash"] = hex.EncodeToString(h.Sum(nil))
		// seek to start
		if _, err = file.Seek(0, io.SeekStart); err != nil {
			return err
		}
		createData["proof_code"], err = d.calProofCode(file, stream.GetSize())
		if err != nil {
			return fmt.Errorf("cal proof code error: %s", err.Error())
		}
		_, err = d.request("/adrive/v1.0/openFile/create", http.MethodPost, func(req *resty.Request) {
			req.SetBody(createData).SetResult(&createResp)
		})
		if err != nil {
			return err
		}
		if createResp.RapidUpload {
			log.Debugf("[aliyundrive_open] rapid upload success, file id: %s", createResp.FileId)
			return nil
		}
		// failed to rapid upload, try normal upload
		if _, err = file.Seek(0, io.SeekStart); err != nil {
			return err
		}
	}
	log.Debugf("[aliyundrive_open] create file success, resp: %+v", createResp)
	return d.normalUpload(ctx, stream, up, createResp, count, partSize)
}

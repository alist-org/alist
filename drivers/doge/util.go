package s3

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
)

// do others that not defined in Driver interface

func (d *Doge) initSession() error {
	var err error
	credentialsTmp, err := getCredentials(d.AccessKeyID, d.SecretAccessKey)
	if err != nil {
		return err
	}
	cfg := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(credentialsTmp.AccessKeyId, credentialsTmp.SecretAccessKey, credentialsTmp.SessionToken),
		Region:           &d.Region,
		Endpoint:         &d.Endpoint,
		S3ForcePathStyle: aws.Bool(d.ForcePathStyle),
	}
	d.Session, err = session.NewSession(cfg)
	return err
}

func (d *Doge) getClient(link bool) *s3.S3 {
	client := s3.New(d.Session)
	if link && d.CustomHost != "" {
		client.Handlers.Build.PushBack(func(r *request.Request) {
			if r.HTTPRequest.Method != http.MethodGet {
				return
			}
			//判断CustomHost是否以http://或https://开头
			split := strings.SplitN(d.CustomHost, "://", 2)
			if utils.SliceContains([]string{"http", "https"}, split[0]) {
				r.HTTPRequest.URL.Scheme = split[0]
				r.HTTPRequest.URL.Host = split[1]
			} else {
				r.HTTPRequest.URL.Host = d.CustomHost
			}
		})
	}
	return client
}

type TmpTokenResponse struct {
	Code int                  `json:"code"`
	Msg  string               `json:"msg"`
	Data TmpTokenResponseData `json:"data,omitempty"`
}
type TmpTokenResponseData struct {
	Credentials Credentials `json:"Credentials"`
}
type Credentials struct {
	AccessKeyId     string `json:"accessKeyId,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	SessionToken    string `json:"sessionToken,omitempty"`
}

func getCredentials(AccessKey, SecretKey string) (rst Credentials, err error) {
	apiPath := "/auth/tmp_token.json"
	reqBody, err := json.Marshal(map[string]interface{}{"channel": "OSS_FULL", "scopes": []string{"*"}})
	if err != nil {
		return rst, err
	}

	signStr := apiPath + "\n" + string(reqBody)
	hmacObj := hmac.New(sha1.New, []byte(SecretKey))
	hmacObj.Write([]byte(signStr))
	sign := hex.EncodeToString(hmacObj.Sum(nil))
	Authorization := "TOKEN " + AccessKey + ":" + sign

	req, err := http.NewRequest("POST", "https://api.dogecloud.com"+apiPath, strings.NewReader(string(reqBody)))
	if err != nil {
		return rst, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", Authorization)
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return rst, err
	}
	defer resp.Body.Close()
	ret, err := io.ReadAll(resp.Body)
	if err != nil {
		return rst, err
	}
	var tmpTokenResp TmpTokenResponse
	err = json.Unmarshal(ret, &tmpTokenResp)
	if err != nil {
		return rst, err
	}
	return tmpTokenResp.Data.Credentials, nil
}

func getKey(path string, dir bool) string {
	path = strings.TrimPrefix(path, "/")
	if path != "" && dir {
		path += "/"
	}
	return path
}

var defaultPlaceholderName = ".alist"

func getPlaceholderName(placeholder string) string {
	if placeholder == "" {
		return defaultPlaceholderName
	}
	return placeholder
}

func (d *Doge) listV1(prefix string, args model.ListArgs) ([]model.Obj, error) {
	prefix = getKey(prefix, true)
	log.Debugf("list: %s", prefix)
	files := make([]model.Obj, 0)
	marker := ""
	for {
		input := &s3.ListObjectsInput{
			Bucket:    &d.Bucket,
			Marker:    &marker,
			Prefix:    &prefix,
			Delimiter: aws.String("/"),
		}
		listObjectsResult, err := d.client.ListObjects(input)
		if err != nil {
			return nil, err
		}
		for _, object := range listObjectsResult.CommonPrefixes {
			name := path.Base(strings.Trim(*object.Prefix, "/"))
			file := model.Object{
				//Id:        *object.Key,
				Name:     name,
				Modified: d.Modified,
				IsFolder: true,
			}
			files = append(files, &file)
		}
		for _, object := range listObjectsResult.Contents {
			name := path.Base(*object.Key)
			if !args.S3ShowPlaceholder && (name == getPlaceholderName(d.Placeholder) || name == d.Placeholder) {
				continue
			}
			file := model.Object{
				//Id:        *object.Key,
				Name:     name,
				Size:     *object.Size,
				Modified: *object.LastModified,
			}
			files = append(files, &file)
		}
		if listObjectsResult.IsTruncated == nil {
			return nil, errors.New("IsTruncated nil")
		}
		if *listObjectsResult.IsTruncated {
			marker = *listObjectsResult.NextMarker
		} else {
			break
		}
	}
	return files, nil
}

func (d *Doge) listV2(prefix string, args model.ListArgs) ([]model.Obj, error) {
	prefix = getKey(prefix, true)
	files := make([]model.Obj, 0)
	var continuationToken, startAfter *string
	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            &d.Bucket,
			ContinuationToken: continuationToken,
			Prefix:            &prefix,
			Delimiter:         aws.String("/"),
			StartAfter:        startAfter,
		}
		listObjectsResult, err := d.client.ListObjectsV2(input)
		if err != nil {
			return nil, err
		}
		log.Debugf("resp: %+v", listObjectsResult)
		for _, object := range listObjectsResult.CommonPrefixes {
			name := path.Base(strings.Trim(*object.Prefix, "/"))
			file := model.Object{
				//Id:        *object.Key,
				Name:     name,
				Modified: d.Modified,
				IsFolder: true,
			}
			files = append(files, &file)
		}
		for _, object := range listObjectsResult.Contents {
			if strings.HasSuffix(*object.Key, "/") {
				continue
			}
			name := path.Base(*object.Key)
			if !args.S3ShowPlaceholder && (name == getPlaceholderName(d.Placeholder) || name == d.Placeholder) {
				continue
			}
			file := model.Object{
				//Id:        *object.Key,
				Name:     name,
				Size:     *object.Size,
				Modified: *object.LastModified,
			}
			files = append(files, &file)
		}
		if !aws.BoolValue(listObjectsResult.IsTruncated) {
			break
		}
		if listObjectsResult.NextContinuationToken != nil {
			continuationToken = listObjectsResult.NextContinuationToken
			continue
		}
		if len(listObjectsResult.Contents) == 0 {
			break
		}
		startAfter = listObjectsResult.Contents[len(listObjectsResult.Contents)-1].Key
	}
	return files, nil
}

func (d *Doge) copy(ctx context.Context, src string, dst string, isDir bool) error {
	if isDir {
		return d.copyDir(ctx, src, dst)
	}
	return d.copyFile(ctx, src, dst)
}

func (d *Doge) copyFile(ctx context.Context, src string, dst string) error {
	srcKey := getKey(src, false)
	dstKey := getKey(dst, false)
	input := &s3.CopyObjectInput{
		Bucket:     &d.Bucket,
		CopySource: aws.String("/" + d.Bucket + "/" + srcKey),
		Key:        &dstKey,
	}
	_, err := d.client.CopyObject(input)
	return err
}

func (d *Doge) copyDir(ctx context.Context, src string, dst string) error {
	objs, err := op.List(ctx, d, src, model.ListArgs{S3ShowPlaceholder: true})
	if err != nil {
		return err
	}
	for _, obj := range objs {
		cSrc := path.Join(src, obj.GetName())
		cDst := path.Join(dst, obj.GetName())
		if obj.IsDir() {
			err = d.copyDir(ctx, cSrc, cDst)
		} else {
			err = d.copyFile(ctx, cSrc, cDst)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Doge) removeDir(ctx context.Context, src string) error {
	objs, err := op.List(ctx, d, src, model.ListArgs{})
	if err != nil {
		return err
	}
	for _, obj := range objs {
		cSrc := path.Join(src, obj.GetName())
		if obj.IsDir() {
			err = d.removeDir(ctx, cSrc)
		} else {
			err = d.removeFile(cSrc)
		}
		if err != nil {
			return err
		}
	}
	_ = d.removeFile(path.Join(src, getPlaceholderName(d.Placeholder)))
	_ = d.removeFile(path.Join(src, d.Placeholder))
	return nil
}

func (d *Doge) removeFile(src string) error {
	key := getKey(src, false)
	input := &s3.DeleteObjectInput{
		Bucket: &d.Bucket,
		Key:    &key,
	}
	_, err := d.client.DeleteObject(input)
	return err
}

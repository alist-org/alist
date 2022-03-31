package s3

import (
	"errors"
	"fmt"
	"github.com/Xhofe/alist/conf"
	"github.com/Xhofe/alist/drivers/base"
	"github.com/Xhofe/alist/model"
	"github.com/Xhofe/alist/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"path"
	"strings"
)

var sessionsMap map[string]*session.Session

func (driver S3) NewSession(account *model.Account) (*session.Session, error) {
	cfg := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(account.AccessKey, account.AccessSecret, ""),
		Region:           &account.Region,
		Endpoint:         &account.Endpoint,
		S3ForcePathStyle: aws.Bool(account.Bool1),
	}
	return session.NewSession(cfg)
}

func (driver S3) GetClient(account *model.Account, link bool) (*s3.S3, error) {
	s, ok := sessionsMap[account.Name]
	if !ok {
		return nil, fmt.Errorf("can't find [%s] session", account.Name)
	}
	client := s3.New(s)
	if link && account.CustomHost != "" {
		cURL, err := url.Parse(account.CustomHost)
		if err != nil {
			return nil, err
		}
		client.Handlers.Build.PushBack(func(r *request.Request) {
			if r.HTTPRequest.Method != http.MethodGet {
				return
			}
			r.HTTPRequest.URL.Scheme = cURL.Scheme
			r.HTTPRequest.URL.Host = cURL.Host
		})
	}
	return client, nil
}

func (driver S3) List(prefix string, account *model.Account) ([]model.File, error) {
	prefix = driver.GetKey(prefix, account, true)
	log.Debugf("list: %s", prefix)
	client, err := driver.GetClient(account, false)
	if err != nil {
		return nil, err
	}
	files := make([]model.File, 0)
	marker := ""
	for {
		input := &s3.ListObjectsInput{
			Bucket:    &account.Bucket,
			Marker:    &marker,
			Prefix:    &prefix,
			Delimiter: aws.String("/"),
		}
		listObjectsResult, err := client.ListObjects(input)

		if err != nil {
			return nil, err
		}
		for _, object := range listObjectsResult.CommonPrefixes {
			name := utils.Base(strings.Trim(*object.Prefix, "/"))
			file := model.File{
				//Id:        *object.Key,
				Name:      name,
				Driver:    driver.Config().Name,
				UpdatedAt: account.UpdatedAt,
				TimeStr:   "-",
				Type:      conf.FOLDER,
			}
			files = append(files, file)
		}
		for _, object := range listObjectsResult.Contents {
			name := utils.Base(*object.Key)
			if name == getPlaceholderName(account.Zone) {
				continue
			}
			file := model.File{
				//Id:        *object.Key,
				Name:      name,
				Size:      *object.Size,
				Driver:    driver.Config().Name,
				UpdatedAt: object.LastModified,
				Type:      utils.GetFileType(path.Ext(*object.Key)),
			}
			files = append(files, file)
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

func (driver S3) ListV2(prefix string, account *model.Account) ([]model.File, error) {
	prefix = driver.GetKey(prefix, account, true)
	//if prefix == "" {
	//	prefix = "/"
	//}
	log.Debugf("list: %s", prefix)
	client, err := driver.GetClient(account, false)
	if err != nil {
		return nil, err
	}
	files := make([]model.File, 0)
	var continuationToken, startAfter *string
	for {
		input := &s3.ListObjectsV2Input{
			Bucket:            &account.Bucket,
			ContinuationToken: continuationToken,
			Prefix:            &prefix,
			Delimiter:         aws.String("/"),
			StartAfter:        startAfter,
		}
		listObjectsResult, err := client.ListObjectsV2(input)
		if err != nil {
			return nil, err
		}
		log.Debugf("resp: %+v", listObjectsResult)
		for _, object := range listObjectsResult.CommonPrefixes {
			name := utils.Base(strings.Trim(*object.Prefix, "/"))
			file := model.File{
				//Id:        *object.Key,
				Name:      name,
				Driver:    driver.Config().Name,
				UpdatedAt: account.UpdatedAt,
				TimeStr:   "-",
				Type:      conf.FOLDER,
			}
			files = append(files, file)
		}
		for _, object := range listObjectsResult.Contents {
			name := utils.Base(*object.Key)
			if name == getPlaceholderName(account.Zone) {
				continue
			}
			file := model.File{
				//Id:        *object.Key,
				Name:      name,
				Size:      *object.Size,
				Driver:    driver.Config().Name,
				UpdatedAt: object.LastModified,
				Type:      utils.GetFileType(path.Ext(*object.Key)),
			}
			files = append(files, file)
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

func (driver S3) GetKey(path string, account *model.Account, dir bool) string {
	path = utils.Join(account.RootFolder, path)
	path = strings.TrimPrefix(path, "/")
	if path != "" && dir {
		path += "/"
	}
	return path
}

func init() {
	sessionsMap = make(map[string]*session.Session)
	base.RegisterDriver(&S3{})
}

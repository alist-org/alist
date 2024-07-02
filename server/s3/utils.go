// Credits: https://pkg.go.dev/github.com/rclone/rclone@v1.65.2/cmd/serve/s3
// Package s3 implements a fake s3 server for alist
package s3

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/alist-org/alist/v3/internal/conf"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/fs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/internal/op"
	"github.com/alist-org/alist/v3/internal/setting"
	"github.com/alist-org/gofakes3"
)

type Bucket struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func getAndParseBuckets() ([]Bucket, error) {
	var res []Bucket
	err := json.Unmarshal([]byte(setting.GetStr(conf.S3Buckets)), &res)
	return res, err
}

func getBucketByName(name string) (Bucket, error) {
	buckets, err := getAndParseBuckets()
	if err != nil {
		return Bucket{}, err
	}
	for _, b := range buckets {
		if b.Name == name {
			return b, nil
		}
	}
	return Bucket{}, gofakes3.BucketNotFound(name)
}

func getDirEntries(path string) ([]model.Obj, error) {
	ctx := context.Background()
	meta, _ := op.GetNearestMeta(path)
	fi, err := fs.Get(context.WithValue(ctx, "meta", meta), path, &fs.GetArgs{})
	if errs.IsNotFoundError(err) {
		return nil, gofakes3.ErrNoSuchKey
	} else if err != nil {
		return nil, gofakes3.ErrNoSuchKey
	}

	if !fi.IsDir() {
		return nil, gofakes3.ErrNoSuchKey
	}

	dirEntries, err := fs.List(context.WithValue(ctx, "meta", meta), path, &fs.ListArgs{})
	if err != nil {
		return nil, err
	}

	return dirEntries, nil
}

// func getFileHashByte(node interface{}) []byte {
// 	b, err := hex.DecodeString(getFileHash(node))
// 	if err != nil {
// 		return nil
// 	}
// 	return b
// }

func getFileHash(node interface{}) string {
	// var o fs.Object

	// switch b := node.(type) {
	// case vfs.Node:
	// 	fsObj, ok := b.DirEntry().(fs.Object)
	// 	if !ok {
	// 		fs.Debugf("serve s3", "File uploading - reading hash from VFS cache")
	// 		in, err := b.Open(os.O_RDONLY)
	// 		if err != nil {
	// 			return ""
	// 		}
	// 		defer func() {
	// 			_ = in.Close()
	// 		}()
	// 		h, err := hash.NewMultiHasherTypes(hash.NewHashSet(hash.MD5))
	// 		if err != nil {
	// 			return ""
	// 		}
	// 		_, err = io.Copy(h, in)
	// 		if err != nil {
	// 			return ""
	// 		}
	// 		return h.Sums()[hash.MD5]
	// 	}
	// 	o = fsObj
	// case fs.Object:
	// 	o = b
	// }

	// hash, err := o.Hash(context.Background(), hash.MD5)
	// if err != nil {
	// 	return ""
	// }
	// return hash
	return ""
}

func prefixParser(p *gofakes3.Prefix) (path, remaining string) {
	idx := strings.LastIndexByte(p.Prefix, '/')
	if idx < 0 {
		return "", p.Prefix
	}
	return p.Prefix[:idx], p.Prefix[idx+1:]
}

// // FIXME this could be implemented by VFS.MkdirAll()
// func mkdirRecursive(path string, VFS *vfs.VFS) error {
// 	path = strings.Trim(path, "/")
// 	dirs := strings.Split(path, "/")
// 	dir := ""
// 	for _, d := range dirs {
// 		dir += "/" + d
// 		if _, err := VFS.Stat(dir); err != nil {
// 			err := VFS.Mkdir(dir, 0777)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}
// 	return nil
// }

// func rmdirRecursive(p string, VFS *vfs.VFS) {
// 	dir := path.Dir(p)
// 	if !strings.ContainsAny(dir, "/\\") {
// 		// might be bucket(root)
// 		return
// 	}
// 	if _, err := VFS.Stat(dir); err == nil {
// 		err := VFS.Remove(dir)
// 		if err != nil {
// 			return
// 		}
// 		rmdirRecursive(dir, VFS)
// 	}
// }

func authlistResolver() map[string]string {
	s3accesskeyid := setting.GetStr(conf.S3AccessKeyId)
	s3secretaccesskey := setting.GetStr(conf.S3SecretAccessKey)
	if s3accesskeyid == "" && s3secretaccesskey == "" {
		return nil
	}
	authList := make(map[string]string)
	authList[s3accesskeyid] = s3secretaccesskey
	return authList
}

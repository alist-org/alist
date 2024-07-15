package halalcloud

import (
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	"github.com/city404/v6-public-rpc-proto/go/v6/common"
	pubUserFile "github.com/city404/v6-public-rpc-proto/go/v6/userfile"
	"google.golang.org/grpc"
	"time"
)

type AuthService struct {
	appID          string
	appVersion     string
	appSecret      string
	grpcConnection *grpc.ClientConn
	dopts          halalOptions
	tr             *TokenResp
}

type TokenResp struct {
	AccessToken           string `json:"accessToken,omitempty"`
	AccessTokenExpiredAt  int64  `json:"accessTokenExpiredAt,omitempty"`
	RefreshToken          string `json:"refreshToken,omitempty"`
	RefreshTokenExpiredAt int64  `json:"refreshTokenExpiredAt,omitempty"`
}

type UserInfo struct {
	Identity string `json:"identity,omitempty"`
	UpdateTs int64  `json:"updateTs,omitempty"`
	Name     string `json:"name,omitempty"`
	CreateTs int64  `json:"createTs,omitempty"`
}

type OrderByInfo struct {
	Field string `json:"field,omitempty"`
	Asc   bool   `json:"asc,omitempty"`
}

type ListInfo struct {
	Token   string         `json:"token,omitempty"`
	Limit   int64          `json:"limit,omitempty"`
	OrderBy []*OrderByInfo `json:"order_by,omitempty"`
	Version int32          `json:"version,omitempty"`
}

type FilesList struct {
	Files    []*Files                `json:"files,omitempty"`
	ListInfo *common.ScanListRequest `json:"list_info,omitempty"`
}

var _ model.Obj = (*Files)(nil)

type Files pubUserFile.File

func (f *Files) GetSize() int64 {
	return f.Size
}

func (f *Files) GetName() string {
	return f.Name
}

func (f *Files) ModTime() time.Time {
	return time.UnixMilli(f.UpdateTs)
}

func (f *Files) CreateTime() time.Time {
	return time.UnixMilli(f.UpdateTs)
}

func (f *Files) IsDir() bool {
	return f.Dir
}

func (f *Files) GetHash() utils.HashInfo {
	return utils.HashInfo{}
}

func (f *Files) GetID() string {
	if len(f.Identity) == 0 {
		f.Identity = "/"
	}
	return f.Identity
}

func (f *Files) GetPath() string {
	return f.Path
}

type SteamFile struct {
	file model.File
}

func (s *SteamFile) Read(p []byte) (n int, err error) {
	return s.file.Read(p)
}

func (s *SteamFile) Close() error {
	return s.file.Close()
}

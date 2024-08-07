package halalcloud

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/utils"
	pbPublicUser "github.com/city404/v6-public-rpc-proto/go/v6/user"
	pubUserFile "github.com/city404/v6-public-rpc-proto/go/v6/userfile"
	"github.com/google/uuid"
	"github.com/ipfs/go-cid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"hash"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	AppID      = "alist/10001"
	AppVersion = "1.0.0"
	AppSecret  = "bR4SJwOkvnG5WvVJ"
)

const (
	grpcServer     = "grpcuserapi.2dland.cn:443"
	grpcServerAuth = "grpcuserapi.2dland.cn"
)

func (d *HalalCloud) NewAuthServiceWithOauth(options ...HalalOption) (*AuthService, error) {

	aService := &AuthService{}
	err2 := errors.New("")

	svc := d.HalalCommon.AuthService
	for _, opt := range options {
		opt.apply(&svc.dopts)
	}

	grpcOptions := svc.dopts.grpcOptions
	grpcOptions = append(grpcOptions, grpc.WithAuthority(grpcServerAuth), grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})), grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctxx := svc.signContext(method, ctx)
		err := invoker(ctxx, method, req, reply, cc, opts...) // invoking RPC method
		return err
	}))

	grpcConnection, err := grpc.NewClient(grpcServer, grpcOptions...)
	if err != nil {
		return nil, err
	}
	defer grpcConnection.Close()
	userClient := pbPublicUser.NewPubUserClient(grpcConnection)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	stateString := uuid.New().String()
	// queryValues.Add("callback", oauthToken.Callback)
	oauthToken, err := userClient.CreateAuthToken(ctx, &pbPublicUser.LoginRequest{
		ReturnType: 2,
		State:      stateString,
		ReturnUrl:  "",
	})
	if err != nil {
		return nil, err
	}
	if len(oauthToken.State) < 1 {
		oauthToken.State = stateString
	}

	if oauthToken.Url != "" {

		return nil, fmt.Errorf(`need verify: <a target="_blank" href="%s">Click Here</a>`, oauthToken.Url)
	}

	return aService, err2

}

func (d *HalalCloud) NewAuthService(refreshToken string, options ...HalalOption) (*AuthService, error) {
	svc := d.HalalCommon.AuthService

	if len(refreshToken) < 1 {
		refreshToken = d.Addition.RefreshToken
	}

	if len(d.tr.AccessToken) > 0 {
		accessTokenExpiredAt := d.tr.AccessTokenExpiredAt
		current := time.Now().UnixMilli()
		if accessTokenExpiredAt < current {
			// access token expired
			d.tr.AccessToken = ""
			d.tr.AccessTokenExpiredAt = 0
		} else {
			svc.tr.AccessTokenExpiredAt = accessTokenExpiredAt
			svc.tr.AccessToken = d.tr.AccessToken
		}
	}

	for _, opt := range options {
		opt.apply(&svc.dopts)
	}

	grpcOptions := svc.dopts.grpcOptions
	grpcOptions = append(grpcOptions, grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(10*1024*1024), grpc.MaxCallRecvMsgSize(10*1024*1024)), grpc.WithAuthority(grpcServerAuth), grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})), grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctxx := svc.signContext(method, ctx)
		err := invoker(ctxx, method, req, reply, cc, opts...) // invoking RPC method
		if err != nil {
			grpcStatus, ok := status.FromError(err)

			if ok && grpcStatus.Code() == codes.Unauthenticated && strings.Contains(grpcStatus.Err().Error(), "invalid accesstoken") && len(refreshToken) > 0 {
				// refresh token
				refreshResponse, err := pbPublicUser.NewPubUserClient(cc).Refresh(ctx, &pbPublicUser.Token{
					RefreshToken: refreshToken,
				})
				if err != nil {
					return err
				}
				if len(refreshResponse.AccessToken) > 0 {
					svc.tr.AccessToken = refreshResponse.AccessToken
					svc.tr.AccessTokenExpiredAt = refreshResponse.AccessTokenExpireTs
					svc.OnAccessTokenRefreshed(refreshResponse.AccessToken, refreshResponse.AccessTokenExpireTs, refreshResponse.RefreshToken, refreshResponse.RefreshTokenExpireTs)
				}
				// retry
				ctxx := svc.signContext(method, ctx)
				err = invoker(ctxx, method, req, reply, cc, opts...) // invoking RPC method
				if err != nil {
					return err
				} else {
					return nil
				}
			}
		}
		return err
	}))
	grpcConnection, err := grpc.NewClient(grpcServer, grpcOptions...)

	if err != nil {
		return nil, err
	}

	svc.grpcConnection = grpcConnection
	return svc, err
}

func (s *AuthService) OnAccessTokenRefreshed(accessToken string, accessTokenExpiredAt int64, refreshToken string, refreshTokenExpiredAt int64) {
	s.tr.AccessToken = accessToken
	s.tr.AccessTokenExpiredAt = accessTokenExpiredAt
	s.tr.RefreshToken = refreshToken
	s.tr.RefreshTokenExpiredAt = refreshTokenExpiredAt

	if s.dopts.onTokenRefreshed != nil {
		s.dopts.onTokenRefreshed(accessToken, accessTokenExpiredAt, refreshToken, refreshTokenExpiredAt)
	}

}

func (s *AuthService) GetGrpcConnection() *grpc.ClientConn {
	return s.grpcConnection
}

func (s *AuthService) Close() {
	_ = s.grpcConnection.Close()
}

func (s *AuthService) signContext(method string, ctx context.Context) context.Context {
	var kvString []string
	currentTimeStamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	bufferedString := bytes.NewBufferString(method)
	kvString = append(kvString, "timestamp", currentTimeStamp)
	bufferedString.WriteString(currentTimeStamp)
	kvString = append(kvString, "appid", s.appID)
	bufferedString.WriteString(s.appID)
	kvString = append(kvString, "appversion", s.appVersion)
	bufferedString.WriteString(s.appVersion)
	if s.tr != nil && len(s.tr.AccessToken) > 0 {
		authorization := "Bearer " + s.tr.AccessToken
		kvString = append(kvString, "authorization", authorization)
		bufferedString.WriteString(authorization)
	}
	bufferedString.WriteString(s.appSecret)
	sign := GetMD5Hash(bufferedString.String())
	kvString = append(kvString, "sign", sign)
	return metadata.AppendToOutgoingContext(ctx, kvString...)
}

func (d *HalalCloud) GetCurrentOpDir(dir model.Obj, args []string, index int) string {
	currentDir := dir.GetPath()
	if len(currentDir) == 0 {
		currentDir = "/"
	}
	opPath := currentDir + "/" + args[index]
	if strings.HasPrefix(args[index], "/") {
		opPath = args[index]
	}
	return opPath
}

func (d *HalalCloud) GetCurrentDir(dir model.Obj) string {
	currentDir := dir.GetPath()
	if len(currentDir) == 0 {
		currentDir = "/"
	}
	return currentDir
}

type Common struct {
}

func getRawFiles(addr *pubUserFile.SliceDownloadInfo) ([]byte, error) {

	if addr == nil {
		return nil, errors.New("addr is nil")
	}

	client := http.Client{
		Timeout: time.Duration(60 * time.Second), // Set timeout to 5 seconds
	}
	resp, err := client.Get(addr.DownloadAddress)
	if err != nil {

		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s, body: %s", resp.Status, body)
	}

	if addr.Encrypt > 0 {
		cd := uint8(addr.Encrypt)
		for idx := 0; idx < len(body); idx++ {
			body[idx] = body[idx] ^ cd
		}
	}

	if addr.StoreType != 10 {

		sourceCid, err := cid.Decode(addr.Identity)
		if err != nil {
			return nil, err
		}
		checkCid, err := sourceCid.Prefix().Sum(body)
		if err != nil {
			return nil, err
		}
		if !checkCid.Equals(sourceCid) {
			return nil, fmt.Errorf("bad cid: %s, body: %s", checkCid.String(), body)
		}
	}

	return body, nil

}

type openObject struct {
	ctx     context.Context
	mu      sync.Mutex
	d       []*pubUserFile.SliceDownloadInfo
	id      int
	skip    int64
	chunk   *[]byte
	chunks  *[]chunkSize
	closed  bool
	sha     string
	shaTemp hash.Hash
}

// get the next chunk
func (oo *openObject) getChunk(ctx context.Context) (err error) {
	if oo.id >= len(*oo.chunks) {
		return io.EOF
	}
	var chunk []byte
	err = utils.Retry(3, time.Second, func() (err error) {
		chunk, err = getRawFiles(oo.d[oo.id])
		return err
	})
	if err != nil {
		return err
	}
	oo.id++
	oo.chunk = &chunk
	return nil
}

// Read reads up to len(p) bytes into p.
func (oo *openObject) Read(p []byte) (n int, err error) {
	oo.mu.Lock()
	defer oo.mu.Unlock()
	if oo.closed {
		return 0, fmt.Errorf("read on closed file")
	}
	// Skip data at the start if requested
	for oo.skip > 0 {
		//size := 1024 * 1024
		_, size, err := oo.ChunkLocation(oo.id)
		if err != nil {
			return 0, err
		}
		if oo.skip < int64(size) {
			break
		}
		oo.id++
		oo.skip -= int64(size)
	}
	if len(*oo.chunk) == 0 {
		err = oo.getChunk(oo.ctx)
		if err != nil {
			return 0, err
		}
		if oo.skip > 0 {
			*oo.chunk = (*oo.chunk)[oo.skip:]
			oo.skip = 0
		}
	}
	n = copy(p, *oo.chunk)
	*oo.chunk = (*oo.chunk)[n:]

	oo.shaTemp.Write(*oo.chunk)

	return n, nil
}

// Close closed the file - MAC errors are reported here
func (oo *openObject) Close() (err error) {
	oo.mu.Lock()
	defer oo.mu.Unlock()
	if oo.closed {
		return nil
	}
	// 校验Sha1
	if string(oo.shaTemp.Sum(nil)) != oo.sha {
		return fmt.Errorf("failed to finish download: %w", err)
	}

	oo.closed = true
	return nil
}

func GetMD5Hash(text string) string {
	tHash := md5.Sum([]byte(text))
	return hex.EncodeToString(tHash[:])
}

// chunkSize describes a size and position of chunk
type chunkSize struct {
	position int64
	size     int
}

func getChunkSizes(sliceSize []*pubUserFile.SliceSize) (chunks []chunkSize) {
	chunks = make([]chunkSize, 0)
	for _, s := range sliceSize {
		// 对最后一个做特殊处理
		if s.EndIndex == 0 {
			s.EndIndex = s.StartIndex
		}
		for j := s.StartIndex; j <= s.EndIndex; j++ {
			chunks = append(chunks, chunkSize{position: j, size: int(s.Size)})
		}
	}
	return chunks
}

func (oo *openObject) ChunkLocation(id int) (position int64, size int, err error) {
	if id < 0 || id >= len(*oo.chunks) {
		return 0, 0, errors.New("invalid arguments")
	}

	return (*oo.chunks)[id].position, (*oo.chunks)[id].size, nil
}

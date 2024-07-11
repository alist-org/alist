package halalcloud

import "google.golang.org/grpc"

func defaultOptions() halalOptions {
	return halalOptions{
		// onRefreshTokenRefreshed: func(string) {},
		grpcOptions: []grpc.DialOption{
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024 * 1024 * 32)),
			// grpc.WithMaxMsgSize(1024 * 1024 * 1024),
		},
	}
}

type HalalOption interface {
	apply(*halalOptions)
}

// halalOptions configure a RPC call. halalOptions are set by the HalalOption
// values passed to Dial.
type halalOptions struct {
	onTokenRefreshed func(accessToken string, accessTokenExpiredAt int64, refreshToken string, refreshTokenExpiredAt int64)
	grpcOptions      []grpc.DialOption
}

// funcDialOption wraps a function that modifies halalOptions into an
// implementation of the DialOption interface.
type funcDialOption struct {
	f func(*halalOptions)
}

func (fdo *funcDialOption) apply(do *halalOptions) {
	fdo.f(do)
}

func newFuncDialOption(f func(*halalOptions)) *funcDialOption {
	return &funcDialOption{
		f: f,
	}
}

func WithRefreshTokenRefreshedCallback(s func(accessToken string, accessTokenExpiredAt int64, refreshToken string, refreshTokenExpiredAt int64)) HalalOption {
	return newFuncDialOption(func(o *halalOptions) {
		o.onTokenRefreshed = s
	})
}

func WithGrpcDialOptions(opts ...grpc.DialOption) HalalOption {
	return newFuncDialOption(func(o *halalOptions) {
		o.grpcOptions = opts
	})
}

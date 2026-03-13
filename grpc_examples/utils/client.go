package utils

import (
	"time"

	"github.com/block-vision/sui-go-sdk/common/grpcconn"
	"github.com/block-vision/sui-go-sdk/constant"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type DefaultConfig struct {
	Target  string
	Token   string
	Timeout time.Duration
	Retries int
	UseTLS  bool
}

func NewDefaultConfig() *DefaultConfig {
	return &DefaultConfig{
		Target:  constant.SuiMainnetGrpcEndpoint,
		Token:   constant.SuiMainnetGrpcToken,
		Timeout: time.Second * 30,
		Retries: 3,
		UseTLS:  true,
	}
}

func CreateGrpcClient(config *DefaultConfig) *grpcconn.SuiGrpcClient {
	opts := []grpcconn.GrpcConnOption{
		grpcconn.WithTimeout(config.Timeout),
		grpcconn.WithRetryCount(config.Retries),
	}

	if config.UseTLS {
		opts = append(opts, grpcconn.WithDialOptions(
			grpc.WithTransportCredentials(credentials.NewTLS(nil)),
			grpc.WithMaxMsgSize(20*1024*1024), // 20MB
		))
	} else {
		opts = append(opts, grpcconn.WithDialOptions(
			grpc.WithMaxMsgSize(20*1024*1024), // 20MB
		))
	}

	return grpcconn.NewSuiGrpcClientWithAuth(config.Target, config.Token, opts...)
}

func CreateGrpcClientWithDefaults() *grpcconn.SuiGrpcClient {
	config := NewDefaultConfig()
	return CreateGrpcClient(config)
}

func CreateGrpcClientSimple(target, token string) *grpcconn.SuiGrpcClient {
	config := &DefaultConfig{
		Target:  target,
		Token:   token,
		Timeout: time.Second * 30,
		Retries: 3,
		UseTLS:  true,
	}
	return CreateGrpcClient(config)
}

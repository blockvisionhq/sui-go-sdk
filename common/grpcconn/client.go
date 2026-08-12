package grpcconn

import (
	"context"
	"fmt"

	v2 "github.com/block-vision/sui-go-sdk/pb/sui/rpc/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type SuiGrpcClient struct {
	conn                         *GrpcConn
	nameService                  v2.NameServiceClient
	ledgerService                v2.LedgerServiceClient
	stateService                 v2.StateServiceClient
	movePackageService           v2.MovePackageServiceClient
	subscriptionService          v2.SubscriptionServiceClient
	transactionExecutionService  v2.TransactionExecutionServiceClient
	signatureVerificationService v2.SignatureVerificationServiceClient
}

func NewSuiGrpcClient(target string, opts ...GrpcConnOption) *SuiGrpcClient {
	conn := NewGrpcConn(target, opts...)
	return &SuiGrpcClient{
		conn: conn,
	}
}

// NewSuiGrpcClientFromConn builds a SuiGrpcClient over an externally-owned
// *grpc.ClientConn. The caller retains ownership of conn; Close on the returned
// client is a no-op for the underlying connection.
func NewSuiGrpcClientFromConn(conn *grpc.ClientConn) *SuiGrpcClient {
	return &SuiGrpcClient{
		conn: NewGrpcConnFromClientConn(conn),
	}
}

func NewSuiGrpcClientWithAuth(target, token string, opts ...GrpcConnOption) *SuiGrpcClient {
	authOpts := []GrpcConnOption{
		WithDialOptions(
			grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
				md := metadata.Pairs(
					"authorization", "Bearer "+token,
					"x-api-key", token,
					"x-token", token,
				)
				ctx = metadata.NewOutgoingContext(ctx, md)
				return invoker(ctx, method, req, reply, cc, callOpts...)
			}),

			grpc.WithChainStreamInterceptor(func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				md := metadata.Pairs(
					"authorization", "Bearer "+token,
					"x-api-key", token,
					"x-token", token,
				)
				ctx = metadata.NewOutgoingContext(ctx, md)
				return streamer(ctx, desc, cc, method, opts...)
			}),
		),
	}

	allOpts := append(authOpts, opts...)
	conn := NewGrpcConn(target, allOpts...)
	return &SuiGrpcClient{conn: conn}
}

func (c *SuiGrpcClient) Connect(ctx context.Context) error {
	if err := c.conn.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	grpcConn, err := c.conn.GetConn(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %v", err)
	}

	c.nameService = v2.NewNameServiceClient(grpcConn)
	c.ledgerService = v2.NewLedgerServiceClient(grpcConn)
	c.stateService = v2.NewStateServiceClient(grpcConn)
	c.movePackageService = v2.NewMovePackageServiceClient(grpcConn)
	c.subscriptionService = v2.NewSubscriptionServiceClient(grpcConn)
	c.transactionExecutionService = v2.NewTransactionExecutionServiceClient(grpcConn)
	c.signatureVerificationService = v2.NewSignatureVerificationServiceClient(grpcConn)
	return nil
}

func (c *SuiGrpcClient) Close() error {
	return c.conn.Close()
}

func (c *SuiGrpcClient) NameService(ctx context.Context) (v2.NameServiceClient, error) {
	if c.nameService == nil {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}
	return c.nameService, nil
}

func (c *SuiGrpcClient) LedgerService(ctx context.Context) (v2.LedgerServiceClient, error) {
	if c.ledgerService == nil {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}
	return c.ledgerService, nil
}

func (c *SuiGrpcClient) StateService(ctx context.Context) (v2.StateServiceClient, error) {
	if c.stateService == nil {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}
	return c.stateService, nil
}

func (c *SuiGrpcClient) MovePackageService(ctx context.Context) (v2.MovePackageServiceClient, error) {
	if c.movePackageService == nil {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}
	return c.movePackageService, nil
}

func (c *SuiGrpcClient) SubscriptionService(ctx context.Context) (v2.SubscriptionServiceClient, error) {
	if c.subscriptionService == nil {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}
	return c.subscriptionService, nil
}

func (c *SuiGrpcClient) TransactionExecutionService(ctx context.Context) (v2.TransactionExecutionServiceClient, error) {
	if c.transactionExecutionService == nil {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}
	return c.transactionExecutionService, nil
}

func (c *SuiGrpcClient) SignatureVerificationService(ctx context.Context) (v2.SignatureVerificationServiceClient, error) {
	if c.signatureVerificationService == nil {
		if err := c.Connect(ctx); err != nil {
			return nil, err
		}
	}
	return c.signatureVerificationService, nil
}

func (c *SuiGrpcClient) CallWithRetry(ctx context.Context, method string, req interface{}, reply interface{}) error {
	return c.conn.Call(ctx, method, req, reply)
}

func (c *SuiGrpcClient) GetConnection(ctx context.Context) (*grpc.ClientConn, error) {
	return c.conn.GetConn(ctx)
}

func (c *SuiGrpcClient) GetMetadata(ctx context.Context) (metadata.MD, bool) {
	return metadata.FromOutgoingContext(ctx)
}

func (c *SuiGrpcClient) CreateContextWithMetadata(ctx context.Context, pairs ...string) context.Context {
	md := metadata.Pairs(pairs...)
	return metadata.NewOutgoingContext(ctx, md)
}

type BatchRequest struct {
	Method string
	Req    interface{}
	Reply  interface{}
}

type BatchResponse struct {
	Index int
	Error error
}

func (c *SuiGrpcClient) BatchCall(ctx context.Context, requests []BatchRequest) []BatchResponse {
	responses := make([]BatchResponse, len(requests))

	ch := make(chan BatchResponse, len(requests))

	for i, req := range requests {
		go func(index int, request BatchRequest) {
			err := c.CallWithRetry(ctx, request.Method, request.Req, request.Reply)
			ch <- BatchResponse{
				Index: index,
				Error: err,
			}
		}(i, req)
	}

	for i := 0; i < len(requests); i++ {
		resp := <-ch
		responses[resp.Index] = resp
	}

	return responses
}

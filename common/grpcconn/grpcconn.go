package grpcconn

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const (
	defaultTimeout          = time.Second * 30
	defaultRetryCount       = 3
	defaultKeepAlive        = time.Second * 30
	defaultKeepaliveTimeout = time.Second * 5
)

type GrpcConn struct {
	mu         sync.RWMutex
	conn       *grpc.ClientConn
	rl         *rate.Limiter
	target     string
	retryCount int
	timeout    time.Duration
	dialOpts   []grpc.DialOption
	external   bool
}

type GrpcConnOption func(*GrpcConn)

func WithRetryCount(count int) GrpcConnOption {
	return func(g *GrpcConn) {
		g.retryCount = count
	}
}

func WithTimeout(timeout time.Duration) GrpcConnOption {
	return func(g *GrpcConn) {
		g.timeout = timeout
	}
}

func WithDialOptions(opts ...grpc.DialOption) GrpcConnOption {
	return func(g *GrpcConn) {
		g.dialOpts = append(g.dialOpts, opts...)
	}
}

func WithRateLimiter(rl *rate.Limiter) GrpcConnOption {
	return func(g *GrpcConn) {
		g.rl = rl
	}
}

func newDefaultRateLimiter() *rate.Limiter {
	return rate.NewLimiter(rate.Every(1*time.Second), 1000) // 1000 request every 1 seconds
}

func defaultDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                defaultKeepAlive,
			Timeout:             defaultKeepaliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithMaxMsgSize(20 * 1024 * 1024), // 20MB
	}
}

func NewGrpcConn(target string, opts ...GrpcConnOption) *GrpcConn {
	g := &GrpcConn{
		target:     target,
		timeout:    defaultTimeout,
		retryCount: defaultRetryCount,
		dialOpts:   defaultDialOptions(),
		// rl:         newDefaultRateLimiter(),
	}

	for _, opt := range opts {
		opt(g)
	}

	return g
}

// NewGrpcConnFromClientConn wraps an externally-owned *grpc.ClientConn so callers
// can reuse their own connection instead of letting the SDK dial one. The provided
// conn's lifecycle stays with the caller: Close is a no-op on it.
func NewGrpcConnFromClientConn(conn *grpc.ClientConn) *GrpcConn {
	return &GrpcConn{
		conn:       conn,
		timeout:    defaultTimeout,
		retryCount: defaultRetryCount,
		external:   true,
	}
}

func (g *GrpcConn) Connect(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.conn != nil {
		return nil
	}

	conn, err := grpc.DialContext(ctx, g.target, g.dialOpts...)
	if err != nil {
		return err
	}

	g.conn = conn
	return nil
}

func (g *GrpcConn) GetConn(ctx context.Context) (*grpc.ClientConn, error) {
	g.mu.RLock()
	if g.conn != nil && g.conn.GetState().String() != "SHUTDOWN" {
		conn := g.conn
		g.mu.RUnlock()
		return conn, nil
	}
	g.mu.RUnlock()

	if err := g.Connect(ctx); err != nil {
		return nil, err
	}

	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.conn, nil
}

func (g *GrpcConn) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.external {
		// Externally-owned connection: caller owns its lifecycle, don't close it.
		g.conn = nil
		return nil
	}

	if g.conn != nil {
		err := g.conn.Close()
		g.conn = nil
		return err
	}
	return nil
}

func (g *GrpcConn) Call(ctx context.Context, method string, req interface{}, reply interface{}) error {
	if g.rl != nil {
		if err := g.rl.Wait(ctx); err != nil {
			return fmt.Errorf("rate limit error: %v", err)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, g.timeout)
	defer cancel()

	var lastErr error

	for i := 0; i <= g.retryCount; i++ {
		conn, err := g.GetConn(ctx)
		if err != nil {
			lastErr = fmt.Errorf("get connection error: %v", err)
			continue
		}

		err = conn.Invoke(ctx, method, req, reply)
		if err == nil {
			return nil
		}

		lastErr = fmt.Errorf("invoke %s error (attempt %d): %v", method, i+1, err)

		if isConnectionError(err) {
			g.mu.Lock()
			if g.conn != nil {
				g.conn.Close()
				g.conn = nil
			}
			g.mu.Unlock()
		}

		if i < g.retryCount {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Millisecond * 100 * time.Duration(i+1)):
			}
		}
	}

	return lastErr
}

func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "connection", "unavailable", "refused", "timeout", "deadline")
}

func contains(str string, keywords ...string) bool {
	for _, keyword := range keywords {
		if len(str) >= len(keyword) {
			for i := 0; i <= len(str)-len(keyword); i++ {
				if str[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

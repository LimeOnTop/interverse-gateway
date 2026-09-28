package clients

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Retries transient UNAVAILABLE (Docker DNS blips) without hanging forever.
// waitForReady must stay false — otherwise a dead backend waits until nginx 504.
const grpcRetryServiceConfig = `{
  "methodConfig": [{
    "name": [{"service": ""}],
    "waitForReady": false,
    "retryPolicy": {
      "maxAttempts": 4,
      "initialBackoff": "0.05s",
      "maxBackoff": "0.5s",
      "backoffMultiplier": 2,
      "retryableStatusCodes": ["UNAVAILABLE"]
    }
  }]
}`

const defaultRPCTimeout = 12 * time.Second

func dialGRPC(target, serviceName string) *grpc.ClientConn {
	return dialGRPCWithTimeout(target, serviceName, defaultRPCTimeout)
}

func dialGRPCWithTimeout(target, serviceName string, timeout time.Duration) *grpc.ClientConn {
	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(grpcRetryServiceConfig),
		grpc.WithUnaryInterceptor(rpcTimeoutInterceptor(timeout)),
	)
	if err != nil {
		panic(fmt.Sprintf("connect to %s (%s): %v", serviceName, target, err))
	}
	log.Printf("grpc client ready: %s -> %s (rpc timeout %s)", serviceName, target, timeout)
	return conn
}

func rpcTimeoutInterceptor(timeout time.Duration) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if _, hasDeadline := ctx.Deadline(); !hasDeadline {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, timeout)
			defer cancel()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

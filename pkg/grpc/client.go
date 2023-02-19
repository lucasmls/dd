package grpc

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	gGRPC "google.golang.org/grpc"
)

// Client is the gRPC client itself
type Client struct {
	address               string
	logger                *zap.SugaredLogger
	additionalDialOptions []gGRPC.DialOption
}

// NewClient is the gRPC client constructor
func NewClient(
	address string,
	logger *zap.SugaredLogger,
	additionalDialoptions []gGRPC.DialOption,
) (*Client, error) {
	return &Client{
		address:               address,
		logger:                logger,
		additionalDialOptions: additionalDialoptions,
	}, nil
}

// MustNewClient is the gRPC client constructor
// It panics if any error is found
func MustNewClient(
	address string,
	logger *zap.SugaredLogger,
	additionalDialoptions []gGRPC.DialOption,
) *Client {
	client, err := NewClient(
		address,
		logger,
		additionalDialoptions,
	)
	if err != nil {
		panic(err)
	}

	return client
}

// Connect into a gRPC server
func (c Client) Connect(ctx context.Context) (*gGRPC.ClientConn, error) {
	c.logger.Info("connecting to gRPC server", zap.String("address", c.address))

	dialOptions := []gGRPC.DialOption{
		// Note the use of insecure transport here. TLS is recommended in production.
		gGRPC.WithInsecure(),
		gGRPC.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()),
		gGRPC.WithStreamInterceptor(otelgrpc.StreamClientInterceptor()),
	}

	dialOptions = append(dialOptions, c.additionalDialOptions...)

	conn, err := gGRPC.DialContext(
		ctx,
		c.address,
		dialOptions...,
	)
	if err != nil {
		c.logger.Error("failed to connect into gRPC server", zap.Error(err))
		return nil, err
	}

	return conn, nil
}

// MustConnect into a gRPC server
// It panics if any error is found
func (c Client) MustConnect(ctx context.Context) *gGRPC.ClientConn {
	connection, err := c.Connect(ctx)
	if err != nil {
		panic(err)
	}

	return connection
}

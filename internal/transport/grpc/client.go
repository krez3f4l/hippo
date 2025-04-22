package grpcclient

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/krez3f4l/audit_logger/pkg/domain/audit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"

	"hippo/internal/platform/config"
)

const timeoutDial = 2 * time.Second

type Client struct {
	conn        *grpc.ClientConn
	auditClient audit.AuditServiceClient
	timeout     time.Duration
}

func NewClient(cfg config.GrpcAuditClient) (*Client, error) {
	address := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))

	var (
		creds credentials.TransportCredentials
		err   error
	)

	if cfg.CertFilePath != "" {
		// dev or prod env
		creds, err = credentials.NewClientTLSFromFile(cfg.CertFilePath, "")
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS cert: %w", err)
		}
	} else {
		// local env, not required
		creds = insecure.NewCredentials()
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeoutDial)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(creds),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to dial gRPC server: %w", err)
	}

	return &Client{
		conn:        conn,
		auditClient: audit.NewAuditServiceClient(conn),
		timeout:     cfg.Timeout,
	}, nil
}

func (c *Client) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) SendLogRequest(ctx context.Context, req audit.LogItem) error {
	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	action, err := audit.ToPbAction(req.Action)
	if err != nil {
		return fmt.Errorf("invalid action: %w", err)
	}

	entity, err := audit.ToPbEntity(req.Entity)
	if err != nil {
		return fmt.Errorf("invalid entity: %w", err)
	}

	_, err = c.auditClient.Log(ctx, &audit.LogRequest{
		Action:    action,
		Entity:    entity,
		EntityId:  req.EntityID,
		Timestamp: timestamppb.New(req.Timestamp),
	})

	if err != nil {
		return fmt.Errorf("failed to send audit log: %w", err)
	}

	return nil
}

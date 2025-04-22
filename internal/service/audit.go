package service

import (
	"context"

	"github.com/krez3f4l/audit_logger/pkg/domain/audit"
)

type AuditClient interface {
	SendLogRequest(ctx context.Context, req audit.LogItem) error
}

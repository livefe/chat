package service

import (
	"context"

	"app/pkg/errs"
	"app/pkg/logger"
)

func clientErrorText(ctx context.Context, log *logger.Logger, err error, fallback string) string {
	if msg, ok := errs.ClientMessage(err); ok {
		return msg
	}
	log.Error(ctx, fallback, logger.Error(err))
	return fallback
}

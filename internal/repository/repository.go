package repository

import (
	"context"
	"iot-server/internal/model"
	"time"
)

const defaultQueryTimeout = 5 * time.Second

// helper
func pageMeta(page, pageSize int, total int64) *model.PageMetadata {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	totalPage := (total + int64(pageSize) - 1) / int64(pageSize)
	return &model.PageMetadata{
		Page:      page,
		Size:      pageSize,
		TotalItem: total,
		TotalPage: totalPage,
	}
}

func ctxWithTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, defaultQueryTimeout)
}

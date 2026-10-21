package types

import "time"

const cachePrefix = "app:"

const (
	Expire45s = 45 * time.Second
	Expire5m  = 5 * time.Minute
	Expire10m = 10 * time.Minute
	Expire1h  = 1 * time.Hour
	Expire1d  = 24 * time.Hour
)

type IDRequest struct {
	ID uint `json:"id" form:"id" uri:"id" binding:"required"`
}

type IDResponse struct {
	ID uint `json:"id"`
}

type PageQuery struct {
	Page     int `form:"page" binding:"required,min=1"`
	PageSize int `form:"pageSize" binding:"required,min=1,max=200"`
}

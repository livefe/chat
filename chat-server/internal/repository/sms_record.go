package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"
)

type SMSRecordRepository interface {
	Create(ctx context.Context, record *model.SMSRecord) error
}

type smsRecordRepository struct {
	db *database.DB
}

func NewSMSRecordRepository(db *database.DB) SMSRecordRepository {
	return &smsRecordRepository{db: db}
}

func (r *smsRecordRepository) Create(ctx context.Context, record *model.SMSRecord) error {
	return database.Conn(ctx, r.db).Create(record).Error
}

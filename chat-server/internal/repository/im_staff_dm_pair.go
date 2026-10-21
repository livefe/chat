package repository

import (
	"context"

	"app/internal/model"
	"app/pkg/database"

	"gorm.io/gorm/clause"
)

type IMStaffDMPairRepository interface {
	GetByEmployees(ctx context.Context, empMin, empMax uint) (*model.IMStaffDMPair, error)
	Create(ctx context.Context, pair *model.IMStaffDMPair) error
}

type imStaffDMPairRepository struct {
	db *database.DB
}

func NewIMStaffDMPairRepository(db *database.DB) IMStaffDMPairRepository {
	return &imStaffDMPairRepository{db: db}
}

func (r *imStaffDMPairRepository) GetByEmployees(ctx context.Context, empMin, empMax uint) (*model.IMStaffDMPair, error) {
	var pair model.IMStaffDMPair
	err := database.Conn(ctx, r.db).
		Where("emp_min_id = ? AND emp_max_id = ?", empMin, empMax).
		First(&pair).Error
	if err != nil {
		return nil, err
	}
	return &pair, nil
}

func (r *imStaffDMPairRepository) Create(ctx context.Context, pair *model.IMStaffDMPair) error {
	return database.Conn(ctx, r.db).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "emp_min_id"}, {Name: "emp_max_id"}},
		DoNothing: true,
	}).Create(pair).Error
}

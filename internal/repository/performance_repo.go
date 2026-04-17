package repository

import (
	"context"
	"go-attendance-api/internal/model"
	"gorm.io/gorm"
)

type PerformanceRepository interface {
	// Goals
	FindGoalsByUserID(ctx context.Context, userID uint) ([]model.PerformanceGoal, error)
	FindGoalByID(ctx context.Context, id uint) (*model.PerformanceGoal, error)
	CreateGoal(ctx context.Context, goal *model.PerformanceGoal) error
	UpdateGoal(ctx context.Context, goal *model.PerformanceGoal) error

	// Cycles
	FindAllCycles(ctx context.Context) ([]model.PerformanceCycle, error)
	FindCycleByID(ctx context.Context, id uint) (*model.PerformanceCycle, error)

	// Appraisals
	FindAppraisalsByCycleID(ctx context.Context, cycleID uint) ([]model.Appraisal, error)
	FindAppraisalByID(ctx context.Context, id uint) (*model.Appraisal, error)
	UpdateAppraisal(ctx context.Context, appraisal *model.Appraisal) error
}

type performanceRepository struct {
	db *gorm.DB
}

func NewPerformanceRepository(db *gorm.DB) PerformanceRepository {
	return &performanceRepository{db: db}
}

func (r *performanceRepository) FindGoalsByUserID(ctx context.Context, userID uint) ([]model.PerformanceGoal, error) {
	var goals []model.PerformanceGoal
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&goals).Error
	return goals, err
}

func (r *performanceRepository) FindGoalByID(ctx context.Context, id uint) (*model.PerformanceGoal, error) {
	var goal model.PerformanceGoal
	err := r.db.WithContext(ctx).First(&goal, id).Error
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

func (r *performanceRepository) CreateGoal(ctx context.Context, goal *model.PerformanceGoal) error {
	return r.db.WithContext(ctx).Create(goal).Error
}

func (r *performanceRepository) UpdateGoal(ctx context.Context, goal *model.PerformanceGoal) error {
	return r.db.WithContext(ctx).Save(goal).Error
}

func (r *performanceRepository) FindAllCycles(ctx context.Context) ([]model.PerformanceCycle, error) {
	var cycles []model.PerformanceCycle
	err := r.db.WithContext(ctx).Find(&cycles).Error
	return cycles, err
}

func (r *performanceRepository) FindCycleByID(ctx context.Context, id uint) (*model.PerformanceCycle, error) {
	var cycle model.PerformanceCycle
	err := r.db.WithContext(ctx).First(&cycle, id).Error
	if err != nil {
		return nil, err
	}
	return &cycle, nil
}

func (r *performanceRepository) FindAppraisalsByCycleID(ctx context.Context, cycleID uint) ([]model.Appraisal, error) {
	var appraisals []model.Appraisal
	err := r.db.WithContext(ctx).Where("cycle_id = ?", cycleID).Preload("User").Find(&appraisals).Error
	return appraisals, err
}

func (r *performanceRepository) FindAppraisalByID(ctx context.Context, id uint) (*model.Appraisal, error) {
	var appraisal model.Appraisal
	err := r.db.WithContext(ctx).First(&appraisal, id).Error
	if err != nil {
		return nil, err
	}
	return &appraisal, nil
}

func (r *performanceRepository) UpdateAppraisal(ctx context.Context, appraisal *model.Appraisal) error {
	return r.db.WithContext(ctx).Save(appraisal).Error
}

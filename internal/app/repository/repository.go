package repository

import (
	"database/sql"
	"errors"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pediatric-dose-backend/internal/app/ds"
)

var (
	ErrDrugNotFound    = errors.New("препарат не найден")
	ErrDrugDraftExists = errors.New("у пользователя уже есть препарат в статусе черновик")
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(dsn string) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return &Repository{db: db}, nil
}

func (r *Repository) GetPublishedDrugs() ([]ds.Drug, error) {
	var drugs []ds.Drug
	err := r.db.Where("drug_status = ?", ds.DrugStatusPublished).Order("id").Find(&drugs).Error
	if err != nil {
		return nil, err
	}

	return drugs, nil
}

func (r *Repository) GetPublishedDrugsByAdultDose(minAdultDoseMg, maxAdultDoseMg float64) ([]ds.Drug, error) {
	var drugs []ds.Drug
	err := r.db.Where("drug_status = ? AND recommended_adult_dose_mg BETWEEN ? AND ?", ds.DrugStatusPublished, minAdultDoseMg, maxAdultDoseMg).
		Order("id").
		Find(&drugs).Error
	if err != nil {
		return nil, err
	}

	return drugs, nil
}

func (r *Repository) GetDrugByID(drugID int) (ds.Drug, error) {
	var drug ds.Drug
	err := r.db.Where("id = ? AND drug_status = ?", drugID, ds.DrugStatusPublished).First(&drug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ds.Drug{}, ErrDrugNotFound
	}

	return drug, err
}

func (r *Repository) GetNextDrug(drugID int) (ds.Drug, error) {
	var drug ds.Drug
	err := r.db.Where("id > ? AND drug_status = ?", drugID, ds.DrugStatusPublished).Order("id").First(&drug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.GetFirstPublishedDrug()
	}

	return drug, err
}

func (r *Repository) GetFirstPublishedDrug() (ds.Drug, error) {
	var drug ds.Drug
	err := r.db.Where("drug_status = ?", ds.DrugStatusPublished).Order("id").First(&drug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ds.Drug{}, ErrDrugNotFound
	}

	return drug, err
}

func (r *Repository) GetDraftDrug(creatorID uint) (ds.Drug, error) {
	var drug ds.Drug
	err := r.db.Where("creator_id = ? AND drug_status = ?", creatorID, ds.DrugStatusDraft).First(&drug).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ds.Drug{}, ErrDrugNotFound
	}

	return drug, err
}

func (r *Repository) GetDrugLikesCount(drugID uint) (int64, error) {
	var likesCount int64
	err := r.db.Model(&ds.DrugLike{}).Where("drug_id = ?", drugID).Count(&likesCount).Error

	return likesCount, err
}

func (r *Repository) CreateDrugDraft(creatorID uint, drugName string) error {
	_, err := r.GetDraftDrug(creatorID)
	if err == nil {
		return ErrDrugDraftExists
	}
	if !errors.Is(err, ErrDrugNotFound) {
		return err
	}

	drug := ds.Drug{
		DrugName:   drugName,
		DrugStatus: ds.DrugStatusDraft,
		CreatorID:  creatorID,
	}

	return r.db.Create(&drug).Error
}

func (r *Repository) PublishDrugDraft(creatorID uint, shortInfo string, recommendedAdultDoseMg, maxDailyDoseMg float64) (uint, error) {
	drug, err := r.GetDraftDrug(creatorID)
	if err != nil {
		return 0, err
	}

	err = r.db.Model(&drug).Updates(map[string]interface{}{
		"short_info":                shortInfo,
		"recommended_adult_dose_mg": recommendedAdultDoseMg,
		"max_daily_dose_mg":         maxDailyDoseMg,
		"drug_status":               ds.DrugStatusPublished,
		"published_at":              time.Now(),
	}).Error
	if err != nil {
		return 0, err
	}

	return drug.ID, nil
}

func (r *Repository) DeleteDrug(drugID int) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	query := "UPDATE drugs SET drug_status = 'deleted' WHERE id = $1 AND drug_status = 'published' RETURNING id"

	row := sqlDB.QueryRow(query, drugID)

	var deletedDrugID uint
	err = row.Scan(&deletedDrugID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDrugNotFound
	}

	return err
}

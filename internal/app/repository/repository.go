package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/minio/minio-go/v7"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"pediatric-dose-backend/internal/app/ds"
)

var (
	ErrDrugNotFound    = errors.New("препарат не найден")
	ErrDrugDraftExists = errors.New("у пользователя уже есть препарат в статусе черновик")
	ErrUserNotFound    = errors.New("пользователь не найден")
	ErrUserLoginTaken  = errors.New("логин занят")
)

type Repository struct {
	db            *gorm.DB
	minio         *minio.Client
	minioEndpoint string
	minioBucket   string
}

func NewRepository(dsn string, minioConfig MinioConfig) (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	minioClient, err := newMinioClient(minioConfig)
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:            db,
		minio:         minioClient,
		minioEndpoint: minioConfig.Endpoint,
		minioBucket:   minioConfig.Bucket,
	}, nil
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

func (r *Repository) CreateDrugDraft(creatorID uint, drugName, imageURL, videoURL string) (ds.Drug, error) {
	_, err := r.GetDraftDrug(creatorID)
	if err == nil {
		return ds.Drug{}, ErrDrugDraftExists
	}
	if !errors.Is(err, ErrDrugNotFound) {
		return ds.Drug{}, err
	}

	drug := ds.Drug{
		DrugName:   drugName,
		DrugStatus: ds.DrugStatusDraft,
		ImageURL:   imageURL,
		VideoURL:   videoURL,
		CreatorID:  creatorID,
	}

	if err := r.db.Create(&drug).Error; err != nil {
		return ds.Drug{}, err
	}

	return drug, nil
}

func (r *Repository) PublishDrugDraft(creatorID uint, shortInfo string, recommendedAdultDoseMg, maxDailyDoseMg float64) (ds.Drug, error) {
	drug, err := r.GetDraftDrug(creatorID)
	if err != nil {
		return ds.Drug{}, err
	}

	err = r.db.Model(&drug).Updates(map[string]interface{}{
		"short_info":                shortInfo,
		"recommended_adult_dose_mg": recommendedAdultDoseMg,
		"max_daily_dose_mg":         maxDailyDoseMg,
		"drug_status":               ds.DrugStatusPublished,
		"published_at":              time.Now(),
	}).Error
	if err != nil {
		return ds.Drug{}, err
	}

	return r.GetDrugByID(int(drug.ID))
}

func (r *Repository) DeleteDrug(drugID int, creatorID uint) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}

	query := "UPDATE drugs SET drug_status = 'deleted' WHERE id = $1 AND creator_id = $2 AND drug_status = 'published' RETURNING id"

	row := sqlDB.QueryRow(query, drugID, creatorID)

	var deletedDrugID uint
	err = row.Scan(&deletedDrugID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrDrugNotFound
	}

	return err
}

func (r *Repository) SetDrugLike(userID uint, drugID int) error {
	drug, err := r.GetDrugByID(drugID)
	if err != nil {
		return err
	}

	like := ds.DrugLike{
		UserID: userID,
		DrugID: drug.ID,
	}

	return r.db.Where("user_id = ? AND drug_id = ?", userID, drug.ID).FirstOrCreate(&like).Error
}

func (r *Repository) RemoveDrugLike(userID uint, drugID int) error {
	drug, err := r.GetDrugByID(drugID)
	if err != nil {
		return err
	}

	return r.db.Where("user_id = ? AND drug_id = ?", userID, drug.ID).Delete(&ds.DrugLike{}).Error
}

func (r *Repository) CreateUser(login, password string) (ds.User, error) {
	var existing ds.User
	err := r.db.Where("login = ?", login).First(&existing).Error
	if err == nil {
		return ds.User{}, ErrUserLoginTaken
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return ds.User{}, err
	}

	user := ds.User{
		Login:    login,
		Password: password,
	}

	if err := r.db.Create(&user).Error; err != nil {
		return ds.User{}, err
	}

	return user, nil
}

func (r *Repository) GetUserByLogin(login string) (ds.User, error) {
	var user ds.User
	err := r.db.Where("login = ?", login).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ds.User{}, ErrUserNotFound
	}

	return user, err
}

func (r *Repository) SetDrugMedia(drugID uint, imageURL, videoURL string) (ds.Drug, error) {
	err := r.db.Model(&ds.Drug{}).Where("id = ?", drugID).Updates(map[string]interface{}{
		"image_url": imageURL,
		"video_url": videoURL,
	}).Error
	if err != nil {
		return ds.Drug{}, err
	}

	var drug ds.Drug
	err = r.db.Where("id = ?", drugID).First(&drug).Error

	return drug, err
}

package schemes

import "pediatric-dose-backend/internal/app/ds"

type Drug struct {
	ID                     uint    `json:"id"`
	DrugName               string  `json:"drug_name"`
	ShortInfo              string  `json:"short_info"`
	ImageURL               string  `json:"image_url"`
	VideoURL               string  `json:"video_url"`
	RecommendedAdultDoseMg float64 `json:"recommended_adult_dose_mg"`
	MaxDailyDoseMg         float64 `json:"max_daily_dose_mg"`
	LikesCount             int64   `json:"likes_count"`
	IsCreator              int     `json:"is_creator"`
}

type CreateDrugRequest struct {
	DrugName string `form:"drug_name" binding:"required,max=100"`
}

type PublishDrugRequest struct {
	ShortInfo              string  `json:"short_info" binding:"required,max=500"`
	RecommendedAdultDoseMg float64 `json:"recommended_adult_dose_mg" binding:"required,gt=0"`
	MaxDailyDoseMg         float64 `json:"max_daily_dose_mg" binding:"required,gt=0"`
}

type LikeDrugRequest struct {
	Value *int `json:"value" binding:"required,oneof=0 1"`
}

type RegisterUserRequest struct {
	Login    string `json:"login" binding:"required,max=25"`
	Password string `json:"password" binding:"required,max=100"`
}

type LoginUserRequest struct {
	Login    string `json:"login" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type User struct {
	ID    uint   `json:"id"`
	Login string `json:"login"`
}

func NewDrug(drug ds.Drug, likesCount int64, currentUserID uint) Drug {
	isCreator := 0
	if drug.CreatorID == currentUserID {
		isCreator = 1
	}

	return Drug{
		ID:                     drug.ID,
		DrugName:               drug.DrugName,
		ShortInfo:              drug.ShortInfo.String,
		ImageURL:               drug.ImageURL,
		VideoURL:               drug.VideoURL,
		RecommendedAdultDoseMg: drug.RecommendedAdultDoseMg.Float64,
		MaxDailyDoseMg:         drug.MaxDailyDoseMg.Float64,
		LikesCount:             likesCount,
		IsCreator:              isCreator,
	}
}

func NewUser(user ds.User) User {
	return User{
		ID:    user.ID,
		Login: user.Login,
	}
}

package handler

import (
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"pediatric-dose-backend/internal/app/ds"
	"pediatric-dose-backend/internal/app/repository"
	"pediatric-dose-backend/internal/app/schemes"
	"pediatric-dose-backend/internal/app/singleton"
)

const (
	adultDoseFilterMinMg = 0
	adultDoseFilterMaxMg = 1000
)

type Handler struct {
	Repository *repository.Repository
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetDrugs(ctx *gin.Context) {
	minAdultDoseValue := ctx.Query("min_adult_dose")
	maxAdultDoseValue := ctx.Query("max_adult_dose")

	minAdultDoseMg := parseAdultDose(minAdultDoseValue, adultDoseFilterMinMg)
	maxAdultDoseMg := parseAdultDose(maxAdultDoseValue, adultDoseFilterMaxMg)
	if minAdultDoseMg > maxAdultDoseMg {
		minAdultDoseMg, maxAdultDoseMg = maxAdultDoseMg, minAdultDoseMg
	}

	var drugs []ds.Drug
	var err error

	if minAdultDoseValue == "" && maxAdultDoseValue == "" {
		drugs, err = h.Repository.GetPublishedDrugs()
	} else {
		drugs, err = h.Repository.GetPublishedDrugsByAdultDose(minAdultDoseMg, maxAdultDoseMg)
	}
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	response := make([]schemes.Drug, 0, len(drugs))
	for _, drug := range drugs {
		response = append(response, h.newDrugResponse(drug))
	}

	ctx.JSON(http.StatusOK, response)
}

func (h *Handler) GetDrugFeed(ctx *gin.Context) {
	drugIDValue := strings.Trim(ctx.Param("drug_id"), "/")

	var drug ds.Drug
	var err error

	if drugIDValue == "" {
		drug, err = h.Repository.GetFirstPublishedDrug()
	} else {
		drugID, convertErr := strconv.Atoi(drugIDValue)
		if convertErr != nil {
			ctx.Status(http.StatusNotFound)
			return
		}

		if ctx.Query("next") == "true" {
			drug, err = h.Repository.GetNextDrug(drugID)
		} else {
			drug, err = h.Repository.GetDrugByID(drugID)
		}
	}
	if err != nil {
		h.abortWithDrugError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, h.newDrugResponse(drug))
}

func (h *Handler) GetDrugDraft(ctx *gin.Context) {
	drug, err := h.Repository.GetDraftDrug(singleton.GetCurrentUser().ID)
	if err != nil {
		h.abortWithDrugError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, h.newDrugResponse(drug))
}

func (h *Handler) CreateDrug(ctx *gin.Context) {
	var request schemes.CreateDrugRequest
	if err := ctx.ShouldBind(&request); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	image, err := ctx.FormFile("image")
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	video, err := ctx.FormFile("video")
	if err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	drug, err := h.Repository.CreateDrugDraft(singleton.GetCurrentUser().ID, strings.TrimSpace(request.DrugName), "", "")
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	imageURL, err := h.uploadDrugMedia(drug.ID, image, "image/jpeg", "jpg")
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	videoURL, err := h.uploadDrugMedia(drug.ID, video, "video/mp4", "mp4")
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	drug, err = h.Repository.SetDrugMedia(drug.ID, imageURL, videoURL)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusInternalServerError)
		return
	}

	ctx.JSON(http.StatusCreated, h.newDrugResponse(drug))
}

func (h *Handler) uploadDrugMedia(drugID uint, file *multipart.FileHeader, contentType, extension string) (string, error) {
	source, err := file.Open()
	if err != nil {
		return "", err
	}
	defer source.Close()

	objectName := fmt.Sprintf("drug_%d.%s", drugID, extension)

	return h.Repository.UploadDrugMedia(objectName, contentType, source, file.Size)
}

func (h *Handler) PublishDrug(ctx *gin.Context) {
	var request schemes.PublishDrugRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	if request.MaxDailyDoseMg < request.RecommendedAdultDoseMg {
		ctx.Status(http.StatusBadRequest)
		return
	}

	drug, err := h.Repository.PublishDrugDraft(
		singleton.GetCurrentUser().ID,
		strings.TrimSpace(request.ShortInfo),
		request.RecommendedAdultDoseMg,
		request.MaxDailyDoseMg,
	)
	if err != nil {
		h.abortWithDrugError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, h.newDrugResponse(drug))
}

func (h *Handler) DeleteDrug(ctx *gin.Context) {
	drugID, err := strconv.Atoi(ctx.Param("drug_id"))
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	if err := h.Repository.DeleteDrug(drugID, singleton.GetCurrentUser().ID); err != nil {
		h.abortWithDrugError(ctx, err)
		return
	}

	ctx.Status(http.StatusOK)
}

func (h *Handler) LikeDrug(ctx *gin.Context) {
	drugID, err := strconv.Atoi(ctx.Param("drug_id"))
	if err != nil {
		ctx.Status(http.StatusNotFound)
		return
	}

	var request schemes.LikeDrugRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	currentUserID := singleton.GetCurrentUser().ID

	if *request.Value == 1 {
		err = h.Repository.SetDrugLike(currentUserID, drugID)
	} else {
		err = h.Repository.RemoveDrugLike(currentUserID, drugID)
	}
	if err != nil {
		h.abortWithDrugError(ctx, err)
		return
	}

	drug, err := h.Repository.GetDrugByID(drugID)
	if err != nil {
		h.abortWithDrugError(ctx, err)
		return
	}

	ctx.JSON(http.StatusOK, h.newDrugResponse(drug))
}

func (h *Handler) RegisterUser(ctx *gin.Context) {
	var request schemes.RegisterUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	user, err := h.Repository.CreateUser(strings.TrimSpace(request.Login), request.Password)
	if err != nil {
		logrus.Error(err)
		ctx.Status(http.StatusBadRequest)
		return
	}

	ctx.JSON(http.StatusCreated, schemes.NewUser(user))
}

func (h *Handler) LoginUser(ctx *gin.Context) {
	var request schemes.LoginUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}

	user, err := h.Repository.GetUserByLogin(strings.TrimSpace(request.Login))
	if err != nil || user.Password != request.Password {
		ctx.Status(http.StatusNotFound)
		return
	}

	ctx.JSON(http.StatusOK, schemes.NewUser(user))
}

func (h *Handler) LogoutUser(ctx *gin.Context) {
	ctx.Status(http.StatusOK)
}

func (h *Handler) newDrugResponse(drug ds.Drug) schemes.Drug {
	likesCount, err := h.Repository.GetDrugLikesCount(drug.ID)
	if err != nil {
		logrus.Error(err)
	}

	return schemes.NewDrug(drug, likesCount, singleton.GetCurrentUser().ID)
}

func (h *Handler) abortWithDrugError(ctx *gin.Context, err error) {
	if errors.Is(err, repository.ErrDrugNotFound) {
		ctx.Status(http.StatusNotFound)
		return
	}

	logrus.Error(err)
	ctx.Status(http.StatusInternalServerError)
}

func parseAdultDose(value string, fallback float64) float64 {
	dose, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}

	if dose < adultDoseFilterMinMg {
		return adultDoseFilterMinMg
	}

	return dose
}

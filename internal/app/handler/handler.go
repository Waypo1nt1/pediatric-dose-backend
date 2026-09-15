package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"pediatric-dose-backend/internal/app/ds"
	"pediatric-dose-backend/internal/app/repository"
)

const (
	adultDoseSliderMinMg = 0
	adultDoseSliderMaxMg = 1000
	adultDoseScaleMarks  = 4
	maxDrugDoseMg        = 100000
	maxDrugNameLength    = 100
	maxShortInfoLength   = 500
	currentCreatorID     = 1
)

type Handler struct {
	Repository *repository.Repository
}

type DrugCard struct {
	ds.Drug
	LikesCount int64
}

func NewHandler(r *repository.Repository) *Handler {
	return &Handler{
		Repository: r,
	}
}

func (h *Handler) GetDrugCatalog(ctx *gin.Context) {
	minAdultDoseValue := ctx.Query("min_adult_dose")
	maxAdultDoseValue := ctx.Query("max_adult_dose")

	minAdultDoseMg := parseAdultDose(minAdultDoseValue, adultDoseSliderMinMg)
	maxAdultDoseMg := parseAdultDose(maxAdultDoseValue, adultDoseSliderMaxMg)
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
	}

	drugCards := make([]DrugCard, 0, len(drugs))
	for _, drug := range drugs {
		likesCount, err := h.Repository.GetDrugLikesCount(drug.ID)
		if err != nil {
			logrus.Error(err)
		}

		drugCards = append(drugCards, DrugCard{
			Drug:       drug,
			LikesCount: likesCount,
		})
	}

	ctx.HTML(http.StatusOK, "drugs_catalog.html", gin.H{
		"DrugCards":          drugCards,
		"MinAdultDose":       minAdultDoseMg,
		"MaxAdultDose":       maxAdultDoseMg,
		"AdultDoseSliderMin": adultDoseSliderMinMg,
		"AdultDoseSliderMax": adultDoseSliderMaxMg,
		"AdultDoseScale":     adultDoseScale(),
		"ActiveTab":          "catalog",
	})
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
			logrus.Error(convertErr)
			ctx.HTML(http.StatusNotFound, "drug_feed.html", gin.H{
				"Drug":      ds.Drug{},
				"ActiveTab": "feed",
			})
			return
		}

		if ctx.Query("next") == "true" {
			drug, err = h.Repository.GetNextDrug(drugID)
		} else {
			drug, err = h.Repository.GetDrugByID(drugID)
		}
	}

	if err != nil {
		logrus.Error(err)

		status := http.StatusInternalServerError
		if errors.Is(err, repository.ErrDrugNotFound) {
			status = http.StatusNotFound
		}

		ctx.HTML(status, "drug_feed.html", gin.H{
			"Drug":      ds.Drug{},
			"ActiveTab": "feed",
		})
		return
	}

	likesCount, err := h.Repository.GetDrugLikesCount(drug.ID)
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "drug_feed.html", gin.H{
		"Drug":       drug,
		"LikesCount": likesCount,
		"ActiveTab":  "feed",
	})
}

func (h *Handler) GetDrugDraft(ctx *gin.Context) {
	h.renderDrugDraft(ctx, http.StatusOK, "")
}

func (h *Handler) CreateDrugDraft(ctx *gin.Context) {
	drugName := strings.TrimSpace(ctx.PostForm("drug_name"))
	if drugName == "" || utf8.RuneCountInString(drugName) > maxDrugNameLength {
		h.renderDrugDraft(ctx, http.StatusBadRequest, "Укажите наименование препарата до 100 символов")
		return
	}

	err := h.Repository.CreateDrugDraft(currentCreatorID, drugName)
	if err != nil {
		logrus.Error(err)
	}

	ctx.Redirect(http.StatusFound, "/drug-draft")
}

func (h *Handler) PublishDrugDraft(ctx *gin.Context) {
	shortInfo := strings.TrimSpace(ctx.PostForm("short_info"))
	recommendedAdultDoseMg, adultDoseErr := strconv.ParseFloat(ctx.PostForm("recommended_adult_dose_mg"), 64)
	maxDailyDoseMg, maxDailyDoseErr := strconv.ParseFloat(ctx.PostForm("max_daily_dose_mg"), 64)

	if shortInfo == "" || utf8.RuneCountInString(shortInfo) > maxShortInfoLength ||
		adultDoseErr != nil || maxDailyDoseErr != nil ||
		recommendedAdultDoseMg <= 0 || maxDailyDoseMg < recommendedAdultDoseMg || maxDailyDoseMg > maxDrugDoseMg {
		h.renderDrugDraft(ctx, http.StatusBadRequest, "Заполните краткое описание и обе дозы: максимум в сутки не меньше взрослой дозы")
		return
	}

	drugID, err := h.Repository.PublishDrugDraft(currentCreatorID, shortInfo, recommendedAdultDoseMg, maxDailyDoseMg)
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusFound, "/drug-draft")
		return
	}

	ctx.Redirect(http.StatusFound, "/drug-feed/"+strconv.FormatUint(uint64(drugID), 10))
}

func (h *Handler) DeleteDrug(ctx *gin.Context) {
	drugID, err := strconv.Atoi(ctx.PostForm("drug_id"))
	if err != nil {
		logrus.Error(err)
		ctx.Redirect(http.StatusFound, "/drugs")
		return
	}

	err = h.Repository.DeleteDrug(drugID)
	if err != nil {
		logrus.Error(err)
	}

	ctx.Redirect(http.StatusFound, "/drugs")
}

func (h *Handler) renderDrugDraft(ctx *gin.Context, status int, errorMessage string) {
	drug, err := h.Repository.GetDraftDrug(currentCreatorID)
	if err != nil && !errors.Is(err, repository.ErrDrugNotFound) {
		logrus.Error(err)
	}

	ctx.HTML(status, "drug_draft.html", gin.H{
		"Drug":         drug,
		"ErrorMessage": errorMessage,
		"ActiveTab":    "draft",
	})
}

func parseAdultDose(value string, fallback float64) float64 {
	dose, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}

	if dose < adultDoseSliderMinMg {
		return adultDoseSliderMinMg
	}

	if dose > adultDoseSliderMaxMg {
		return adultDoseSliderMaxMg
	}

	return dose
}

func adultDoseScale() []int {
	step := (adultDoseSliderMaxMg - adultDoseSliderMinMg) / adultDoseScaleMarks

	scale := make([]int, 0, adultDoseScaleMarks+1)
	for mark := adultDoseSliderMinMg; mark <= adultDoseSliderMaxMg; mark += step {
		scale = append(scale, mark)
	}

	return scale
}

package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"pediatric-dose-backend/internal/app/repository"
)

const (
	adultDoseSliderMinMg = 0
	adultDoseSliderMaxMg = 1000
	adultDoseScaleMarks  = 4
)

type Handler struct {
	Repository   *repository.Repository
	MediaBaseURL string
}

type DrugCard struct {
	repository.Drug
	LikesCount int
}

func NewHandler(r *repository.Repository, mediaBaseURL string) *Handler {
	return &Handler{
		Repository:   r,
		MediaBaseURL: mediaBaseURL,
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

	var drugs []repository.Drug
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
		drugCards = append(drugCards, DrugCard{
			Drug:       drug,
			LikesCount: len(drug.LikedByUserIDs),
		})
	}

	ctx.HTML(http.StatusOK, "drugs_catalog.html", gin.H{
		"DrugCards":          drugCards,
		"MinAdultDose":       minAdultDoseMg,
		"MaxAdultDose":       maxAdultDoseMg,
		"AdultDoseSliderMin": adultDoseSliderMinMg,
		"AdultDoseSliderMax": adultDoseSliderMaxMg,
		"AdultDoseScale":     adultDoseScale(),
		"MediaBaseURL":       h.MediaBaseURL,
		"ActiveTab":          "catalog",
	})
}

func (h *Handler) GetDrugFeed(ctx *gin.Context) {
	drugIDValue := strings.Trim(ctx.Param("drug_id"), "/")

	var drug repository.Drug
	var err error

	if drugIDValue == "" {
		drug, err = h.Repository.GetFirstPublishedDrug()
	} else {
		drugID, convertErr := strconv.Atoi(drugIDValue)
		if convertErr != nil {
			logrus.Error(convertErr)
		}

		if ctx.Query("next") == "true" {
			drug, err = h.Repository.GetNextDrug(drugID)
		} else {
			drug, err = h.Repository.GetDrugByID(drugID)
		}
	}

	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "drug_feed.html", gin.H{
		"Drug":         drug,
		"LikesCount":   len(drug.LikedByUserIDs),
		"MediaBaseURL": h.MediaBaseURL,
		"ActiveTab":    "feed",
	})
}

func (h *Handler) GetDrugDraft(ctx *gin.Context) {
	drug, err := h.Repository.GetDraftDrug()
	if err != nil {
		logrus.Error(err)
	}

	ctx.HTML(http.StatusOK, "drug_draft.html", gin.H{
		"Drug":         drug,
		"MediaBaseURL": h.MediaBaseURL,
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

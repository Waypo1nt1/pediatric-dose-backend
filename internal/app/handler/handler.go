package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"pediatric-dose-backend/internal/app/repository"
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
	maxAdultDoseValue := ctx.Query("max_adult_dose")

	maxAdultDoseMg, err := strconv.ParseFloat(maxAdultDoseValue, 64)
	if err != nil {
		maxAdultDoseMg = 0
	}

	drugs, err := h.Repository.GetPublishedDrugs(maxAdultDoseMg)
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
		"DrugCards":    drugCards,
		"MaxAdultDose": maxAdultDoseValue,
		"MediaBaseURL": h.MediaBaseURL,
		"ActiveTab":    "catalog",
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

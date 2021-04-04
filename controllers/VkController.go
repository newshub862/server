package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"newshub-server/models"
	"newshub-server/services"
)

type VkController struct {
	service *services.VkService
	config  *models.Config
}

func NewVkCtrl(cfg *models.Config) *VkController {
	ctrl := new(VkController)
	ctrl.config = cfg
	ctrl.service = services.NewVkService(cfg)

	return ctrl
}

func (ctrl *VkController) GetPageData(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	pageData := models.VkPageData{
		News:   ctrl.service.GetNews(claims.Id, 1, 0),
		Groups: ctrl.service.GetAllGroups(claims.Id),
	}

	json.NewEncoder(w).Encode(pageData)
}

func (ctrl *VkController) GetNews(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	page, err := strconv.Atoi(r.URL.Query().Get("page"))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	groupID := int64(0)

	if r.FormValue("group_id") != "" {
		groupID, err = strconv.ParseInt(r.FormValue("group_id"), 10, 64)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	news := ctrl.service.GetNews(claims.Id, page, groupID)

	if err := json.NewEncoder(w).Encode(news); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (ctrl *VkController) Search(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	groupID := int64(0)
	var err error

	if r.FormValue("group_id") != "" {
		groupID, err = strconv.ParseInt(r.FormValue("group_id"), 10, 64)

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	}

	news := ctrl.service.Search(r.FormValue("q"), groupID, claims.Id)

	if err := json.NewEncoder(w).Encode(news); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

package controllers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"newshub-server/models"
	"newshub-server/services"
	"strconv"

	"github.com/gorilla/mux"
)

// RssController - request handlers
type RssController struct {
	service *services.RssService
	config  *models.Config
}

// NewRssCtrl - init service
func NewRssCtrl(cfg *models.Config) *RssController {
	ctrl := new(RssController)
	ctrl.config = cfg
	ctrl.service = services.NewRssService(cfg)

	return ctrl
}

// GetAll - get feed list
func (ctrl *RssController) GetAll(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	feeds := ctrl.service.GetRss(claims.Id)

	json.NewEncoder(w).Encode(feeds)
}

// GetArticles - get articles for feed
func (ctrl *RssController) GetArticles(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["feed_id"], 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	page, err := strconv.Atoi(r.FormValue("page"))

	if err != nil {
		log.Println("page is invalid:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := getClaims(r)
	feed := ctrl.service.GetArticles(id, claims.Id, page)

	json.NewEncoder(w).Encode(feed)
}

// GetArticle - get one article
func (ctrl *RssController) GetArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	feedID, err := strconv.ParseInt(vars["feed_id"], 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := getClaims(r)
	article := ctrl.service.GetArticle(id, feedID, claims.Id)

	if article == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	go func() {
		if article.IsRead {
			return
		}
		data := models.ArticlesUpdateData{
			ArticleId:  article.Id,
			IsRead:     true,
			IsBookmark: article.IsBookmark,
		}
		ctrl.service.ArticleUpdate(claims.Id, data)
	}()

	json.NewEncoder(w).Encode(article)
}

// AddFeed - add feed
func (ctrl *RssController) AddFeed(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	filters := models.RssFilters{}

	if err := json.NewDecoder(r.Body).Decode(&filters); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if filters.Url == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctrl.service.AddFeed(filters.Url, claims.Id)
	ctrl.GetAll(w, r)
}

// Delete - delete feed
func (ctrl *RssController) Delete(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctrl.service.Delete(id, claims.Id)
	ctrl.GetAll(w, r)
}

// UploadOpml - upload, parse OPML and update feeds
func (ctrl *RssController) UploadOpml(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	file, _, err := r.FormFile("file")

	if err != nil || file == nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := ioutil.ReadAll(file)

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ctrl.service.Import(data, claims.Id)
	ctrl.GetAll(w, r)
}

// CreateOpml - create OPML file
func (ctrl *RssController) CreateOpml(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	opmlBytes := ctrl.service.Export(claims.Id)

	w.Write(opmlBytes)
}

// SetNewFeedName - set new feed name
func (ctrl *RssController) SetNewFeedName(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	feedID, err := strconv.ParseInt(vars["id"], 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	feeds := models.FeedUpdateData{}

	if err := json.NewDecoder(r.Body).Decode(&feeds); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	feeds.FeedId = feedID
	claims := getClaims(r)
	feed := ctrl.service.SetNewName(feeds, claims.Id)
	jsonData, _ := json.Marshal(feed)

	w.Write(jsonData)
}

// GetBookmarks - get bookmark list
func (ctrl *RssController) GetBookmarks(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	page, err := strconv.Atoi(r.FormValue("page"))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	articles := ctrl.service.GetBookmarks(page, claims.Id)

	json.NewEncoder(w).Encode(articles)
}

// Search - search by articles
func (ctrl *RssController) Search(w http.ResponseWriter, r *http.Request) {
	searchString := r.FormValue("search_string")
	isBookmark, _ := strconv.ParseBool(r.FormValue("is_bookmark"))
	feedID, err := strconv.ParseInt(r.FormValue("feed_id"), 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := getClaims(r)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ctrl.service.Search(searchString, isBookmark, feedID, claims.Id))
}

// UpdateArticle - update by id
func (ctrl *RssController) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data := models.ArticlesUpdateData{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if id != data.ArticleId {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	claims := getClaims(r)
	article := ctrl.service.ArticleUpdate(claims.Id, data)

	if err := json.NewEncoder(w).Encode(article); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

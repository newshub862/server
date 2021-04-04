package services

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net/http"
	"sync"

	"newshub-server/models"

	"golang.org/x/net/html/charset"
	"gorm.io/gorm"
)

// RssService - service
type RssService struct {
	db         *gorm.DB
	config     *models.Config
	UnreadOnly bool
}

func NewRssService(config *models.Config) *RssService {
	return &RssService{
		db:     getDb(),
		config: config,
	}
}

func (service *RssService) SetDb(db *gorm.DB) {
	service.db = db
}

func (service *RssService) SetConfig(cfg *models.Config) {
	service.config = cfg
}

// GetRss - get all rss
func (service *RssService) GetRss(id int64) []models.Feed {
	var rss []models.Feeds
	service.db.
		Preload("Articles", "IsRead=?", "0").
		Where(&models.Feeds{UserId: id}).
		Find(&rss)
	feeds := make([]models.Feed, len(rss))
	var wg sync.WaitGroup

	for i, item := range rss {
		wg.Add(1)
		go func(item models.Feeds, i int) {
			count := len(item.Articles)
			item.Articles = nil
			feeds[i] = models.Feed{Feed: item, ArticlesCount: count, ExistUnread: count > 0}

			wg.Done()
		}(item, i)
	}

	wg.Wait()

	return feeds
}

// GetArticles - get articles for rss by id
func (service *RssService) GetArticles(id int64, userID int64, page int) *models.ArticlesJSON {
	var articles []models.Articles
	var count int64
	offset := service.config.PageSize * (page - 1)
	whereObject := models.Articles{FeedId: id}

	query := service.db.Where(&whereObject).
		Select("Id, Title, IsBookmark, IsRead, Link, FeedId").
		Limit(service.config.PageSize).
		Offset(offset).
		Order("Id desc")
	queryCount := service.db.Model(&whereObject).Where(&whereObject)

	var settings models.Settings
	service.db.Where(models.Settings{UserId: userID}).Find(&settings)

	if settings.UnreadOnly {
		whereNotObject := models.Articles{IsRead: true}
		query = query.Not(&whereNotObject)
		queryCount = queryCount.Not(&whereNotObject)
	}

	query.Find(&articles)
	queryCount.Count(&count)

	return &models.ArticlesJSON{Articles: articles, Count: count}
}

// GetArticle - get one article
func (service *RssService) GetArticle(id int64, feedID int64, userID int64) *models.Articles {
	rss := service.GetRss(userID)

	if len(rss) == 0 {
		return nil
	}

	// get article
	var article models.Articles
	service.db.Where(&models.Articles{Id: id, FeedId: feedID}).First(&article)

	var settings models.Settings // todo: to func
	service.db.Where(models.Settings{UserId: userID}).Find(&settings)

	// update state
	article.IsRead = true
	service.db.Save(&article)

	if settings.MarkSameRead {
		go service.markSameArticles(article.Link, article.FeedId)
	}

	return &article
}

// Import - import OPML file
func (service *RssService) Import(data []byte, userID int64) {
	// parse opml
	var opml models.OPML
	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.CharsetReader = charset.NewReaderLabel
	err := decoder.Decode(&opml)

	if err != nil {
		log.Println("OPML import error: ", err.Error())
		return
	}

	dbExec(func(db *gorm.DB) {
		for _, outline := range opml.Outlines {
			feed := models.Feeds{
				Name:   outline.Title,
				Url:    outline.URL,
				UserId: userID,
			}
			db.Save(&feed)
		}
	})
}

// Export - export feeds to OPML file
func (service *RssService) Export(userID int64) []byte {
	// get data from DB
	var rss []models.Feeds
	service.db.Where(&models.Feeds{UserId: userID}).Find(&rss)
	opml := models.OPML{
		HeadText: "Feeds",
		Version:  1.1,
		Outlines: make([]models.OPMLOutline, 0, len(rss)),
	}

	// create array of structures
	for _, feed := range rss {
		outline := models.OPMLOutline{
			Title: feed.Name,
			URL:   feed.Url,
			Text:  feed.Name,
		}
		opml.Outlines = append(opml.Outlines, outline)
	}

	// create OPML file bytes
	xmlString, _ := xml.Marshal(opml)
	fmt.Println(string(xmlString))

	return xmlString
}

// AddFeed - add new feed
func (service *RssService) AddFeed(url string, userID int64) {
	// get rss xml
	response, err := http.Get(url)

	if err != nil {
		log.Println("Get XML error: ", err.Error())
		return
	}

	defer response.Body.Close()

	// parse feed xml and create structure
	var xmlModel models.XMLFeed
	decoder := xml.NewDecoder(response.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	err = decoder.Decode(&xmlModel)

	if err != nil {
		log.Println("XML unmarshall error on URL: ", url, err.Error())
		return
	}

	// insert in DB
	err = service.db.Create(&models.Feeds{Url: url, UserId: userID, Name: xmlModel.RssName}).Error
	// todo: send message for update

	if err != nil {
		log.Println("insert error", err.Error())
	}
}

// Delete - remove feed
func (service *RssService) Delete(id int64, userID int64) {
	feed := service.GetRss(userID)

	if len(feed) == 0 {
		return
	}

	service.db.Where(models.Articles{FeedId: id}).Delete(models.Articles{})
	service.db.Delete(models.Feeds{Id: id})
}

// SetNewName - update feed name
func (service *RssService) SetNewName(data models.FeedUpdateData, userID int64) models.Feeds {
	feed := models.Feeds{}
	service.db.Where(&models.Feeds{Id: data.FeedId, UserId: userID}).First(&feed)

	if feed.Id == 0 {
		return feed
	}

	if data.IsReadAll {
		service.db.Model(&models.Articles{}).
			Where(&models.Articles{FeedId: feed.Id}).
			Not(&models.Articles{IsRead: true}).
			UpdateColumn("is_read = ?", true)
	}
	if data.Name != "" {
		feed.Name = data.Name
		service.db.Save(&feed)
	}

	return feed
}

// GetBookmarks - get all bookmarks
func (service *RssService) GetBookmarks(page int, userID int64) *models.ArticlesJSON {
	var articles []models.Articles
	whereCond := "articles.IsBookmark = true and feeds.UserId = ?"
	offset := service.config.PageSize * (page - 1)
	var count int64

	service.db.Where(whereCond, userID).
		Joins("join feeds on articles.FeedId = feeds.Id").
		Select("Id, Title, IsBookmark, IsRead").
		Limit(service.config.PageSize).
		Offset(offset).
		Order("Id desc").
		Find(&articles)
	service.db.Model(&models.Articles{}).Where(whereCond, userID).
		Joins("join feeds on articles.FeedId = feeds.Id").Count(&count)

	return &models.ArticlesJSON{Articles: articles, Count: count}
}

// Search - search articles by title or body
func (service *RssService) Search(searchString string, isBookmark bool, feedID int64, userID int64) *models.ArticlesJSON {
	var articles []models.Articles
	query := service.db.
		Joins("join feeds on articles.FeedId = feeds.Id").
		Select("articles.Id, articles.Title, articles.IsBookmark, articles.IsRead, articles.Link").
		Where("(articles.Title LIKE ? OR articles.Body LIKE ?) and feeds.UserId = ?", "%"+searchString+"%", "%"+searchString+"%", userID)

	if feedID != 0 {
		query = query.Where(&models.Articles{Id: feedID})
	}
	if isBookmark {
		query = query.Where("articles.IsBookmark = 1")
	}

	query.Find(&articles)

	return &models.ArticlesJSON{Articles: articles}
}

func (service *RssService) ArticleUpdate(userID int64, data models.ArticlesUpdateData) models.Articles {
	service.db = service.db.Debug()
	whereCond := "articles.Id = ? and feeds.UserId = ?"
	article := models.Articles{}
	err := service.db.
		Joins("join feeds on articles.FeedId = feeds.Id").
		Where(whereCond, data.ArticleId, userID).
		First(&article).
		Error
	if err != nil {
		log.Println("get article for update error:", err)
		return article
	}

	article.IsBookmark = data.IsBookmark
	article.IsRead = data.IsRead

	if err := service.db.Save(&article).Error; err != nil {
		log.Println("update article error:", err)
	}

	return article
}

func (service *RssService) markSameArticles(url string, feedID int64) {
	// service.db.Model(&models.Articles{}).Where("Link = ? and FeedId != ?").
	// 	Not(&models.Articles{Id: feedID}).
	// 	UpdateColumn("IsRead = ?", true)
}

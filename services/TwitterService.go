package services

import (
	"log"
	"strconv"

	"newshub-server/models"

	"gorm.io/gorm"
)

type TwitterService struct {
	db     *gorm.DB
	config *models.Config
}

func (service *TwitterService) SetConfig(config *models.Config) {
	service.config = config
}

func (service *TwitterService) SetDb(db *gorm.DB) {
	service.db = db
}

// Init - create new struct pointer with collection
func NewTwitterService(config *models.Config) *TwitterService {
	return &TwitterService{db: getDb(), config: config}
}

func (service *TwitterService) GetNews(id int64, page int, sourceID int64) []models.TwitterNewsView {
	var dbModels []models.TwitterNews
	cond := models.TwitterNews{UserId: id}
	offset := service.config.PageSize * (page - 1)

	if sourceID != 0 {
		cond.SourceId = sourceID
	}

	err := service.db.Where(&cond).
		Limit(service.config.PageSize).
		Offset(offset).
		Order("Id desc").
		Find(&dbModels).
		Error
	if err != nil {
		log.Printf("get twitter news for %d error: %s", id, err)
	}

	return getNewsView(dbModels)
}

func (service *TwitterService) GetAllSources(id int64) []models.TwitterSource {
	var result []models.TwitterSource

	err := service.db.Where(&models.TwitterSource{UserId: id}).Find(&result).Error
	if err != nil {
		log.Printf("get twitter news for %d error: %s", id, err)
	}

	return result
}

func (service *TwitterService) Search(searchString string, sourceID int64, userID int64) []models.TwitterNewsView {
	var dbModels []models.TwitterNews
	query := service.db.Where("Text LIKE ? and UserId = ?", "%"+searchString+"%", userID)

	if sourceID != 0 {
		query = query.Where(&models.TwitterNews{SourceId: sourceID})
	}

	query.Order("Id desc").Find(&dbModels)

	return getNewsView(dbModels)
}

func getNewsView(dbModels []models.TwitterNews) []models.TwitterNewsView {
	result := make([]models.TwitterNewsView, len(dbModels))

	for index, item := range dbModels {
		result[index] = models.TwitterNewsView{
			SourceId:    item.SourceId,
			ExpandedUrl: item.ExpandedUrl,
			Image:       item.Image,
			Text:        item.Text,
			Id:          strconv.FormatInt(item.Id, 10), // string for js
			TweetId:     strconv.FormatInt(item.TweetId, 10),
		}
	}

	return result
}

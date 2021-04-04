package services

import (
	"newshub-server/models"

	"gorm.io/gorm"
)

// VkService - service
type VkService struct {
	db     *gorm.DB
	config *models.Config
}

// Init - create new struct pointer with collection
func NewVkService(config *models.Config) *VkService {
	return &VkService{db: getDb(), config: config}
}

func (service *VkService) SetDb(db *gorm.DB) {
	service.db = db
}

func (service *VkService) SetConfig(cfg *models.Config) {
	service.config = cfg
}

func (service *VkService) GetNews(id int64, page int, groupID int64) []models.VkNews {
	var result []models.VkNews
	conditions := models.VkNews{
		UserId: id,
	}
	offset := service.config.PageSize * (page - 1)

	if groupID != 0 {
		conditions.GroupId = groupID
	}

	service.db.Where(&conditions).
		Limit(service.config.PageSize).
		Offset(offset).
		Order("Id desc").
		Find(&result)

	return result
}

func (service *VkService) GetAllGroups(id int64) []models.VkGroup {
	var result []models.VkGroup

	service.db.Where(&models.VkGroup{UserId: id}).Find(&result)

	return result
}

func (service *VkService) Search(searchString string, groupID int64, userID int64) []models.VkNews {
	var result []models.VkNews
	query := service.db.Where("Text LIKE ? and userId = ?", "%"+searchString+"%", userID)

	if groupID != 0 {
		query = query.Where(&models.VkNews{GroupId: groupID})
	}

	query.Order("Id desc").Find(&result)

	return result
}

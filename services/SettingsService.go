package services

import (
	"errors"
	"log"

	"newshub-server/models"

	"gorm.io/gorm"
)

type SettingsService struct {
	db *gorm.DB
}

// Init - create new struct pointer with collection
func (service SettingsService) Init(config *models.Config) *SettingsService {
	return &SettingsService{db: getDb()}
}

func (service *SettingsService) SetDb(db *gorm.DB) {
	service.db = db
}

func (service *SettingsService) Create(userId int64) {
	settings := models.Settings{UserId: userId}
	service.db.Create(&settings)
}

func (service *SettingsService) Update(settings models.Settings) (models.Settings, error) {
	err := service.db.
		Where(&models.Settings{UserId: settings.UserId}).
		Delete(&models.Settings{}).
		Error
	if err != nil {
		return models.Settings{}, err
	}

	if err := service.db.Save(&settings).Error; err != nil {
		log.Println("save settings error:", err)
		return models.Settings{}, errors.New("save settings error")
	}

	return settings, nil
}

func (service *SettingsService) Get(userId int64) models.Settings {
	settings := models.Settings{UserId: userId}
	service.db.Find(&settings)

	return settings
}

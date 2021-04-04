package services

import (
	"fmt"
	"newshub-server/models"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB
var cfg *models.Config

func Setup(config *models.Config) {
	cfg = config
	// db = getDb()

}

func dbExec(closure func(db *gorm.DB)) {
	if db == nil {
		return
	}

	closure(db)
}

func getDb() *gorm.DB {
	if db != nil {
		return db
	}

	defer migrate()

	if cfg.Driver == "sqlite3" {
		sqliteDB, err := gorm.Open(sqlite.Open(cfg.ConnectionString), &gorm.Config{})
		if err != nil {
			panic("open db error: " + err.Error())
		}

		db = sqliteDB
		return db
	}

	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.DbHost, cfg.DbUser, cfg.DbPassword, cfg.DbName, cfg.DbPort,
	)

	pgdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("open db error: " + err.Error())
	}

	sqlDB, err := pgdb.DB()
	if err != nil {
		panic("open db error: " + err.Error())
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	db = pgdb

	return db
}

func migrate() {
	db := getDb()
	db.AutoMigrate(&models.Users{})
	db.AutoMigrate(&models.Feeds{})
	db.AutoMigrate(&models.Articles{})
	db.AutoMigrate(&models.Settings{})
	db.AutoMigrate(&models.VkNews{})
	db.AutoMigrate(&models.VkGroup{})
	db.AutoMigrate(&models.TwitterNews{})
	db.AutoMigrate(&models.TwitterSource{})
}

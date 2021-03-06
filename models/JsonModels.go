package models

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/dgrijalva/jwt-go"
)

// Config - app config, create from config file
type Config struct {
	Address          string `json:"address"`
	Driver           string `json:"driver"`
	ConnectionString string `json:"connection_string"`
	DbHost           string `json:"db_host"`
	DbName           string `json:"db_name"`
	DbUser           string `json:"db_user"`
	DbPassword       string `json:"db_password"`
	DbPort           int    `json:"db_port"`
	JwtSign          string `json:"jwt_sign"`
	PageSize         int    `json:"page_size"`
}

// NewConfig return new config struct pointer
func NewConfig(path string) *Config {
	fromEnv := os.Getenv("FROM_ENV") == "true"
	cfg := new(Config)
	var jsonBytes []byte

	if fromEnv {
		jsonBytes = []byte(os.Getenv("CFG"))
	} else {
		var err error

		jsonBytes, err = ioutil.ReadFile(path)
		if err != nil {
			panic("Read config file error")
		}
	}

	// set default values
	cfg.PageSize = 20

	if err := json.Unmarshal(jsonBytes, cfg); err != nil {
		panic(err.Error())
	}

	return cfg
}

type SettingsData struct {
	UserId               int64  `json:"UserId"`
	VkLogin              string `json:"VkLogin"`
	VkPassword           string `json:"VkPassword"`
	TwitterName          string `json:"TwitterName"`
	VkNewsEnabled        bool   `json:"VkNewsEnabled"`
	TwitterEnabled       bool   `json:"TwitterEnabled"`
	TwitterSimpleVersion bool   `json:"TwitterSimpleVersion"`
	MarkSameRead         bool   `json:"MarkSameRead"`
	UnreadOnly           bool   `json:"UnreadOnly"`
	RssEnabled           bool   `json:"RssEnabled"`
	ShowPreviewButton    bool   `json:"ShowPreviewButton"`
	ShowTabButton        bool   `json:"ShowTabButton"`
	ShowReadButton       bool   `json:"ShowReadButton"`
	ShowLinkButton       bool   `json:"ShowLinkButton"`
	ShowBookmarkButton   bool   `json:"ShowBookmarkButton"`
}

type VkData struct {
	GroupId      int64  `json:"GroupId"`
	SearchString string `json:"SearchString"`
}

type TwitterData struct {
	SourceId     int64  `json:"SourceId"`
	SearchString string `json:"SearchString"`
}

type RssFilters struct {
	Search      string `json:"search"`
	MarkAllRead bool   `json:"mark_all_read"`
	Name        string `json:"name"`
	Url         string `json:"url"`
}

type AuthData struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

type JwtClaims struct {
	*jwt.MapClaims
	Id  int64
	Exp int64
}

func (JwtClaims) Valid() error {
	return nil
}

type ArticlesUpdateData struct {
	ArticleId  int64 `json:"article_id"`
	IsRead     bool  `json:"is_read"`
	IsBookmark bool  `json:"is_bookmark"`
}

type FeedUpdateData struct {
	FeedId    int64  `json:"feed_id"`
	Name      string `json:"name"`
	IsReadAll bool   `json:"is_read_all"`
}

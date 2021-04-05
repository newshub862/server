package main

import (
	"flag"
	"log"
	"net/http"

	"newshub-server/controllers"
	"newshub-server/middleware"
	"newshub-server/models"
	"newshub-server/services"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var conf *models.Config

const defaultConfigPath = "./cfg.json"

func init() {
	// read config file
	pathPtr := flag.String("config", defaultConfigPath, "Path for configuration file")
	flag.Parse()

	conf = models.NewConfig(*pathPtr)
}

func createRouter() http.Handler {
	rssCtrl := controllers.NewRssCtrl(conf)
	userCtrl := controllers.NewUserCtrl(conf)
	vkCtrl := controllers.NewVkCtrl(conf)
	twitterCtrl := controllers.NewTwitterCtrl(conf)
	router := mux.NewRouter()
	router.StrictSlash(true)

	// rss
	router.HandleFunc("/rss", rssCtrl.GetAll).Methods(http.MethodGet)
	router.HandleFunc("/rss", rssCtrl.AddFeed).Methods(http.MethodPost)
	router.HandleFunc("/rss/{id}", rssCtrl.Delete).Methods(http.MethodDelete)
	router.HandleFunc("/rss/{id}", rssCtrl.SetNewFeedName).Methods(http.MethodPut)
	router.HandleFunc("/rss/search", rssCtrl.Search).Methods(http.MethodGet)
	router.HandleFunc("/rss/opml", rssCtrl.UploadOpml).Methods(http.MethodPost)
	router.HandleFunc("/rss/opml", rssCtrl.CreateOpml).Methods(http.MethodGet)

	// articles
	router.HandleFunc("/rss/{feed_id}/articles", rssCtrl.GetArticles).Methods(http.MethodGet)
	router.HandleFunc("/rss/{feed_id}/articles/{id}", rssCtrl.GetArticle).Methods(http.MethodGet)
	router.HandleFunc("/rss/{feed_id}/articles/{id}", rssCtrl.UpdateArticle).Methods(http.MethodPut)
	router.HandleFunc("/rss/articles/bookmarks", rssCtrl.GetBookmarks)

	// user
	router.HandleFunc("/auth", userCtrl.Auth).Methods(http.MethodPost)
	router.HandleFunc("/registration", userCtrl.Registration).Methods(http.MethodPost)
	router.HandleFunc("/users/settings", userCtrl.GetUserSettings).Methods(http.MethodGet)
	router.HandleFunc("/users/settings", userCtrl.SaveSettings).Methods(http.MethodPut)
	router.HandleFunc("/users/refresh", userCtrl.RefreshToken).Methods(http.MethodPut)

	// vk
	router.HandleFunc("/vk", vkCtrl.GetPageData)
	router.HandleFunc("/vk/news", vkCtrl.GetNews)
	router.HandleFunc("/vk/search", vkCtrl.Search)

	// twitter
	router.HandleFunc("/twitter", twitterCtrl.GetPageData).Methods(http.MethodGet)
	router.HandleFunc("/twitter/news", twitterCtrl.GetNews).Methods(http.MethodGet)
	router.HandleFunc("/twitter/sources", twitterCtrl.GetSources).Methods(http.MethodGet)
	router.HandleFunc("/twitter/search", twitterCtrl.Search).Methods(http.MethodGet)

	// middleware
	amw := middleware.AuthenticationMiddleware{}
	amw.Populate(conf)

	router.Use(amw.Middleware)

	return router
}

func main() {
	// todo: websocket for update feed list
	services.Setup(conf)
	controllers.Config = conf

	router := createRouter()
	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "HEAD", "OPTIONS", "DELETE"})

	log.Println("server start on", conf.Address)

	if err := http.ListenAndServe(conf.Address, handlers.CORS(originsOk, headersOk, methodsOk)(router)); err != nil {
		panic("start server error: " + err.Error())
	}
}

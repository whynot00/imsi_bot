package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/whynot00/imsi-bot/internal/config"
	"github.com/whynot00/imsi-bot/internal/db"
	"github.com/whynot00/imsi-bot/internal/handler"
	"github.com/whynot00/imsi-bot/internal/handler/middleware"
	"github.com/whynot00/imsi-bot/internal/repo"
	"github.com/whynot00/imsi-bot/internal/service"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file, reading from environment")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	database, err := db.Connect(cfg.Postgres)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer database.Close()
	fmt.Fprintln(os.Stderr, "connected to postgres")

	// --- repos ---
	deviceRepo := repo.NewDeviceRepo(database)
	parametrRepo := repo.NewParametrRepo(database)
	rkRepo := repo.NewRKRepo(database)
	searchRepo := repo.NewSearchRepo(database)
	userRepo := repo.NewUserRepo(database)

	// --- services ---
	importSvc := service.NewImportService(deviceRepo, parametrRepo, rkRepo)
	searchSvc := service.NewSearchService(searchRepo, deviceRepo)

	// --- handlers ---
	searchH := handler.NewSearchHandler(searchSvc)
	uploadH := handler.NewUploadHandler(importSvc)

	// --- router ---
	r := gin.Default()

	// статика
	r.StaticFile("/", "./static/index.html")
	r.Static("/static", "./static")

	// api
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	auth := middleware.TelegramAuth(botToken, userRepo)

	api := r.Group("/api", auth)
	{
		api.GET("/search", searchH.Search)
		api.GET("/search/suggest", searchH.Suggest)
		api.POST("/upload/parametr", uploadH.UploadParametr)
		api.POST("/upload/rk", uploadH.UploadRK)
	}

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

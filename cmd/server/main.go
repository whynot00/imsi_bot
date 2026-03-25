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
	internalUploadH := handler.NewInternalUploadHandler(importSvc)

	// --- router ---
	r := gin.Default()
	r.MaxMultipartMemory = 512 << 20 // 512 MB

	// статика
	r.StaticFile("/", "./static/index.html")
	r.StaticFile("/upload", "./static/upload.html")
	r.Static("/static", "./static")

	// Telegram Mini App API
	botToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	if botToken == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}
	tgAuth := middleware.TelegramAuth(botToken, userRepo)

	api := r.Group("/api", tgAuth)
	{
		api.GET("/search", searchH.Search)
		api.GET("/search/suggest", searchH.Suggest)
		api.POST("/upload/parametr", uploadH.UploadParametr)
		api.POST("/upload/rk", uploadH.UploadRK)
	}

	// Internal API — api key auth, для загрузки больших файлов
	internal := r.Group("/internal", middleware.ApiKeyAuth())
	{
		internal.POST("/upload/parametr", internalUploadH.UploadParametr)
		internal.POST("/upload/rk", internalUploadH.UploadRK)
	}

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

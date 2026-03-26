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
	ipRepo := repo.NewIPRepo(database)

	// --- services ---
	importSvc := service.NewImportService(database, deviceRepo, parametrRepo, rkRepo)
	searchSvc := service.NewSearchService(searchRepo, deviceRepo)

	// --- job store (shared between upload handlers) ---
	jobs := handler.NewJobStore()

	// --- handlers ---
	searchH := handler.NewSearchHandler(searchSvc)
	uploadH := handler.NewUploadHandler(importSvc, jobs)
	internalUploadH := handler.NewInternalUploadHandler(importSvc, jobs)
	userH := handler.NewUserHandler(userRepo)
	ipH := handler.NewIPHandler(ipRepo)

	// --- log store (last 500 entries) ---
	logStore := middleware.NewLogStore(500)

	// --- handlers ---
	logsH := handler.NewLogsHandler(logStore)

	// --- router ---
	r := gin.Default()
	r.MaxMultipartMemory = 512 << 20 // 512 MB

	// IP whitelist + request logger на весь сервер
	r.Use(middleware.IPWhitelist(ipRepo))
	r.Use(middleware.RequestLogger(logStore))

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
		api.GET("/upload/status/:id", uploadH.JobStatus)
	}

	// Internal API — api key auth (IP whitelist already global)
	internal := r.Group("/internal", middleware.ApiKeyAuth())
	{
		internal.POST("/upload/parametr", internalUploadH.UploadParametr)
		internal.POST("/upload/rk", internalUploadH.UploadRK)
		internal.GET("/upload/status/:id", internalUploadH.JobStatus)

		internal.GET("/users", userH.List)
		internal.POST("/users", userH.Create)
		internal.DELETE("/users/:id", userH.Delete)

		internal.GET("/ips", ipH.List)
		internal.POST("/ips", ipH.Create)
		internal.DELETE("/ips/:id", ipH.Delete)

		internal.GET("/logs", logsH.All)
		internal.GET("/logs/errors", logsH.Errors)
	}

	addr := ":" + cfg.Port
	log.Printf("listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

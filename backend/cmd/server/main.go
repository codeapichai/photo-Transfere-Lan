package main

import (
	"log"
	"os"
	"path/filepath"

	"phototransferlan/backend/internal/application/service"
	"phototransferlan/backend/internal/infrastructure/database"
	infrarepo "phototransferlan/backend/internal/infrastructure/repository"
	"phototransferlan/backend/internal/infrastructure/storage"
	httpapi "phototransferlan/backend/internal/presentation/http"
	wsapi "phototransferlan/backend/internal/presentation/websocket"
)

func main() {
	dataDir := filepath.Join(os.Getenv("APPDATA"), "PhotoTransferLAN")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatal(err)
	}

	db, err := database.Open(filepath.Join(dataDir, "photo_transfer.db"))
	if err != nil {
		log.Fatal(err)
	}

	uploadsDir := filepath.Join(os.Getenv("USERPROFILE"), "Pictures", "PhotoTransferLAN")
	store := storage.NewLocalStore(uploadsDir)
	hub := wsapi.NewHub()

	userRepo := infrarepo.NewGormUserRepository(db)
	uploadRepo := infrarepo.NewGormUploadRepository(db)
	settingsRepo := infrarepo.NewGormSettingsRepository(db)
	sessionRepo := infrarepo.NewGormSessionRepository(db)
	tokenRepo := infrarepo.NewGormTemporaryTokenRepository(db)
	logRepo := infrarepo.NewGormActivityLogRepository(db)

	authSvc := service.NewAuthService(userRepo)
	securitySvc := service.NewSecurityService(sessionRepo, tokenRepo, settingsRepo, logRepo)
	settingsSvc := service.NewSettingsService(settingsRepo, logRepo)
	logSvc := service.NewLogService(logRepo)
	uploadSvc := service.NewUploadService(uploadRepo, settingsRepo, logRepo, store, hub)
	dashboardSvc := service.NewDashboardService(uploadRepo, settingsRepo)

	app := httpapi.NewServer(httpapi.Dependencies{
		Auth:      authSvc,
		Security:  securitySvc,
		Settings:  settingsSvc,
		Logs:      logSvc,
		Uploads:   uploadSvc,
		Dashboard: dashboardSvc,
		Hub:       hub,
	})

	cert := os.Getenv("PT_HTTPS_CERT")
	key := os.Getenv("PT_HTTPS_KEY")
	if cert != "" && key != "" {
		log.Fatal(app.ListenTLS(":8080", cert, key))
	}
	log.Fatal(app.Listen(":8080"))
}

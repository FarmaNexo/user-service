// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/farmanexo/user-service/internal/application/handlers"
	"github.com/farmanexo/user-service/internal/application/postprocessors"
	"github.com/farmanexo/user-service/internal/application/preprocessors"
	"github.com/farmanexo/user-service/internal/infrastructure/messaging"
	"github.com/farmanexo/user-service/internal/infrastructure/persistence/postgres"
	"github.com/farmanexo/user-service/internal/infrastructure/security"
	"github.com/farmanexo/user-service/internal/infrastructure/storage"
	"github.com/farmanexo/user-service/internal/presentation/http/controllers"
	"github.com/farmanexo/user-service/internal/presentation/http/middlewares"
	"github.com/farmanexo/user-service/internal/presentation/http/routes"
	"github.com/farmanexo/user-service/pkg/config"
	"github.com/farmanexo/user-service/pkg/mediator"

	// Swagger docs
	_ "github.com/farmanexo/user-service/docs"

	"go.uber.org/zap"
	pgdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// @title           FarmaNexo User Service API
// @version         1.0
// @description     Servicio de gestión de perfil de usuario para FarmaNexo - Microservicio con CQRS y Clean Architecture
// @termsOfService  https://farmanexo.pe/terms

// @contact.name    FarmaNexo API Support
// @contact.url     https://farmanexo.pe/support
// @contact.email   support@farmanexo.pe

// @license.name    Apache 2.0
// @license.url     http://www.apache.org/licenses/LICENSE-2.0.html

// @host            localhost:4002
// @BasePath        /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"

// @tag.name         Profile
// @tag.description  Endpoints de perfil de usuario

// @tag.name         Avatar
// @tag.description  Endpoints de avatar de usuario

// @tag.name         Addresses
// @tag.description  Endpoints de direcciones del usuario

// @tag.name         Preferences
// @tag.description  Endpoints de preferencias del usuario

// @tag.name         Health
// @tag.description  Endpoints de salud del servicio

func main() {
	env := getEnvironment()
	cfg, err := config.LoadConfig(env)
	if err != nil {
		panic(fmt.Sprintf("Error cargando configuración: %v", err))
	}

	logger := initLogger(cfg)
	defer logger.Sync()

	logger.Info("Iniciando User Service",
		zap.String("environment", cfg.Environment),
		zap.Int("port", cfg.Server.Port),
	)

	db := initDatabase(cfg, logger)

	logger.Info("Auto-migration deshabilitado - Usar migraciones manuales")

	// ========================================
	// REPOSITORIOS
	// ========================================
	profileRepo := postgres.NewProfileRepository(db, logger)
	addressRepo := postgres.NewAddressRepository(db, logger)
	prefsRepo := postgres.NewPreferencesRepository(db, logger)

	// ========================================
	// SERVICIOS
	// ========================================
	jwtService := security.NewJWTService(cfg.JWT.Secret, logger)

	// S3 File Storage
	fileStorage, err := storage.NewS3FileStorage(cfg.AWS, logger)
	if err != nil {
		logger.Fatal("Error inicializando S3 FileStorage", zap.Error(err))
	}

	// SQS Event Publisher
	eventPublisher, err := messaging.NewSQSEventPublisher(cfg.AWS, cfg.SQS, logger)
	if err != nil {
		logger.Fatal("Error inicializando SQS EventPublisher", zap.Error(err))
	}

	// SQS Event Consumer (Auth events)
	eventConsumer, err := messaging.NewSQSEventConsumer(cfg.AWS, cfg.SQS, profileRepo, prefsRepo, addressRepo, logger)
	if err != nil {
		logger.Fatal("Error inicializando SQS EventConsumer", zap.Error(err))
	}

	// Iniciar consumer en background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eventConsumer.Start(ctx)

	// ========================================
	// MEDIATOR
	// ========================================
	med := mediator.NewMediator()

	// ========================================
	// HANDLERS
	// ========================================

	// Profile handlers
	getProfileHandler := handlers.NewGetProfileHandler(profileRepo, logger)
	mediator.RegisterHandler(med, getProfileHandler)

	updateProfileHandler := handlers.NewUpdateProfileHandler(profileRepo, eventPublisher, logger)
	mediator.RegisterHandler(med, updateProfileHandler)

	// Avatar handlers
	uploadAvatarHandler := handlers.NewUploadAvatarHandler(profileRepo, fileStorage, eventPublisher, logger)
	mediator.RegisterHandler(med, uploadAvatarHandler)

	deleteAvatarHandler := handlers.NewDeleteAvatarHandler(profileRepo, fileStorage, eventPublisher, logger)
	mediator.RegisterHandler(med, deleteAvatarHandler)

	// Address handlers
	listAddressesHandler := handlers.NewListAddressesHandler(addressRepo, logger)
	mediator.RegisterHandler(med, listAddressesHandler)

	createAddressHandler := handlers.NewCreateAddressHandler(addressRepo, eventPublisher, logger)
	mediator.RegisterHandler(med, createAddressHandler)

	updateAddressHandler := handlers.NewUpdateAddressHandler(addressRepo, logger)
	mediator.RegisterHandler(med, updateAddressHandler)

	deleteAddressHandler := handlers.NewDeleteAddressHandler(addressRepo, logger)
	mediator.RegisterHandler(med, deleteAddressHandler)

	// Preferences handlers
	getPreferencesHandler := handlers.NewGetPreferencesHandler(prefsRepo, logger)
	mediator.RegisterHandler(med, getPreferencesHandler)

	updatePreferencesHandler := handlers.NewUpdatePreferencesHandler(prefsRepo, logger)
	mediator.RegisterHandler(med, updatePreferencesHandler)

	// ========================================
	// VALIDATORS (mediator interface{} pattern)
	// ========================================

	// ========================================
	// PREPROCESSORS Y POSTPROCESSORS
	// ========================================
	sanitizePreProcessor := preprocessors.NewSanitizeInputPreProcessor(logger)
	med.RegisterPreProcessor(sanitizePreProcessor)

	auditPostProcessor := postprocessors.NewLogAuditPostProcessor(logger)
	med.RegisterPostProcessor(auditPostProcessor)

	logger.Info("Mediator configurado",
		zap.Int("handlers", 10),
		zap.Int("preprocessors", 1),
		zap.Int("postprocessors", 1),
	)

	// ========================================
	// MIDDLEWARES
	// ========================================
	authMiddleware := middlewares.NewAuthMiddleware(jwtService, logger)

	// ========================================
	// CONTROLADORES Y RUTAS
	// ========================================
	userController := controllers.NewUserController(med, logger)
	router := routes.SetupRoutes(userController, authMiddleware)

	// ========================================
	// SERVIDOR HTTP
	// ========================================
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("Servidor HTTP iniciado",
			zap.String("address", server.Addr),
			zap.String("swagger_url", fmt.Sprintf("http://localhost:%d/swagger/index.html", cfg.Server.Port)),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Error iniciando servidor", zap.Error(err))
		}
	}()

	// ========================================
	// GRACEFUL SHUTDOWN
	// ========================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Iniciando graceful shutdown...")

	// Detener consumer
	eventConsumer.Stop()
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error en shutdown", zap.Error(err))
	}

	logger.Info("Servidor detenido exitosamente")
}

func getEnvironment() string {
	env := os.Getenv("ENV")
	if env == "" {
		env = "local"
	}
	return env
}

func initLogger(cfg *config.Config) *zap.Logger {
	var logger *zap.Logger
	var err error

	if cfg.IsProduction() || cfg.IsUAT() {
		logger, err = zap.NewProduction()
	} else {
		logger, err = zap.NewDevelopment()
	}

	if err != nil {
		panic(fmt.Sprintf("Error inicializando logger: %v", err))
	}

	return logger
}

func initDatabase(cfg *config.Config, logger *zap.Logger) *gorm.DB {
	gormLogLevel := gormlogger.Silent
	if cfg.IsDevelopment() {
		gormLogLevel = gormlogger.Info
	}

	gormLogger := gormlogger.Default.LogMode(gormLogLevel)

	db, err := gorm.Open(pgdriver.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		logger.Fatal("Error conectando a PostgreSQL",
			zap.Error(err),
			zap.String("host", cfg.Database.Host),
			zap.Int("port", cfg.Database.Port),
		)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logger.Fatal("Error obteniendo SQL DB", zap.Error(err))
	}

	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	logger.Info("Conexión a PostgreSQL establecida",
		zap.String("host", cfg.Database.Host),
		zap.String("database", cfg.Database.DBName),
	)

	return db
}

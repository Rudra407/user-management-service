package main

import (
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
	"github.com/user/user-management-service/api/handlers"
	"github.com/user/user-management-service/api/middleware"
	"github.com/user/user-management-service/config"
	"github.com/user/user-management-service/internal/models"
	"github.com/user/user-management-service/internal/repositories"
	"github.com/user/user-management-service/internal/services"
	"github.com/user/user-management-service/utils"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := utils.NewLogger(cfg.Log.Level)
	log := logger.WithField("service", "user-management")
	log.Info("Starting user management service")

	// Connect to database
	log.Info("Connecting to database...")
	db, err := gorm.Open("postgres", cfg.DBConnectionString())
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Set up database options
	db.LogMode(cfg.Log.Level == "debug")
	db.SingularTable(true)

	// Migrate database
	log.Info("Running database migrations...")
	if err := models.SetupUserTable(db); err != nil {
		log.WithError(err).Fatal("Failed to set up database tables")
	}

	// Migrate organization tables
	if err := models.SetupOrganizationTable(db); err != nil {
		log.WithError(err).Fatal("Failed to set up organization table")
	}

	// Migrate user_organization tables
	if err := models.SetupUserOrganizationTable(db); err != nil {
		log.WithError(err).Fatal("Failed to set up user_organization table")
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db, logger)
	orgRepo := repositories.NewOrganizationRepository(db, logger)

	// Initialize services
	userService := services.NewUserService(userRepo, cfg, logger, orgRepo)
	orgService := services.NewOrganizationService(orgRepo, cfg, logger)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService, logger)
	orgHandler := handlers.NewOrganizationHandler(orgService, logger)

	// Initialize echo
	e := echo.New()
	e.HideBanner = true

	// Set up middlewares
	e.Use(middleware.RequestLogger(logger))
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	// Create JWT middleware
	jwtMiddleware := middleware.JWTMiddleware(cfg, logger)

	// Register routes
	userHandler.RegisterRoutes(e, jwtMiddleware)
	orgHandler.RegisterRoutes(e)

	// Add health check endpoint
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "healthy"})
	})

	// Start server
	serverAddr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.WithField("addr", serverAddr).Info("Server starting")
	if err := e.Start(serverAddr); err != nil {
		log.WithError(err).Fatal("Server stopped unexpectedly")
	}
}

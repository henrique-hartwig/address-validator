package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/henrique/address-validator/config"
	"github.com/henrique/address-validator/internal/handlers"
	"github.com/henrique/address-validator/internal/middleware"
	"github.com/henrique/address-validator/internal/services"
	"github.com/joho/godotenv"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/henrique/address-validator/docs"
)

// @title           Address Validator API
// @version         1.0
// @description     API to validate and normalize addresses in free-form text with automatic typo corrections.
// @description     The API accepts addresses entered naturally by the user and returns normalized components.

// @host      localhost:3000
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter 'Bearer' followed by your token. Example: Bearer your_token_here

// @accept  json
// @produce json

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	cfg := config.Load()

	if cfg.APIToken == "" {
		log.Fatal("API_TOKEN environment variable is required")
	}

	cache, err := services.NewCacheService(
		cfg.RedisHost,
		cfg.RedisPort,
		cfg.RedisPassword,
		cfg.RedisDB,
		cfg.CacheTTL,
	)
	if err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}
	defer cache.Close()

	geocodingService := services.NewGeocodingService(
		cfg.GeocodingAAPIKey,
		cfg.GeocodingABaseURL,
		cfg.GeocodingBAPIKey,
		cfg.GeocodingBBaseURL,
		cache,
	)
	validatorService := services.NewValidatorService(geocodingService, cache)

	addressHandler := handlers.NewAddressHandler(validatorService)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())

	router.GET("/health", healthCheck)

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/api/v1")
	v1.Use(middleware.BearerAuth(cfg.APIToken))
	v1.Use(middleware.ValidateHeaders())
	{
		v1.POST("/validate-address", addressHandler.ValidateAddress)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("Swagger documentation: http://localhost:%s/swagger/index.html", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// healthCheck godoc
// @Summary      Health check
// @Description  Check if the service is working
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": "address-validator",
	})
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
	"typetype/internal/config"
	"typetype/internal/database"
	"typetype/internal/handlers"
	"typetype/internal/middleware"
	"typetype/internal/utils/jwt"

	// "typetype/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

type Server struct {
	router    *gin.Engine
	dbPool    *pgxpool.Pool
	dbQueries *database.Queries // ✅ Fixed: Prefixed with package name 'database'
	jwt       *jwt.JWTManager
}

func init() {
	appEnv := os.Getenv("APP_ENV")

	if appEnv == "" {
		appEnv = "development"
		os.Setenv("APP_ENV", appEnv)
	}

	envFile := fmt.Sprintf(".env.%s", appEnv)

	if err := godotenv.Load(envFile); err != nil {
		log.Printf("Warning: %s not found", envFile)
	}

	log.Printf("Running in %s mode", appEnv)
}

func main() {

	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	cfg := config.Load()
	jwtManager, err := jwt.NewJWTManager(
		cfg.AccessTokenSecret,
		cfg.RefreshTokenSecret,
		15*time.Minute,
		7*24*time.Hour,
	)
	if err != nil {
		log.Fatalf("JWT initialization failed: %v", err)
	}
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // Cache preflight requests for 12 hours
	}),
		gin.Logger(),
		gin.Recovery(),
	)

	// ✅ Initialized: Wrapped the connection pool with your SQLC wrapper
	server := &Server{
		router:    r,
		dbPool:    db,
		dbQueries: database.NewQueries(db),
		jwt:       jwtManager,
	}

	// Register application routes
	server.setupRoutes()

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s...\n", port)

	if err := r.Run(":" + port); err != nil {
		log.Printf("Server error: %v", err)
	}
}

// setupRoutes registers application routes
func (s *Server) setupRoutes() {
	userHandler := &handlers.UserHandler{
		DB:         s.dbQueries,
		JWTManager: s.jwt,
	}
	s.router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to TypeType API",
		})
	})
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "Server is running",
		})
	})

	auth := s.router.Group("/auth")
	{
		auth.POST("/register", userHandler.Register)
		auth.POST("/login", userHandler.LogIn)
		auth.POST("/refresh", userHandler.Refresh)
		auth.POST("/logout", middleware.AuthMiddleware(s.jwt), userHandler.LogOut)
	}

	users := s.router.Group("/user")
	{
		users.GET("/me", middleware.AuthMiddleware(s.jwt), userHandler.GetUser)
	}
}

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/lobster-lobby/lobster-lobby/config"
	"github.com/lobster-lobby/lobster-lobby/handlers"
	"github.com/lobster-lobby/lobster-lobby/middleware"
	"github.com/lobster-lobby/lobster-lobby/repository"
	"github.com/lobster-lobby/lobster-lobby/services"
)

func main() {
	_ = godotenv.Load()

	logger, _ := zap.NewProduction()
	if os.Getenv("ENV") == "development" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	cfg := config.Load()

	mongo, err := repository.Connect(cfg.MongoDBURI, cfg.MongoDBDatabase)
	if err != nil {
		logger.Fatal("failed to connect to MongoDB", zap.Error(err))
	}
	defer mongo.Disconnect()

	// Repositories & services
	userRepo := repository.NewUserRepository(mongo)
	refreshTokenRepo := repository.NewRefreshTokenRepository(mongo)
	policyRepo := repository.NewPolicyRepository(mongo)
	apiKeyRepo := repository.NewAPIKeyRepository(mongo)
	reputationRepo := repository.NewReputationRepository(mongo)
	jwtSvc := services.NewJWTService(cfg.JWTSecret)
	apiKeySvc := services.NewAPIKeyService()
	reputationSvc := services.NewReputationService(reputationRepo, userRepo)

	// Ensure DB indexes
	bgCtx := context.Background()
	if err := userRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure user indexes", zap.Error(err))
	}
	if err := refreshTokenRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure refresh token indexes", zap.Error(err))
	}
	if err := policyRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure policy indexes", zap.Error(err))
	}
	if err := apiKeyRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure API key indexes", zap.Error(err))
	}
	if err := reputationRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure reputation indexes", zap.Error(err))
	}

	authHandler := handlers.NewAuthHandler(userRepo, refreshTokenRepo, jwtSvc)
	policyHandler := handlers.NewPolicyHandler(policyRepo, userRepo, jwtSvc, logger, reputationSvc)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyRepo, apiKeySvc)

	rateLimiter := middleware.NewRateLimiter()

	if cfg.Env != "development" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.Logger(logger))
	r.Use(middleware.CORS(cfg.CORSOrigins))

	r.GET("/health", handlers.Health)
	r.GET("/api/health", handlers.Health)

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			auth.GET("/me", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), authHandler.Me)
		}

		policies := api.Group("/policies")
		policies.Use(middleware.RateLimit(rateLimiter))
		{
			policies.POST("", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), policyHandler.Create)
			policies.GET("", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), policyHandler.List)
			policies.GET("/:idOrSlug", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), policyHandler.Get)
			policies.PATCH("/:id", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), policyHandler.Update)
			policies.DELETE("/:id", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), policyHandler.Delete)
		}

		keys := api.Group("/keys")
		keys.Use(middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc))
		keys.Use(middleware.RateLimit(rateLimiter))
		{
			keys.POST("", apiKeyHandler.Create)
			keys.GET("", apiKeyHandler.List)
			keys.DELETE("/:id", apiKeyHandler.Delete)
		}
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		logger.Info("server starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}
	logger.Info("server stopped")
}

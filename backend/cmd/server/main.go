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
	"github.com/lobster-lobby/lobster-lobby/models"
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
	campaignRepo := repository.NewCampaignRepository(mongo)
	apiKeyRepo := repository.NewAPIKeyRepository(mongo)
	reputationRepo := repository.NewReputationRepository(mongo)
	jwtSvc := services.NewJWTService(cfg.JWTSecret)
	apiKeySvc := services.NewAPIKeyService()
	reputationSvc := services.NewReputationService(reputationRepo, userRepo)
	searchSvc := services.NewSearchService(cfg.MeilisearchURL, cfg.MeilisearchKey, logger)

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
	if err := campaignRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure campaign indexes", zap.Error(err))
	}

	commentRepo := repository.NewCommentRepository(mongo)
	if err := commentRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure comment indexes", zap.Error(err))
	}

	activityRepo := repository.NewActivityRepository(mongo)

	// Optionally rebuild search index from MongoDB on startup
	if cfg.RebuildIndex {
		bgCtxSearch := context.Background()
		go func() {
			policies, _, err := policyRepo.List(bgCtxSearch, repository.PolicyListOpts{Page: 1, PerPage: 10000, Sort: "hot"})
			if err != nil {
				logger.Warn("failed to load policies for index rebuild", zap.Error(err))
				return
			}
			ptrs := make([]*models.Policy, len(policies))
			for i := range policies {
				ptrs[i] = &policies[i]
			}
			if err := searchSvc.BulkIndex(bgCtxSearch, ptrs); err != nil {
				logger.Warn("failed to bulk index policies", zap.Error(err))
				return
			}
			logger.Info("search index rebuilt", zap.Int("count", len(policies)))
		}()
	}

	// Campaign comments
	campaignCommentRepo := repository.NewCampaignCommentRepository(mongo)
	if err := campaignCommentRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure campaign comment indexes", zap.Error(err))
	}

	// Campaign events
	campaignEventRepo := repository.NewCampaignEventRepository(mongo)
	if err := campaignEventRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure campaign event indexes", zap.Error(err))
	}

	authHandler := handlers.NewAuthHandler(userRepo, refreshTokenRepo, jwtSvc)
	policyHandler := handlers.NewPolicyHandler(policyRepo, userRepo, jwtSvc, logger, reputationSvc, searchSvc)
	campaignHandler := handlers.NewCampaignHandler(campaignRepo, policyRepo, userRepo, campaignEventRepo, jwtSvc, reputationSvc, logger)
	campaignCommentHandler := handlers.NewCampaignCommentHandler(campaignCommentRepo, campaignRepo, userRepo, campaignEventRepo, logger)
	campaignEventHandler := handlers.NewCampaignEventHandler(campaignEventRepo, campaignRepo, logger)
	apiKeyHandler := handlers.NewAPIKeyHandler(apiKeyRepo, apiKeySvc)
	dashboardHandler := handlers.NewDashboardHandler(userRepo, policyRepo, activityRepo, reputationSvc, logger)
	searchHandler := handlers.NewSearchHandler(searchSvc, logger)
	debateHandler := handlers.NewDebateHandler(commentRepo, policyRepo, logger, reputationSvc)

	researchRepo := repository.NewResearchRepository(mongo)
	if err := researchRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("research indexes", zap.Error(err))
	}
	researchHandler := handlers.NewResearchHandler(researchRepo, policyRepo, logger)

	nominationRepo := repository.NewNominationRepository(mongo)
	if err := nominationRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("nomination indexes", zap.Error(err))
	}

	summaryRepo := repository.NewSummaryPointRepository(mongo)
	if err := summaryRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("summary indexes", zap.Error(err))
	}
	summaryHandler := handlers.NewSummaryHandler(summaryRepo, commentRepo, policyRepo, userRepo, logger, reputationSvc)

	nominationHandler := handlers.NewNominationHandler(nominationRepo, policyRepo, userRepo, logger)

	// Representative lookup
	repRepo := repository.NewRepresentativeRepository(mongo)
	if err := repRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure representative indexes", zap.Error(err))
	}
	repSvc := services.NewRepresentativeService(repRepo, cfg.GoogleCivicAPIKey)
	repHandler := handlers.NewRepresentativeHandler(repSvc)

	// Campaign assets
	assetRepo := repository.NewAssetRepository(mongo)
	if err := assetRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure asset indexes", zap.Error(err))
	}

	assetHandler := handlers.NewAssetHandler(assetRepo, campaignRepo, userRepo, campaignEventRepo, logger)

	// Debates (standalone)
	debateRepo := repository.NewDebateRepository(mongo)
	if err := debateRepo.EnsureIndexes(bgCtx); err != nil {
		logger.Warn("failed to ensure debate indexes", zap.Error(err))
	}
	debatesHandler := handlers.NewDebatesHandler(debateRepo, logger, reputationSvc)
	moderationHandler := handlers.NewModerationHandler(debateRepo, userRepo, logger)

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
			policies.POST("/check-similar", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), policyHandler.CheckSimilar)
			policies.GET("", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), policyHandler.List)
			policies.GET("/:id", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), policyHandler.Get)
			policies.PATCH("/:id", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), policyHandler.Update)
			policies.DELETE("/:id", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), policyHandler.Delete)
			policies.POST("/:id/bookmark", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), dashboardHandler.BookmarkToggle)
			policies.POST("/:id/amendments", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), policyHandler.CreateAmendment)
			policies.GET("/:id/campaigns", campaignHandler.ListByPolicy)

			// Debate routes
			policies.POST("/:id/debate", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), debateHandler.CreateComment)
			policies.GET("/:id/debate", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), debateHandler.ListComments)
			policies.GET("/:id/debate/:commentId/replies", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), debateHandler.GetReplies)
			policies.PATCH("/:id/debate/:commentId", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), debateHandler.EditComment)
			policies.POST("/:id/debate/:commentId/react", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), debateHandler.ReactToComment)
			policies.POST("/:id/stance", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), debateHandler.SetStance)
			policies.GET("/:id/stance", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), debateHandler.GetStance)

			// Summary / bridging routes
			policies.GET("/:id/debate/summary", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), summaryHandler.ListSummary)
			policies.POST("/:id/debate/summary", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), summaryHandler.CreatePoint)
			policies.POST("/:id/debate/summary/:pointId/endorse", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), summaryHandler.EndorsePoint)
			policies.DELETE("/:id/debate/summary/:pointId/endorse", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), summaryHandler.RemoveEndorsement)

			// Nomination / campaign-readiness routes
			policies.POST("/:id/nominate-for-campaign", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), nominationHandler.Nominate)
			policies.POST("/:id/nominate-for-campaign/endorse", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), nominationHandler.Endorse)
			policies.GET("/:id/campaign-readiness", nominationHandler.CampaignReadiness)

			// Research routes
			policies.POST("/:id/research", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), researchHandler.Create)
			policies.GET("/:id/research", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), researchHandler.List)
			policies.GET("/:id/research/:researchId", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), researchHandler.GetByID)
			policies.PATCH("/:id/research/:researchId", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), researchHandler.Update)
			policies.POST("/:id/research/:researchId/react", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), researchHandler.React)
		}

		campaigns := api.Group("/campaigns")
		campaigns.Use(middleware.RateLimit(rateLimiter))
		{
			campaigns.POST("", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), campaignHandler.Create)
			campaigns.GET("", campaignHandler.List)
			campaigns.GET("/:id", campaignHandler.Get)
			campaigns.PATCH("/:id", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), campaignHandler.Update)
			campaigns.PATCH("/:id/status", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), campaignHandler.UpdateStatus)

			// Campaign assets
			campaigns.POST("/:id/assets", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), assetHandler.CreateTextAsset)
			campaigns.POST("/:id/assets/upload", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), assetHandler.UploadAsset)
			campaigns.GET("/:id/assets", assetHandler.List)
			campaigns.GET("/:id/assets/votes", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), assetHandler.BatchVotes)
			campaigns.GET("/:id/assets/:assetId", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), assetHandler.Get)
			campaigns.GET("/:id/assets/:assetId/file", assetHandler.ServeFile)
			campaigns.POST("/:id/assets/:assetId/download", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), assetHandler.Download)
			campaigns.POST("/:id/assets/:assetId/vote", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), assetHandler.Vote)
			campaigns.POST("/:id/assets/:assetId/share", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), assetHandler.Share)
			campaigns.PATCH("/:id/assets/:assetId", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), assetHandler.Update)
			campaigns.DELETE("/:id/assets/:assetId", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), assetHandler.Delete)

			// Campaign comments (discussion)
			campaigns.GET("/:id/comments", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), campaignCommentHandler.List)
			campaigns.POST("/:id/comments", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), campaignCommentHandler.Create)
			campaigns.PUT("/:id/comments/:commentId", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), campaignCommentHandler.Update)
			campaigns.DELETE("/:id/comments/:commentId", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), campaignCommentHandler.Delete)
			campaigns.POST("/:id/comments/:commentId/vote", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), campaignCommentHandler.Vote)

			// Campaign events (timeline) and metrics
			campaigns.GET("/:id/events", campaignEventHandler.List)
			campaigns.GET("/:id/metrics/activity", campaignEventHandler.GetActivity)
		}

		// Standalone debates
		debates := api.Group("/debates")
		debates.Use(middleware.RateLimit(rateLimiter))
		{
			debates.POST("", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), debatesHandler.CreateDebate)
			debates.GET("", debatesHandler.ListDebates)
			debates.GET("/:slug", middleware.OptionalAuth(jwtSvc, apiKeyRepo, apiKeySvc), debatesHandler.GetDebate)
			debates.POST("/:slug/arguments", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), debatesHandler.CreateArgument)
			debates.POST("/:slug/arguments/:id/vote", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), debatesHandler.VoteOnArgument)
			debates.POST("/:slug/arguments/:id/flag", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), debatesHandler.FlagArgument)
		}

		// Admin moderation
		admin := api.Group("/admin")
		admin.Use(middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc))
		{
			admin.GET("/moderation/queue", moderationHandler.GetQueue)
			admin.POST("/moderation/:id/action", moderationHandler.TakeAction)
		}

		api.GET("/search", searchHandler.Search)
		api.GET("/bookmarks", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), dashboardHandler.BookmarkList)
		api.GET("/dashboard", middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc), dashboardHandler.Dashboard)

		// Representative lookup routes (no auth required for reads)
		api.GET("/representatives", repHandler.List)
		api.GET("/representatives/:id", repHandler.GetByID)

		keys := api.Group("/keys")
		keys.Use(middleware.RequireAuth(jwtSvc, apiKeyRepo, apiKeySvc))
		keys.Use(middleware.RateLimit(rateLimiter))
		{
			keys.POST("", apiKeyHandler.Create)
			keys.GET("", apiKeyHandler.List)
			keys.DELETE("/:id", apiKeyHandler.Delete)
		}
	}

	// Serve frontend static files in production
	if cfg.StaticDir != "" {
		r.Static("/assets", cfg.StaticDir+"/assets")
		r.StaticFile("/favicon.ico", cfg.StaticDir+"/favicon.ico")
		r.StaticFile("/favicon.svg", cfg.StaticDir+"/favicon.svg")
		r.NoRoute(func(c *gin.Context) {
			c.File(cfg.StaticDir + "/index.html")
		})
		logger.Info("serving static files", zap.String("dir", cfg.StaticDir))
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

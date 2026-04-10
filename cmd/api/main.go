package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nla/internal/client/dohod"
	"nla/internal/client/moex"
	"nla/internal/client/openai"
	"nla/internal/config"
	"nla/internal/database"
	"nla/internal/handler"
	mongoRepo "nla/internal/mongo"
	"nla/internal/queue"
	"nla/internal/repository"
	"nla/internal/router"
	"nla/internal/service"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// PostgreSQL
	pgPool, err := database.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}
	defer pgPool.Close()
	log.Println("PostgreSQL connected")

	// MongoDB
	mongoDB, mongoClient, err := database.NewMongo(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("mongodb connection failed: %v", err)
	}
	defer mongoClient.Disconnect(ctx)
	log.Println("MongoDB connected")

	// Redis
	rdb, err := database.NewRedis(ctx, cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}
	defer rdb.Close()
	log.Println("Redis connected")

	// Repositories
	userRepo := repository.NewUserRepository(pgPool)
	favoriteRepo := repository.NewFavoriteRepository(pgPool)

	// Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiration)

	// MongoDB repositories
	analysisRepo := mongoRepo.NewAnalysisRepo(mongoDB)
	queueRepo := mongoRepo.NewQueueRepo(mongoDB)
	ratingRepo := mongoRepo.NewRatingRepo(mongoDB)
	chatRepo := mongoRepo.NewChatRepo(mongoDB)
	issuerRepo := mongoRepo.NewIssuerRepo(mongoDB)
	detailsRepo := mongoRepo.NewDetailsRepo(mongoDB)

	moexClient := moex.NewClient()
	dohodClient := dohod.NewClient()
	bondService := service.NewBondService(moexClient, rdb, issuerRepo, ratingRepo)

	// OpenAI client
	openaiClient := openai.NewClient(openai.Config{
		APIKey:  cfg.OpenAIKey,
		BaseURL: cfg.OpenAIBaseURL,
		Model:   cfg.OpenAIModel,
		Proxy:   cfg.OpenAIProxy,
		Timeout: time.Duration(cfg.OpenAITimeout) * time.Second,
	})

	// Services
	analysisService := service.NewAnalysisService(analysisRepo, openaiClient, "data/prompts/bond_analyst.txt")
	queueService := service.NewQueueService(queueRepo)
	ratingService := service.NewRatingService(ratingRepo)
	chatService := service.NewChatService(chatRepo, openaiClient)
	detailsService := service.NewDetailsService(dohodClient, detailsRepo, ratingRepo, issuerRepo)

	// Handlers
	h := handler.New(authService, cfg.JWTSecret)
	h.SetBondHandler(handler.NewBondHandler(bondService))
	h.SetAnalysisHandler(handler.NewAnalysisHandler(analysisService, queueService))
	h.SetRatingHandler(handler.NewRatingHandler(ratingService))
	h.SetChatHandler(handler.NewChatHandler(chatService))
	h.SetFavoriteHandler(handler.NewFavoriteHandler(favoriteRepo))
	h.SetDetailsHandler(handler.NewDetailsHandler(detailsService, queueService, bondService))

	// Run migrations
	if err := database.RunMigrations(ctx, pgPool); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	log.Println("Migrations applied")

	// Seed default credit ratings
	seedCtx, seedCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer seedCancel()
	if err := ratingService.SeedDefaults(seedCtx); err != nil {
		log.Printf("WARNING: rating seed failed: %v", err)
	} else {
		log.Println("Credit ratings seeded")
	}

	// Router
	r := router.New(h)

	// Queue worker (background goroutine)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	w := queue.NewWorker(queueService, analysisService, bondService, detailsService)
	go w.Run(workerCtx)

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("Server starting on :%s (env: %s)", cfg.Port, cfg.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	srv.Shutdown(shutdownCtx)
}

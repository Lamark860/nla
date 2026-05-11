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
	"nla/internal/queue"
	"nla/internal/repository"
	"nla/internal/router"
	"nla/internal/service"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// PostgreSQL — single backing store after Phase 1.
	pgPool, err := database.NewPostgres(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}
	defer pgPool.Close()
	log.Println("PostgreSQL connected")

	// Repositories
	userRepo := repository.NewUserRepository(pgPool)
	favoriteRepo := repository.NewFavoriteRepository(pgPool)
	analysisRepo := repository.NewAnalysisRepo(pgPool)
	queueRepo := repository.NewQueueRepo(pgPool)
	ratingRepo := repository.NewRatingRepo(pgPool)
	chatRepo := repository.NewChatRepo(pgPool)
	issuerRepo := repository.NewIssuerRepo(pgPool)
	detailsRepo := repository.NewDetailsRepo(pgPool)

	// Services
	authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiration)

	moexClient := moex.NewClient()
	dohodClient := dohod.NewClient()
	bondService := service.NewBondService(moexClient, issuerRepo, ratingRepo)

	openaiClient := openai.NewClient(openai.Config{
		APIKey:  cfg.OpenAIKey,
		BaseURL: cfg.OpenAIBaseURL,
		Model:   cfg.OpenAIModel,
		Proxy:   cfg.OpenAIProxy,
		Timeout: time.Duration(cfg.OpenAITimeout) * time.Second,
	})

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

	// Recompute Score/ScoreOrd for ratings persisted under the old (buggy) scoring
	recomputeCtx, recomputeCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer recomputeCancel()
	if n, err := ratingService.RecomputeAllScores(recomputeCtx); err != nil {
		log.Printf("WARNING: rating recompute failed: %v", err)
	} else if n > 0 {
		log.Printf("Recomputed scores for %d ratings", n)
	}

	// Router
	r := router.New(h)

	// Queue worker (background goroutine)
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	w := queue.NewWorker(queueService, analysisService, bondService, detailsService)
	go w.Run(workerCtx)

	// Sync missing bond_issuers and credit ratings in background
	go func() {
		syncCtx, syncCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer syncCancel()
		if n, err := bondService.SyncMissingIssuers(syncCtx); err != nil {
			log.Printf("WARNING: issuer sync error: %v", err)
		} else if n > 0 {
			log.Printf("Synced %d missing bond_issuers", n)
		}

		ratingCtx, ratingCancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer ratingCancel()
		if n, err := bondService.SyncMissingRatingsFromMoex(ratingCtx); err != nil {
			log.Printf("WARNING: MOEX rating sync error: %v", err)
		} else if n > 0 {
			log.Printf("Synced ratings for %d emitters from MOEX CCI", n)
		}
	}()

	// Server
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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

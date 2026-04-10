package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nla/internal/client/dohod"
	"nla/internal/config"
	"nla/internal/database"
	mongoRepo "nla/internal/mongo"
	"nla/internal/service"
)

func main() {
	onlyMissing := flag.Bool("only-missing", true, "Skip emitters that already have ratings")
	delay := flag.Duration("delay", 2*time.Second, "Delay between dohod.ru requests")
	flag.Parse()

	cfg := config.Load()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT/SIGTERM for graceful stop
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		fmt.Println("\nStopping... (will finish current request)")
		cancel()
	}()

	// MongoDB
	mongoDB, mongoClient, err := database.NewMongo(ctx, cfg.MongoURI, cfg.MongoDB)
	if err != nil {
		log.Fatalf("mongodb: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	// Repos
	issuerRepo := mongoRepo.NewIssuerRepo(mongoDB)
	ratingRepo := mongoRepo.NewRatingRepo(mongoDB)
	detailsRepo := mongoRepo.NewDetailsRepo(mongoDB)
	dohodClient := dohod.NewClient()

	detailsSvc := service.NewDetailsService(dohodClient, detailsRepo, ratingRepo, issuerRepo)

	mode := "only missing"
	if !*onlyMissing {
		mode = "all emitters"
	}
	fmt.Printf("=== Rating Sync ===\n")
	fmt.Printf("Mode: %s, delay: %v\n\n", mode, *delay)

	result, err := detailsSvc.SyncAllRatings(ctx, *onlyMissing, *delay)
	if err != nil && ctx.Err() == nil {
		log.Fatalf("sync error: %v", err)
	}

	fmt.Printf("\n=== Done ===\n")
	fmt.Printf("Total emitters:  %d\n", result.TotalEmitters)
	fmt.Printf("Already rated:   %d\n", result.AlreadyRated)
	fmt.Printf("Processed:       %d\n", result.Processed)
	fmt.Printf("Newly rated:     %d\n", result.NewlyRated)
	fmt.Printf("Errors:          %d\n", result.Errors)
}

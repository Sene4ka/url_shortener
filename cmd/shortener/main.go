package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sene4ka/url_shortener/configs"
	"github.com/Sene4ka/url_shortener/internal/generator"
	"github.com/Sene4ka/url_shortener/internal/repository"
	"github.com/Sene4ka/url_shortener/internal/shortener"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	config := configs.LoadConfig()

	var repo shortener.LinkRepository
	if config.Database.UseInMemory {
		repo = repository.NewInMemoryRepository()
	} else {
		pool, err := pgxpool.New(context.Background(), fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=%s",
			config.Postgres.User,
			config.Postgres.Password,
			config.Postgres.Host,
			config.Postgres.Port,
			config.Postgres.Name,
			config.Postgres.SSLMode,
		))
		if err != nil {
			log.Fatalf("Unable to connect to database: %s", err.Error())
		}
		defer pool.Close()
		repo = repository.NewPostgresRepository(pool)
	}

	gen, err := generator.NewIdGenerator(config.Sonyflake.StartTime, config.Sonyflake.SequenceBits)
	if err != nil {
		log.Fatalf("Unable to create sonyflake instance: %s", err.Error())
	}

	service := shortener.NewLinkService(gen, repo)

	handler := shortener.NewLinkHandler(service, config)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down Shortener service...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		err = handler.Shutdown(ctx)
		if err != nil {
			log.Fatalf("Unable to shutdown Shortener service: %s", err.Error())
		}
	}()

	log.Printf("Shortener server starting on %s:%s", config.Server.Host, config.Server.Port)
	if err = handler.Start(); err != nil {
		log.Fatalf("Failed to start Shortener server: %s", err.Error())
	}
}

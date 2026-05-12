package main

import (
	"context"
	"log"
	"time"

	"example.com/test/internal/repository"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	dbPool, err := repository.NewPool(ctx)
	if err != nil {
		log.Fatalf("connect database: %v", err)
	}
	defer dbPool.Close()

	if err := repository.Migrate(ctx, dbPool); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	log.Println("database migrations completed")
}

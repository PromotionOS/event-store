package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/pressly/goose/v3"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/promotionos/event-store/internal/consumer"
	"github.com/promotionos/event-store/internal/repository"
)

func main() {
	dbURL := os.Getenv("DB_URL")
	redisURL := os.Getenv("REDIS_URL")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	// DB with eventstore schema
	dsn := dbURL + "?search_path=eventstore"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Run migrations
	sqlDB, _ := db.DB()
	runMigrations(sqlDB)

	// Redis
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}
	redisClient := redis.NewClient(opt)
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Wire and start consumer
	repo := repository.NewEventStoreRepositoryImpl(db)
	c := consumer.NewAllEventsConsumer(repo, redisClient)
	c.Start()

	log.Println("Event Store started — consuming all domain events")

	// Minimal API
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "event-store"})
	})

	r.GET("/events", func(c *gin.Context) {
		tenantID := c.Query("tenantId")
		context := c.Query("context")
		aggregateID := c.Query("aggregateId")
		from := c.Query("from")
		to := c.Query("to")

		if tenantID == "" {
			c.JSON(400, gin.H{"error": "tenantId required"})
			return
		}

		var events interface{}
		var queryErr error

		if aggregateID != "" {
			events, queryErr = repo.FindByAggregate(aggregateID, tenantID)
		} else if context != "" {
			events, queryErr = repo.FindByContext(context, tenantID, from, to)
		} else {
			c.JSON(400, gin.H{"error": "provide aggregateId or context"})
			return
		}

		if queryErr != nil {
			c.JSON(500, gin.H{"error": queryErr.Error()})
			return
		}
		c.JSON(200, events)
	})

	log.Printf("Event Store listening on port %s", port)
	r.Run(fmt.Sprintf(":%s", port))
}

func runMigrations(db *sql.DB) {
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Goose dialect error: %v", err)
	}
	if err := goose.Up(db, "db/migrations"); err != nil {
		log.Printf("Migration warning: %v", err)
	}
}

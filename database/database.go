package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	gooseLock "github.com/pressly/goose/v3/lock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: failed to load .env: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func DSN() string {
	host := envOrDefault("GO_BLOG_DATABASE_HOST", "shared-postgres")
	port := envOrDefault("GO_BLOG_DATABASE_PORT", "5432")
	user := envOrDefault("GO_BLOG_DATABASE_USER", "goblog")
	dbname := envOrDefault("GO_BLOG_DATABASE_NAME", "goblog")
	password := os.Getenv("GO_BLOG_DATABASE_PASSWORD")

	if password == "" {
		log.Fatal("GO_BLOG_DATABASE_PASSWORD is required")
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)
}

func Connect() *gorm.DB {
	db, err := gorm.Open(postgres.Open(DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return db
}

func runMigrations(migrations embed.FS, sqlDB *sql.DB) {
	migrationFS, err := fs.Sub(migrations, "migrations")
	if err != nil {
		log.Fatalf("Failed to open migrations directory: %v", err)
	}

	sessionLocker, err := gooseLock.NewPostgresSessionLocker()
	if err != nil {
		log.Fatalf("Failed to create migration session locker: %v", err)
	}

	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		sqlDB,
		migrationFS,
		goose.WithSessionLocker(sessionLocker),
	)
	if err != nil {
		log.Fatalf("Failed to create migration provider: %v", err)
	}

	if _, err := provider.Up(context.Background()); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
}

func RunMigrations(migrations embed.FS) {
	db := Connect()

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database handle: %v", err)
	}
	defer func() { _ = sqlDB.Close() }()

	runMigrations(migrations, sqlDB)
}

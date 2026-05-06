package main

import (
	"fmt"
	"log"
	"os"

	"auth-service/internal/config"
	"auth-service/internal/platform/db"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down|version>")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	m, err := migrate.New("file://migrations", db.MigrateDSN(cfg.Database))
	if err != nil {
		log.Fatalf("failed to init migrate: %v", err)
	}
	defer m.Close()

	switch os.Args[1] {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate up failed: %v", err)
		}
		fmt.Println("migrations applied")
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migrate down failed: %v", err)
		}
		fmt.Println("migrations rolled back")
	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("failed to get version: %v", err)
		}
		fmt.Printf("version: %d, dirty: %v\n", v, dirty)
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}
}

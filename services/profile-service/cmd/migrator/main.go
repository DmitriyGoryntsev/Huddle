package main

import (
	"flag"
	"profile-service/internal/config"
	"profile-service/pkg/logger"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var (
		command        string
		databaseURL    string
		migrationsPath string
		forceVersion   int
	)

	cfg, err := config.New()
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Logger)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	flag.StringVar(&command, "command", "up", "Migration command: up | down | force")
	flag.StringVar(&databaseURL, "database", "", "Database connection URL")
	flag.StringVar(&migrationsPath, "path", "migrations", "Path to migrations")
	flag.IntVar(&forceVersion, "version", -1, "Force migration version")
	flag.Parse()

	if databaseURL == "" {
		log.Fatal("database URL is required")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		databaseURL,
	)
	if err != nil {
		log.Fatalf("failed to init migrate: %v", err)
	}

	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			log.Infof("source close error: %v", srcErr)
		}
		if dbErr != nil {
			log.Infof("db close error: %v", dbErr)
		}
	}()

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration up failed: %v", err)
		}
		log.Info("migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("migration down failed: %v", err)
		}
		log.Info("migrations reverted successfully")

	case "force":
		if forceVersion < 0 {
			log.Fatal("force command requires -version flag")
		}
		if err := m.Force(forceVersion); err != nil {
			log.Fatalf("migration force failed: %v", err)
		}
		log.Infof("migration forced to version %d\n", forceVersion)

	default:
		log.Fatalf("unknown command: %s", command)
	}
}

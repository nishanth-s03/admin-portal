package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"

	"admin-portal/internal/shared/database"
)

func main() {
	log.Println("🚀 DB migration started")

	if err := godotenv.Load(); err != nil {
		log.Fatal("❌ Failed to load .env:", err)
	}

	cfg := database.LoadConfig()

	// 1️⃣ Connect to maintenance DB
	sysDB, err := database.Open(cfg, "postgres")
	if err != nil {
		log.Fatal("❌ Failed to connect to postgres DB:", err)
	}
	defer sysDB.Close()

	// 2️⃣ Ensure DB exists
	exists, err := databaseExists(sysDB, cfg.DBName)
	if err != nil {
		log.Fatal(err)
	}

	if !exists {
		log.Printf("📦 Database %s not found, creating...\n", cfg.DBName)
		if err := createDatabase(sysDB, cfg.DBName); err != nil {
			log.Fatal(err)
		}
		log.Println("✅ Database created")
	} else {
		log.Println("ℹ️ Database already exists")
	}

	sysDB.Close()

	// 3️⃣ Connect to application DB
	db, err := database.Open(cfg, cfg.DBName)
	if err != nil {
		log.Fatal("❌ Failed to connect to application DB:", err)
	}
	defer db.Close()

	// 4️⃣ Ensure schema_migrations table
	if err := ensureSchemaMigrations(db); err != nil {
		log.Fatal(err)
	}

	// 5️⃣ Apply migrations safely
	if err := applyMigrations(db); err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	log.Println("🎉 Migration completed successfully")
}

/* ---------------- Helpers ---------------- */

func databaseExists(db *sql.DB, name string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)`,
		name,
	).Scan(&exists)
	return exists, err
}

func createDatabase(db *sql.DB, name string) error {
	_, err := db.Exec(fmt.Sprintf(`CREATE DATABASE "%s"`, name))
	return err
}

func ensureSchemaMigrations(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(50) PRIMARY KEY,
			applied_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func applyMigrations(db *sql.DB) error {
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return err
	}

	sort.Strings(files)
	ctx := context.Background()

	for _, file := range files {
		version := migrationVersion(file)

		applied, err := isMigrationApplied(db, version)
		if err != nil {
			return err
		}

		if applied {
			log.Println("⏭ Skipping", file)
			continue
		}

		log.Println("▶ Applying", file)

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s failed: %w", file, err)
		}

		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO schema_migrations (version) VALUES ($1)`,
			version,
		); err != nil {
			tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	return nil
}

func migrationVersion(path string) string {
	base := filepath.Base(path)
	return strings.Split(base, "_")[0]
}

func isMigrationApplied(db *sql.DB, version string) (bool, error) {
	var exists bool
	err := db.QueryRow(
		`SELECT EXISTS (
			SELECT 1 FROM schema_migrations WHERE version = $1
		)`,
		version,
	).Scan(&exists)
	return exists, err
}

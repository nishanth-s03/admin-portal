package main

import (
	"sort"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"

	"admin-portal/internal/shared/database"
)

func main() {
	log.Println("🚀 DB migration started")

	// Load env
	if err := godotenv.Load(); err != nil {
		log.Fatal("❌ Failed to load .env:", err)
	}

	cfg := database.LoadConfig()

	// 1️⃣ Connect to maintenance DB (postgres)
	sysDB, err := database.Open(cfg, "postgres")
	if err != nil {
		log.Fatal("❌ Failed to connect to postgres DB:", err)
	}
	defer sysDB.Close()

	// 2️⃣ Check DB existence
	exists, err := databaseExists(sysDB, cfg.DBName)
	if err != nil {
		log.Fatal("❌ DB existence check failed:", err)
	}

	if !exists {
		log.Printf("📦 Database %s not found, creating...\n", cfg.DBName)
		if err := createDatabase(sysDB, cfg.DBName); err != nil {
			log.Fatal("❌ Failed to create DB:", err)
		}
		log.Println("✅ Database created")
	} else {
		log.Println("ℹ️ Database already exists")
	}

	sysDB.Close()

	// 3️⃣ Connect to application DB
	appDB, err := database.Open(cfg, cfg.DBName)
	if err != nil {
		log.Fatal("❌ Failed to connect to application DB:", err)
	}
	defer appDB.Close()

	// 4️⃣ Check schema
	ok, err := schemaExists(appDB)
	if err != nil {
		log.Fatal("❌ Schema check failed:", err)
	}

	if ok {
		log.Println("✅ Tables & triggers already exist. Exiting.")
		return
	}

	// 5️⃣ Run migrations
	log.Println("📜 Running migrations...")
	if err := runMigrations(appDB); err != nil {
		log.Fatal("❌ Migration failed:", err)
	}

	log.Println("🎉 Migration completed successfully")
}

/*-------------------------------------------- Helper Functions --------------------------------------------*/

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

func schemaExists(db *sql.DB) (bool, error) {
	requiredTables := []string{
		"users",
		"password_master",
		"login_logs",
	}

	for _, table := range requiredTables {
		var exists bool
		err := db.QueryRow(
			`SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_name = $1
			)`,
			table,
		).Scan(&exists)

		if err != nil || !exists {
			return false, err
		}
	}

	// check trigger existence
	var triggerExists bool
	err := db.QueryRow(
		`SELECT EXISTS (
			SELECT 1 FROM pg_trigger
			WHERE tgname = 'trg_users_updated'
		)`,
	).Scan(&triggerExists)

	return triggerExists, err
}

func runMigrations(db *sql.DB) error {
	files, err := filepath.Glob("migrations/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(files)
	ctx := context.Background()

	for _, file := range files {
		log.Println("▶ Applying", file)

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return err
		}

		if _, err := db.ExecContext(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("migration %s failed: %w", file, err)
		}
	}

	return nil
}

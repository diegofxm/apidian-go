package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"apidian-go/database/engine"
)

func main() {
	// Cargar .env desde la raíz del proyecto
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Obtener comando
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Conectar a la base de datos
	db, err := connectDB()
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer db.Close()

	// Rutas de migraciones y seeds
	migrationsPath := filepath.Join("database", "migrations")
	seedsPath := filepath.Join("database", "seeds")

	// Crear migrator
	migrator := engine.NewMigrator(db, migrationsPath, seedsPath)

	// Ejecutar comando
	switch command {
	case "migrate":
		if err := migrator.Migrate(); err != nil {
			log.Fatalf("❌ Migration failed: %v", err)
		}

	case "fresh":
		if err := migrator.Fresh(); err != nil {
			log.Fatalf("❌ Fresh failed: %v", err)
		}

	case "status":
		if err := migrator.Status(); err != nil {
			log.Fatalf("❌ Status failed: %v", err)
		}

	case "seed":
		if err := migrator.Seed(); err != nil {
			log.Fatalf("❌ Seed failed: %v", err)
		}

	default:
		fmt.Printf("❌ Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// connectDB conecta a PostgreSQL usando variables de entorno
func connectDB() (*sql.DB, error) {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "")
	dbname := getEnv("DB_NAME", "apidian")
	sslmode := getEnv("DB_SSLMODE", "disable")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	fmt.Printf("✓ Connected to database: %s@%s:%s/%s\n\n", user, host, port, dbname)
	return db, nil
}

// getEnv obtiene una variable de entorno o retorna un valor por defecto
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// printUsage imprime el uso del CLI
func printUsage() {
	fmt.Println("📦 APIDIAN Database Migrator")
	fmt.Println("\nUsage:")
	fmt.Println("  go run database/cmd/migrate/main.go <command>")
	fmt.Println("\nCommands:")
	fmt.Println("  migrate    Run all pending migrations")
	fmt.Println("  fresh      Drop all tables and re-run all migrations")
	fmt.Println("  status     Show migration status")
	fmt.Println("  seed       Run all seed files")
	fmt.Println("\nExamples:")
	fmt.Println("  go run database/cmd/migrate/main.go migrate")
	fmt.Println("  go run database/cmd/migrate/main.go fresh")
	fmt.Println("  go run database/cmd/migrate/main.go status")
	fmt.Println("  go run database/cmd/migrate/main.go seed")
}

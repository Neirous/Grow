package main

import (
	"context"
	"embed"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"grow/internal/db"
	"grow/internal/handlers"
	"grow/internal/service"
)

//go:embed templates/*
var templatesFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	port := flag.String("port", "8080", "HTTP port")
	dbPath := flag.String("db", "grow.db", "SQLite database path")
	flag.Parse()

	// Override db path from env (for Docker)
	if envDB := os.Getenv("DB_PATH"); envDB != "" {
		*dbPath = envDB
	}

	log.Printf("grow starting...")
	log.Printf("Database: %s", *dbPath)

	if err := db.Init(*dbPath); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	// Register API routes
	handlers.RegisterAbilityRoutes(mux)
	handlers.RegisterActivityRoutes(mux)
	handlers.RegisterCompleteRoutes(mux)
	handlers.RegisterLogRoutes(mux)
	handlers.RegisterDashboardRoutes(mux)
	handlers.RegisterGoalRoutes(mux)
	handlers.RegisterStreakRoutes(mux)
	handlers.RegisterSettingsRoutes(mux)
	handlers.RegisterNotionRoutes(mux)
	handlers.RegisterStaticRoutes(mux, templatesFS, staticFS)

	// Start background scheduler
	service.StartScheduler(db.DB)

	server := &http.Server{
		Addr:    ":" + *port,
		Handler: mux,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down...", sig)
		server.Shutdown(context.Background())
	}()

	log.Printf("grow is running at http://localhost:%s", *port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	log.Println("grow stopped")
}

package main

import (
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"

	"library-management-system/config"
	"library-management-system/models"
	"library-management-system/routes"
	"library-management-system/utils"
)

func main() {
	// Load .env file if exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it, using environment variables")
	}

	// Load application configuration
	config.LoadConfig()

	// Initialize database connection
	config.InitDB()
	defer config.CloseDB()

	// Initialize session
	utils.InitSession()

	// Initialize templates
	if err := utils.InitTemplates(); err != nil {
		log.Fatalf("Failed to initialize templates: %v", err)
	}
	log.Println("Templates initialized successfully")

	// Create default librarian account if no users exist
	if err := models.CreateDefaultLibrarian(); err != nil {
		log.Printf("Warning: Failed to create default librarian: %v", err)
	}

	// Create file server for static files
	fileServer := http.FileServer(http.Dir(config.AppConfig.Template.StaticDir))
	http.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Set up routes
	routes.SetupRoutes()

	// Determine host and port
	host := config.AppConfig.Server.Host
	port := config.AppConfig.Server.Port
	addr := host + ":" + port

	// Run the server
	log.Printf("Server starting on %s (host=%s, port=%s)\n", addr, host, port)

	// Create a server with timeouts
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start the server and log any errors
	log.Fatal(server.ListenAndServe())
}

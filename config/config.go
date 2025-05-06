package config

import (
	"os"
	"time"
)

// AppConfig holds the application configuration
var AppConfig struct {
	Server struct {
		Port string
		Host string
	}
	Database struct {
		DSN      string
		MaxConns int
		MaxIdle  int
		Timeout  time.Duration
	}
	Session struct {
		Secret   string
		Name     string
		Lifetime time.Duration
		HttpOnly bool
		Secure   bool
	}
	Template struct {
		CacheParsedTemplates bool
		TemplatesDir         string
		StaticDir            string
	}
}

// LoadConfig loads the application configuration from environment variables
func LoadConfig() {
	// Set server configuration
	AppConfig.Server.Port = getEnvWithDefault("PORT", "10000")
	AppConfig.Server.Host = getEnvWithDefault("HOST", "0.0.0.0")

	// Set database configuration
	AppConfig.Database.DSN = getEnvWithDefault("DATABASE_URL", "")
	AppConfig.Database.MaxConns = 10
	AppConfig.Database.MaxIdle = 5
	AppConfig.Database.Timeout = 5 * time.Second

	// Set session configuration
	AppConfig.Session.Secret = getEnvWithDefault("SESSION_SECRET", "library-management-system-secret")
	AppConfig.Session.Name = "library_session"
	AppConfig.Session.Lifetime = 24 * time.Hour
	AppConfig.Session.HttpOnly = true
	AppConfig.Session.Secure = false // Set to true in production with HTTPS

	// Set template configuration
	AppConfig.Template.CacheParsedTemplates = false // Set to true in production
	AppConfig.Template.TemplatesDir = "templates"
	AppConfig.Template.StaticDir = "static"
}

// getEnvWithDefault gets an environment variable or returns a default value
func getEnvWithDefault(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

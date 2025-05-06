package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq" // PostgreSQL driver
)

var db *sql.DB

// InitDB initializes the database connection
func InitDB() {
	var err error

	// Open database connection
	db, err = sql.Open("postgres", AppConfig.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Set connection pool configuration
	db.SetMaxOpenConns(AppConfig.Database.MaxConns)
	db.SetMaxIdleConns(AppConfig.Database.MaxIdle)
	db.SetConnMaxLifetime(AppConfig.Database.Timeout)

	// Verify connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database successfully")

	// Initialize database schema
	err = initSchema()
	if err != nil {
		log.Fatalf("Failed to initialize database schema: %v", err)
	}
}

// GetDB returns the database connection
func GetDB() *sql.DB {
	return db
}

// CloseDB closes the database connection
func CloseDB() {
	if db != nil {
		db.Close()
	}
}

// initSchema creates the necessary tables if they don't exist
func initSchema() error {
	// Drop users table if broken (for dev/testing only)
	// _, _ = db.Exec(`DROP TABLE IF EXISTS users`)

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			email VARCHAR(100) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL,
			student_id VARCHAR(50),
			phone VARCHAR(20),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS books (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			author VARCHAR(100) NOT NULL,
			isbn VARCHAR(20) UNIQUE NOT NULL,
			publisher VARCHAR(100),
			publication_year INT,
			category VARCHAR(50),
			description TEXT,
			quantity INT NOT NULL DEFAULT 1,
			available INT NOT NULL DEFAULT 1,
			added_by INT REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create books table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS borrows (
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL REFERENCES users(id),
			book_id INT NOT NULL REFERENCES books(id),
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			borrow_date TIMESTAMP,
			due_date TIMESTAMP,
			return_date TIMESTAMP,
			approved_by INT REFERENCES users(id),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to create borrows table: %v", err)
	}

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'librarian'`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for librarian: %v", err)
	}

	if count == 0 {
		_, err = db.Exec(`
			INSERT INTO users (name, email, password_hash, role)
			VALUES ('Admin', 'admin@library.com', '$2a$10$Xsq0d2aRQTbhGI9GZW3uQeT8YXNBJKlHGxnXz0HkstbOGpK/BWrjW', 'librarian')
		`)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %v", err)
		}
		log.Println("Created default librarian user: admin@library.com / password")
	}

	err = db.QueryRow(`SELECT COUNT(*) FROM books`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check for books: %v", err)
	}

	if count == 0 {
		sampleBooks := []struct {
			Title           string
			Author          string
			ISBN            string
			Publisher       string
			PublicationYear int
			Category        string
			Description     string
			Quantity        int
		}{
			{
				Title:           "To Kill a Mockingbird",
				Author:          "Harper Lee",
				ISBN:            "9780061120084",
				Publisher:       "HarperCollins",
				PublicationYear: 1960,
				Category:        "Fiction",
				Description:     "A novel about racial injustice and moral growth in the American South.",
				Quantity:        5,
			},
			{
				Title:           "1984",
				Author:          "George Orwell",
				ISBN:            "9780451524935",
				Publisher:       "Signet Classics",
				PublicationYear: 1949,
				Category:        "Fiction",
				Description:     "A dystopian novel about totalitarianism and mass surveillance.",
				Quantity:        3,
			},
		}

		for _, book := range sampleBooks {
			_, err = db.Exec(`
				INSERT INTO books (title, author, isbn, publisher, publication_year, category, description, quantity, available)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
			`, book.Title, book.Author, book.ISBN, book.Publisher, book.PublicationYear, book.Category, book.Description, book.Quantity)
			if err != nil {
				return fmt.Errorf("failed to add sample book: %v", err)
			}
		}

		log.Println("Added sample books to the database")
	}

	log.Println("Database schema initialized successfully")
	return nil
}

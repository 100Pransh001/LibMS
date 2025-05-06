package main

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    
    _ "github.com/lib/pq"
    "golang.org/x/crypto/bcrypt"
)

func main() {
    // Get database URL from environment
    dbURL := os.Getenv("DATABASE_URL")
    if dbURL == "" {
        log.Fatal("DATABASE_URL environment variable is not set")
    }
    
    // Database connection
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()
    
    // Generate hashed password
    password := "admin123"
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        log.Fatal("Failed to hash password:", err)
    }
    
    // Check if admin exists
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM users WHERE email = $1", "admin@library.com").Scan(&count)
    if err != nil {
        log.Fatal("Failed to query database:", err)
    }
    
    if count > 0 {
        // Update existing admin
        _, err = db.Exec(
            "UPDATE users SET name = $1, password_hash = $2, role = $3 WHERE email = $4",
            "Admin Librarian", 
            string(hashedPassword), 
            "librarian", 
            "admin@library.com",
        )
        if err != nil {
            log.Fatal("Failed to update admin:", err)
        }
        fmt.Println("Admin user updated successfully")
    } else {
        // Create new admin
        _, err = db.Exec(
            "INSERT INTO users (name, email, password_hash, role) VALUES ($1, $2, $3, $4)",
            "Admin Librarian", 
            "admin@library.com", 
            string(hashedPassword), 
            "librarian",
        )
        if err != nil {
            log.Fatal("Failed to create admin:", err)
        }
        fmt.Println("Admin user created successfully")
    }
}

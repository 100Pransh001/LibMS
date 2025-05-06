package models

import (
        "database/sql"
        "errors"
        "time"

        "golang.org/x/crypto/bcrypt"

        "library-management-system/config"
)

var (
        ErrInvalidCredentials = errors.New("invalid email or password")
        ErrDuplicateEmail     = errors.New("email already exists")
)

// User represents a user in the system
type User struct {
        ID           int
        Name         string
        Email        string
        Password     string
        PasswordHash string
        Role         string
        StudentID    sql.NullString
        Phone        sql.NullString
        CreatedAt    time.Time
        UpdatedAt    time.Time
        
        // Computed properties
        IsLibrarian bool
        IsStudent   bool
}

// GetUserByID retrieves a user by ID
func GetUserByID(id int) (*User, error) {
        db := config.GetDB()
        user := &User{}
        
        // Execute query
        err := db.QueryRow(`
                SELECT id, name, email, password_hash, role, student_id, phone, created_at, updated_at
                FROM users
                WHERE id = $1
        `, id).Scan(
                &user.ID,
                &user.Name,
                &user.Email,
                &user.PasswordHash,
                &user.Role,
                &user.StudentID,
                &user.Phone,
                &user.CreatedAt,
                &user.UpdatedAt,
        )
        if err != nil {
                if err == sql.ErrNoRows {
                        return nil, errors.New("user not found")
                }
                return nil, err
        }
        
        // Set computed properties
        user.IsLibrarian = user.Role == "librarian"
        user.IsStudent = user.Role == "student"
        
        return user, nil
}

// GetUserByEmail retrieves a user by email
func GetUserByEmail(email string) (*User, error) {
        db := config.GetDB()
        user := &User{}
        
        // Execute query
        err := db.QueryRow(`
                SELECT id, name, email, password_hash, role, student_id, phone, created_at, updated_at
                FROM users
                WHERE email = $1
        `, email).Scan(
                &user.ID,
                &user.Name,
                &user.Email,
                &user.PasswordHash,
                &user.Role,
                &user.StudentID,
                &user.Phone,
                &user.CreatedAt,
                &user.UpdatedAt,
        )
        if err != nil {
                if err == sql.ErrNoRows {
                        return nil, errors.New("user not found")
                }
                return nil, err
        }
        
        // Set computed properties
        user.IsLibrarian = user.Role == "librarian"
        user.IsStudent = user.Role == "student"
        
        return user, nil
}

// Create saves a new user to the database
func (u *User) Create() error {
        db := config.GetDB()
        
        // Check if email exists
        var count int
        err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = $1`, u.Email).Scan(&count)
        if err != nil {
                return err
        }
        if count > 0 {
                return ErrDuplicateEmail
        }
        
        // Hash password
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
        if err != nil {
                return err
        }
        
        // Execute query
        err = db.QueryRow(`
                INSERT INTO users (name, email, password_hash, role, student_id, phone)
                VALUES ($1, $2, $3, $4, $5, $6)
                RETURNING id, created_at, updated_at
        `, u.Name, u.Email, string(hashedPassword), u.Role, u.StudentID, u.Phone).Scan(
                &u.ID,
                &u.CreatedAt,
                &u.UpdatedAt,
        )
        if err != nil {
                return err
        }
        
        // Set computed properties
        u.IsLibrarian = u.Role == "librarian"
        u.IsStudent = u.Role == "student"
        
        return nil
}

// Update updates an existing user in the database
func (u *User) Update() error {
        db := config.GetDB()
        
        // Check if email exists
        var count int
        err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = $1 AND id != $2`, u.Email, u.ID).Scan(&count)
        if err != nil {
                return err
        }
        if count > 0 {
                return ErrDuplicateEmail
        }
        
        // Execute query
        _, err = db.Exec(`
                UPDATE users
                SET name = $1, email = $2, role = $3, student_id = $4, phone = $5, updated_at = CURRENT_TIMESTAMP
                WHERE id = $6
        `, u.Name, u.Email, u.Role, u.StudentID, u.Phone, u.ID)
        if err != nil {
                return err
        }
        
        // Set computed properties
        u.IsLibrarian = u.Role == "librarian"
        u.IsStudent = u.Role == "student"
        
        return nil
}

// UpdatePassword updates a user's password
func (u *User) UpdatePassword(newPassword string) error {
        db := config.GetDB()
        
        // Hash password
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
        if err != nil {
                return err
        }
        
        // Execute query
        _, err = db.Exec(`
                UPDATE users
                SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
                WHERE id = $2
        `, string(hashedPassword), u.ID)
        
        return err
}

// Delete removes a user from the database
func (u *User) Delete() error {
        db := config.GetDB()
        
        // Execute query
        _, err := db.Exec(`DELETE FROM users WHERE id = $1`, u.ID)
        
        return err
}

// Authenticate checks if the provided email and password match a user
func Authenticate(email, password string) (*User, error) {
        // Debug logging
        println("Authentication attempt for email:", email)
        
        // Get user by email
        user, err := GetUserByEmail(email)
        if err != nil {
                println("Error finding user:", err.Error())
                return nil, ErrInvalidCredentials
        }
        
        println("Found user with hash:", user.PasswordHash)
        
        // Check password
        err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
        if err != nil {
                println("Password comparison failed:", err.Error())
                return nil, ErrInvalidCredentials
        }
        
        println("Authentication successful for user:", user.Name)
        return user, nil
}

// GetAllUsers retrieves all users from the database
func GetAllUsers() ([]*User, error) {
        db := config.GetDB()
        
        // Execute query
        rows, err := db.Query(`
                SELECT id, name, email, password_hash, role, student_id, phone, created_at, updated_at
                FROM users
                ORDER BY id
        `)
        if err != nil {
                return nil, err
        }
        defer rows.Close()
        
        // Parse rows
        var users []*User
        for rows.Next() {
                user := &User{}
                err := rows.Scan(
                        &user.ID,
                        &user.Name,
                        &user.Email,
                        &user.PasswordHash,
                        &user.Role,
                        &user.StudentID,
                        &user.Phone,
                        &user.CreatedAt,
                        &user.UpdatedAt,
                )
                if err != nil {
                        return nil, err
                }
                
                // Set computed properties
                user.IsLibrarian = user.Role == "librarian"
                user.IsStudent = user.Role == "student"
                
                users = append(users, user)
        }
        
        // Check for errors
        if err := rows.Err(); err != nil {
                return nil, err
        }
        
        return users, nil
}

// GetAllStudents retrieves all users with role "student"
func GetAllStudents() ([]*User, error) {
        db := config.GetDB()
        
        // Execute query
        rows, err := db.Query(`
                SELECT id, name, email, password_hash, role, student_id, phone, created_at, updated_at
                FROM users
                WHERE role = 'student'
                ORDER BY id
        `)
        if err != nil {
                return nil, err
        }
        defer rows.Close()
        
        // Parse rows
        var users []*User
        for rows.Next() {
                user := &User{}
                err := rows.Scan(
                        &user.ID,
                        &user.Name,
                        &user.Email,
                        &user.PasswordHash,
                        &user.Role,
                        &user.StudentID,
                        &user.Phone,
                        &user.CreatedAt,
                        &user.UpdatedAt,
                )
                if err != nil {
                        return nil, err
                }
                
                // Set computed properties
                user.IsLibrarian = false // All are students
                user.IsStudent = true    // All are students
                
                users = append(users, user)
        }
        
        // Check for errors
        if err := rows.Err(); err != nil {
                return nil, err
        }
        
        return users, nil
}

// CountUsers returns the total number of users
func CountUsers() (int, error) {
        db := config.GetDB()
        
        var count int
        err := db.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&count)
        
        return count, err
}

// CreateDefaultLibrarian creates a default librarian account if no users exist
func CreateDefaultLibrarian() error {
        // Check if any users exist
        count, err := CountUsers()
        if err != nil {
                return err
        }
        
        // If users exist, don't create default admin
        if count > 0 {
                return nil
        }
        
        // Create default librarian
        librarian := &User{
                Name:     "Admin Librarian",
                Email:    "admin@library.com",
                Password: "admin123", 
                Role:     "librarian",
                Phone:    sql.NullString{String: "1234567890", Valid: true},
        }
        
        return librarian.Create()
}
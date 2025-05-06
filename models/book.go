package models

import (
        "database/sql"
        "errors"
        "fmt"
        "time"

        "library-management-system/config"
)

// Book represents a book in the library system
type Book struct {
        ID              int
        Title           string
        Author          string
        ISBN            string
        Publisher       string
        PublicationYear int
        Category        string
        Genre           string    // Alias for Category
        Description     string
        Quantity        int
        Available       int
        AvailableCopy   int       // Alias for Available
        TotalCopies     int       // Alias for Quantity
        AddedBy         sql.NullInt64 // Using NullInt64 to handle NULL values in the database
        CreatedAt       time.Time
        UpdatedAt       time.Time
        
        // Computed properties
        AddedByUser     *User
}

// SetAliasFields sets alias fields for template compatibility
func (b *Book) SetAliasFields() {
        b.Genre = b.Category
        b.AvailableCopy = b.Available
        b.TotalCopies = b.Quantity
}

// GetBookByID retrieves a book by its ID
func GetBookByID(id int) (*Book, error) {
        db := config.GetDB()
        
        // Execute query
        book := &Book{}
        err := db.QueryRow(`
                SELECT id, title, author, isbn, publisher, publication_year, category, description, 
                        quantity, available, added_by, created_at, updated_at
                FROM books
                WHERE id = $1
        `, id).Scan(
                &book.ID,
                &book.Title,
                &book.Author,
                &book.ISBN,
                &book.Publisher,
                &book.PublicationYear,
                &book.Category,
                &book.Description,
                &book.Quantity,
                &book.Available,
                &book.AddedBy,
                &book.CreatedAt,
                &book.UpdatedAt,
        )
        
        if err != nil {
                if err == sql.ErrNoRows {
                        return nil, errors.New("book not found")
                }
                return nil, err
        }
        
        // Get added by user if available
        if book.AddedBy.Valid && book.AddedBy.Int64 > 0 {
                addedByID := int(book.AddedBy.Int64)
                book.AddedByUser, _ = GetUserByID(addedByID)
        }
        
        // Set alias fields for template compatibility
        book.SetAliasFields()
        
        return book, nil
}

// GetBooks retrieves books with optional search and pagination
func GetBooks(search string, searchBy string, page int) ([]*Book, error) {
        db := config.GetDB()
        
        // Build query
        query := `
                SELECT id, title, author, isbn, publisher, publication_year, category, description, 
                        quantity, available, added_by, created_at, updated_at
                FROM books
        `
        
        var args []interface{}
        var whereClause string
        
        // Add search condition if provided
        if search != "" {
                switch searchBy {
                case "title":
                        whereClause = "WHERE title ILIKE $1"
                case "author":
                        whereClause = "WHERE author ILIKE $1"
                case "isbn":
                        whereClause = "WHERE isbn ILIKE $1"
                case "category":
                        whereClause = "WHERE category ILIKE $1"
                default:
                        whereClause = "WHERE title ILIKE $1 OR author ILIKE $1 OR isbn ILIKE $1"
                }
                args = append(args, "%"+search+"%")
        }
        
        // Add where clause if exists
        if whereClause != "" {
                query += " " + whereClause
        }
        
        // Add order by and pagination
        query += " ORDER BY title ASC"
        
        // Add pagination
        pageSize := 10
        offset := (page - 1) * pageSize
        query += " LIMIT $" + fmt.Sprintf("%d", len(args)+1) + " OFFSET $" + fmt.Sprintf("%d", len(args)+2)
        args = append(args, pageSize, offset)
        
        // Execute query
        rows, err := db.Query(query, args...)
        if err != nil {
                return nil, err
        }
        defer rows.Close()
        
        // Parse rows
        var books []*Book
        for rows.Next() {
                book := &Book{}
                err := rows.Scan(
                        &book.ID,
                        &book.Title,
                        &book.Author,
                        &book.ISBN,
                        &book.Publisher,
                        &book.PublicationYear,
                        &book.Category,
                        &book.Description,
                        &book.Quantity,
                        &book.Available,
                        &book.AddedBy,
                        &book.CreatedAt,
                        &book.UpdatedAt,
                )
                if err != nil {
                        return nil, err
                }
                
                // Set alias fields for template compatibility
                book.SetAliasFields()
                
                books = append(books, book)
        }
        
        // Check for errors
        if err := rows.Err(); err != nil {
                return nil, err
        }
        
        return books, nil
}

// CountBooks counts books with optional search
func CountBooks(search string, searchBy string) (int, error) {
        db := config.GetDB()
        
        // Build query
        query := "SELECT COUNT(*) FROM books"
        
        var args []interface{}
        var whereClause string
        
        // Add search condition if provided
        if search != "" {
                switch searchBy {
                case "title":
                        whereClause = "WHERE title ILIKE $1"
                case "author":
                        whereClause = "WHERE author ILIKE $1"
                case "isbn":
                        whereClause = "WHERE isbn ILIKE $1"
                case "category":
                        whereClause = "WHERE category ILIKE $1"
                default:
                        whereClause = "WHERE title ILIKE $1 OR author ILIKE $1 OR isbn ILIKE $1"
                }
                args = append(args, "%"+search+"%")
        }
        
        // Add where clause if exists
        if whereClause != "" {
                query += " " + whereClause
        }
        
        // Execute query
        var count int
        var err error
        
        if len(args) > 0 {
                err = db.QueryRow(query, args...).Scan(&count)
        } else {
                err = db.QueryRow(query).Scan(&count)
        }
        
        if err != nil {
                return 0, err
        }
        
        return count, nil
}

// GetAllBooks retrieves all books
func GetAllBooks() ([]*Book, error) {
        db := config.GetDB()
        
        // Execute query
        rows, err := db.Query(`
                SELECT id, title, author, isbn, publisher, publication_year, category, description, 
                        quantity, available, added_by, created_at, updated_at
                FROM books
                ORDER BY title ASC
        `)
        if err != nil {
                return nil, err
        }
        defer rows.Close()
        
        // Parse rows
        var books []*Book
        for rows.Next() {
                book := &Book{}
                err := rows.Scan(
                        &book.ID,
                        &book.Title,
                        &book.Author,
                        &book.ISBN,
                        &book.Publisher,
                        &book.PublicationYear,
                        &book.Category,
                        &book.Description,
                        &book.Quantity,
                        &book.Available,
                        &book.AddedBy,
                        &book.CreatedAt,
                        &book.UpdatedAt,
                )
                if err != nil {
                        return nil, err
                }
                
                // Set alias fields for template compatibility
                book.SetAliasFields()
                
                books = append(books, book)
        }
        
        // Check for errors
        if err := rows.Err(); err != nil {
                return nil, err
        }
        
        return books, nil
}

// Create saves a new book to the database
func (b *Book) Create() error {
        db := config.GetDB()
        
        // Execute query
        err := db.QueryRow(`
                INSERT INTO books (title, author, isbn, publisher, publication_year, category, description, 
                        quantity, available, added_by)
                VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
                RETURNING id, created_at, updated_at
        `,
                b.Title,
                b.Author,
                b.ISBN,
                b.Publisher,
                b.PublicationYear,
                b.Category,
                b.Description,
                b.Quantity,
                b.Available,
                b.AddedBy, // sql.NullInt64 will be handled correctly by database/sql
        ).Scan(
                &b.ID,
                &b.CreatedAt,
                &b.UpdatedAt,
        )
        
        return err
}

// Update updates an existing book in the database
func (b *Book) Update() error {
        db := config.GetDB()
        
        // Execute query
        _, err := db.Exec(`
                UPDATE books
                SET title = $1, author = $2, isbn = $3, publisher = $4, publication_year = $5, 
                        category = $6, description = $7, quantity = $8, available = $9, 
                        added_by = $10, updated_at = CURRENT_TIMESTAMP
                WHERE id = $11
        `,
                b.Title,
                b.Author,
                b.ISBN,
                b.Publisher,
                b.PublicationYear,
                b.Category,
                b.Description,
                b.Quantity,
                b.Available,
                b.AddedBy, // Include AddedBy in update
                b.ID,
        )
        
        return err
}

// Delete removes a book from the database
func (b *Book) Delete() error {
        db := config.GetDB()
        
        // Execute query
        _, err := db.Exec("DELETE FROM books WHERE id = $1", b.ID)
        
        return err
}

// IsbnExists checks if a book with the given ISBN already exists
func IsbnExists(isbn string) (bool, error) {
        db := config.GetDB()
        
        var count int
        err := db.QueryRow("SELECT COUNT(*) FROM books WHERE isbn = $1", isbn).Scan(&count)
        if err != nil {
                return false, err
        }
        
        return count > 0, nil
}

// IsbnExistsExcept checks if a book with the given ISBN exists, excluding a specific book ID
func IsbnExistsExcept(isbn string, id int) (bool, error) {
        db := config.GetDB()
        
        var count int
        err := db.QueryRow("SELECT COUNT(*) FROM books WHERE isbn = $1 AND id != $2", isbn, id).Scan(&count)
        if err != nil {
                return false, err
        }
        
        return count > 0, nil
}

// CountAllBooks returns the total number of books
func CountAllBooks() (int, error) {
        db := config.GetDB()
        
        var count int
        err := db.QueryRow("SELECT COUNT(*) FROM books").Scan(&count)
        
        return count, err
}

// CountAvailableBooks returns the total number of books that have at least one copy available
func CountAvailableBooks() (int, error) {
        db := config.GetDB()
        
        var count int
        err := db.QueryRow("SELECT COUNT(*) FROM books WHERE available > 0").Scan(&count)
        
        return count, err
}

// GetTopBorrowedBooks returns the top n most borrowed books
func GetTopBorrowedBooks(limit int) ([]*Book, error) {
        db := config.GetDB()
        
        // Execute query
        rows, err := db.Query(`
                SELECT b.id, b.title, b.author, b.isbn, b.publisher, b.publication_year, b.category, 
                        b.description, b.quantity, b.available, b.added_by, b.created_at, b.updated_at, 
                        COUNT(br.id) as borrow_count
                FROM books b
                JOIN borrows br ON b.id = br.book_id
                GROUP BY b.id
                ORDER BY borrow_count DESC
                LIMIT $1
        `, limit)
        if err != nil {
                return nil, err
        }
        defer rows.Close()
        
        // Parse rows
        var books []*Book
        for rows.Next() {
                book := &Book{}
                var borrowCount int
                err := rows.Scan(
                        &book.ID,
                        &book.Title,
                        &book.Author,
                        &book.ISBN,
                        &book.Publisher,
                        &book.PublicationYear,
                        &book.Category,
                        &book.Description,
                        &book.Quantity,
                        &book.Available,
                        &book.AddedBy,
                        &book.CreatedAt,
                        &book.UpdatedAt,
                        &borrowCount,
                )
                if err != nil {
                        return nil, err
                }
                
                // Set alias fields for template compatibility
                book.SetAliasFields()
                
                books = append(books, book)
        }
        
        // Check for errors
        if err := rows.Err(); err != nil {
                return nil, err
        }
        
        return books, nil
}

// HasActiveOrPendingBorrows checks if a book has any active or pending borrows
func HasActiveOrPendingBorrows(bookID int) (bool, error) {
        db := config.GetDB()
        
        var count int
        err := db.QueryRow(`
                SELECT COUNT(*) 
                FROM borrows 
                WHERE book_id = $1 AND status IN ('pending', 'approved')
        `, bookID).Scan(&count)
        
        if err != nil {
                return false, err
        }
        
        return count > 0, nil
}
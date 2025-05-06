package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"library-management-system/config"
)

// Borrow status constants
const (
	BorrowStatusPending  = "pending"
	BorrowStatusApproved = "approved"
	BorrowStatusRejected = "rejected"
	BorrowStatusReturned = "returned"
)

// Borrow represents a book borrowing record
type Borrow struct {
	ID            int
	UserID        int
	BookID        int
	Status        string
	BorrowDate    *time.Time
	DueDate       *time.Time
	ReturnDate    *time.Time
	ApprovedBy    *int
	CreatedAt     time.Time
	UpdatedAt     time.Time
	RejectionNote string

	// Computed properties
	User      *User
	Book      *Book
	Approver  *User
	IsOverdue bool
}

// GetBorrowByID retrieves a borrow record by ID
func GetBorrowByID(id int) (*Borrow, error) {
	db := config.GetDB()

	// Execute query
	borrow := &Borrow{}
	err := db.QueryRow(`
                SELECT id, user_id, book_id, status, borrow_date, due_date, return_date, 
                        approved_by, created_at, updated_at
                FROM borrows
                WHERE id = $1
        `, id).Scan(
		&borrow.ID,
		&borrow.UserID,
		&borrow.BookID,
		&borrow.Status,
		&borrow.BorrowDate,
		&borrow.DueDate,
		&borrow.ReturnDate,
		&borrow.ApprovedBy,
		&borrow.CreatedAt,
		&borrow.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("borrow record not found")
		}
		return nil, err
	}

	// Get related user
	borrow.User, _ = GetUserByID(borrow.UserID)

	// Get related book
	borrow.Book, _ = GetBookByID(borrow.BookID)

	// Get approver if exists
	if borrow.ApprovedBy != nil {
		borrow.Approver, _ = GetUserByID(*borrow.ApprovedBy)
	}

	// Check if overdue
	if borrow.Status == BorrowStatusApproved && borrow.DueDate != nil {
		if time.Now().After(*borrow.DueDate) {
			borrow.IsOverdue = true
		}
	}

	return borrow, nil
}

// GetBorrowByUserAndBook retrieves a borrow record by user and book ID
func GetBorrowByUserAndBook(userID, bookID int, status string) (*Borrow, error) {
	db := config.GetDB()

	// Build query
	query := `
                SELECT id, user_id, book_id, status, borrow_date, due_date, return_date, 
                        approved_by, created_at, updated_at
                FROM borrows
                WHERE user_id = $1 AND book_id = $2
        `
	args := []interface{}{userID, bookID}

	// Add status filter if provided
	if status != "" {
		query += " AND status = $3"
		args = append(args, status)
	} else {
		query += " ORDER BY created_at DESC LIMIT 1"
	}

	// Execute query
	borrow := &Borrow{}
	err := db.QueryRow(query, args...).Scan(
		&borrow.ID,
		&borrow.UserID,
		&borrow.BookID,
		&borrow.Status,
		&borrow.BorrowDate,
		&borrow.DueDate,
		&borrow.ReturnDate,
		&borrow.ApprovedBy,
		&borrow.CreatedAt,
		&borrow.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Get related user
	borrow.User, _ = GetUserByID(borrow.UserID)

	// Get related book
	borrow.Book, _ = GetBookByID(borrow.BookID)

	// Get approver if exists
	if borrow.ApprovedBy != nil {
		borrow.Approver, _ = GetUserByID(*borrow.ApprovedBy)
	}

	// Check if overdue
	if borrow.Status == BorrowStatusApproved && borrow.DueDate != nil {
		if time.Now().After(*borrow.DueDate) {
			borrow.IsOverdue = true
		}
	}

	return borrow, nil
}

// GetCurrentBorrow retrieves the current active borrow for a user and book
func GetCurrentBorrow(userID, bookID int) (*Borrow, error) {
	return GetBorrowByUserAndBook(userID, bookID, BorrowStatusApproved)
}

// HasPendingBorrowRequest checks if a user has a pending borrow request for a book
func HasPendingBorrowRequest(userID, bookID int) (bool, error) {
	borrow, err := GetBorrowByUserAndBook(userID, bookID, BorrowStatusPending)
	if err != nil {
		return false, err
	}
	return borrow != nil, nil
}

// IsCurrentlyBorrowing checks if a user is currently borrowing a book
func IsCurrentlyBorrowing(userID, bookID int) (bool, error) {
	borrow, err := GetCurrentBorrow(userID, bookID)
	if err != nil {
		return false, err
	}
	return borrow != nil, nil
}

// CreateBorrowRequest creates a new borrow request
func CreateBorrowRequest(userID, bookID int) error {
	db := config.GetDB()

	// Execute query
	_, err := db.Exec(`
                INSERT INTO borrows (user_id, book_id, status)
                VALUES ($1, $2, $3)
        `, userID, bookID, BorrowStatusPending)

	return err
}

// ApproveBorrow approves a borrow request
func ApproveBorrow(id int, approverID int, dueDate time.Time) error {
	db := config.GetDB()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get borrow request
	var bookID int
	var status string
	err = tx.QueryRow("SELECT book_id, status FROM borrows WHERE id = $1", id).Scan(&bookID, &status)
	if err != nil {
		return err
	}

	// Check if request is pending
	if status != BorrowStatusPending {
		return errors.New("borrow request is not in pending status")
	}

	// Check if book is available
	var available int
	err = tx.QueryRow("SELECT available FROM books WHERE id = $1", bookID).Scan(&available)
	if err != nil {
		return err
	}

	if available <= 0 {
		return errors.New("no copies available for borrowing")
	}

	// Update borrow request status
	borrowDate := time.Now()
	_, err = tx.Exec(`
                UPDATE borrows
                SET status = $1, borrow_date = $2, due_date = $3, approved_by = $4, updated_at = CURRENT_TIMESTAMP
                WHERE id = $5
        `, BorrowStatusApproved, borrowDate, dueDate, approverID, id)
	if err != nil {
		return err
	}

	// Update book available count
	_, err = tx.Exec(`
                UPDATE books
                SET available = available - 1, updated_at = CURRENT_TIMESTAMP
                WHERE id = $1
        `, bookID)
	if err != nil {
		return err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

// RejectBorrow rejects a borrow request
func RejectBorrow(id int, approverID int) error {
	db := config.GetDB()

	// Execute query
	_, err := db.Exec(`
                UPDATE borrows
                SET status = $1, approved_by = $2, updated_at = CURRENT_TIMESTAMP
                WHERE id = $3 AND status = $4
        `, BorrowStatusRejected, approverID, id, BorrowStatusPending)

	return err
}

// ReturnBook marks a book as returned
func ReturnBook(id int) error {
	db := config.GetDB()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get borrow request
	var bookID int
	var status string
	err = tx.QueryRow("SELECT book_id, status FROM borrows WHERE id = $1", id).Scan(&bookID, &status)
	if err != nil {
		return err
	}

	// Check if book is currently borrowed
	if status != BorrowStatusApproved {
		return errors.New("book is not currently borrowed")
	}

	// Update borrow request status
	returnDate := time.Now()
	_, err = tx.Exec(`
                UPDATE borrows
                SET status = $1, return_date = $2, updated_at = CURRENT_TIMESTAMP
                WHERE id = $3
        `, BorrowStatusReturned, returnDate, id)
	if err != nil {
		return err
	}

	// Update book available count
	_, err = tx.Exec(`
                UPDATE books
                SET available = available + 1, updated_at = CURRENT_TIMESTAMP
                WHERE id = $1
        `, bookID)
	if err != nil {
		return err
	}

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		return err
	}

	// After the transaction completes successfully, check for pending reservations
	go func() {
		// We're using a goroutine to avoid blocking the return operation
		// if there's an error with reservation processing
		_ = ProcessReservationsForBook(bookID)
	}()

	return nil
}

// GetAllPendingBorrows retrieves all pending borrow requests
func GetAllPendingBorrows() ([]*Borrow, error) {
	db := config.GetDB()

	// Execute query
	rows, err := db.Query(`
                SELECT b.id, b.user_id, b.book_id, b.status, b.borrow_date, b.due_date, b.return_date, 
                        b.approved_by, b.created_at, b.updated_at
                FROM borrows b
                WHERE b.status = $1
                ORDER BY b.created_at ASC
        `, BorrowStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse rows
	var borrows []*Borrow
	for rows.Next() {
		borrow := &Borrow{}
		err := rows.Scan(
			&borrow.ID,
			&borrow.UserID,
			&borrow.BookID,
			&borrow.Status,
			&borrow.BorrowDate,
			&borrow.DueDate,
			&borrow.ReturnDate,
			&borrow.ApprovedBy,
			&borrow.CreatedAt,
			&borrow.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get related user
		borrow.User, _ = GetUserByID(borrow.UserID)

		// Get related book
		borrow.Book, _ = GetBookByID(borrow.BookID)

		borrows = append(borrows, borrow)
	}

	// Check for errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return borrows, nil
}

// GetActiveBorrows retrieves all active borrows
func GetActiveBorrows() ([]*Borrow, error) {
	db := config.GetDB()

	// Execute query
	rows, err := db.Query(`
                SELECT b.id, b.user_id, b.book_id, b.status, b.borrow_date, b.due_date, b.return_date, 
                        b.approved_by, b.created_at, b.updated_at
                FROM borrows b
                WHERE b.status = $1
                ORDER BY b.due_date ASC
        `, BorrowStatusApproved)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse rows
	var borrows []*Borrow
	for rows.Next() {
		borrow := &Borrow{}
		err := rows.Scan(
			&borrow.ID,
			&borrow.UserID,
			&borrow.BookID,
			&borrow.Status,
			&borrow.BorrowDate,
			&borrow.DueDate,
			&borrow.ReturnDate,
			&borrow.ApprovedBy,
			&borrow.CreatedAt,
			&borrow.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get related user
		borrow.User, _ = GetUserByID(borrow.UserID)

		// Get related book
		borrow.Book, _ = GetBookByID(borrow.BookID)

		// Get approver if exists
		if borrow.ApprovedBy != nil {
			borrow.Approver, _ = GetUserByID(*borrow.ApprovedBy)
		}

		borrows = append(borrows, borrow)
	}

	// Check for errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return borrows, nil
}

// GetOverdueBooks retrieves all overdue books
func GetOverdueBooks() ([]*Borrow, error) {
	db := config.GetDB()

	// Execute query
	rows, err := db.Query(`
                SELECT b.id, b.user_id, b.book_id, b.status, b.borrow_date, b.due_date, b.return_date, 
                        b.approved_by, b.created_at, b.updated_at
                FROM borrows b
                WHERE b.status = $1 AND b.due_date < CURRENT_TIMESTAMP
                ORDER BY b.due_date ASC
        `, BorrowStatusApproved)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse rows
	var borrows []*Borrow
	for rows.Next() {
		borrow := &Borrow{}
		err := rows.Scan(
			&borrow.ID,
			&borrow.UserID,
			&borrow.BookID,
			&borrow.Status,
			&borrow.BorrowDate,
			&borrow.DueDate,
			&borrow.ReturnDate,
			&borrow.ApprovedBy,
			&borrow.CreatedAt,
			&borrow.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get related user
		borrow.User, _ = GetUserByID(borrow.UserID)

		// Get related book
		borrow.Book, _ = GetBookByID(borrow.BookID)

		// Get approver if exists
		if borrow.ApprovedBy != nil {
			borrow.Approver, _ = GetUserByID(*borrow.ApprovedBy)
		}

		borrows = append(borrows, borrow)
	}

	// Check for errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return borrows, nil
}

// GetActiveUserBorrows retrieves active borrows for a specific user
func GetActiveUserBorrows(userID int) ([]*Borrow, error) {
	db := config.GetDB()

	// Execute query
	rows, err := db.Query(`
                SELECT b.id, b.user_id, b.book_id, b.status, b.borrow_date, b.due_date, b.return_date, 
                        b.approved_by, b.created_at, b.updated_at
                FROM borrows b
                WHERE b.user_id = $1 AND b.status = $2
                ORDER BY b.due_date ASC
        `, userID, BorrowStatusApproved)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse rows
	var borrows []*Borrow
	for rows.Next() {
		borrow := &Borrow{}
		err := rows.Scan(
			&borrow.ID,
			&borrow.UserID,
			&borrow.BookID,
			&borrow.Status,
			&borrow.BorrowDate,
			&borrow.DueDate,
			&borrow.ReturnDate,
			&borrow.ApprovedBy,
			&borrow.CreatedAt,
			&borrow.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get related book
		borrow.Book, _ = GetBookByID(borrow.BookID)

		borrows = append(borrows, borrow)
	}

	// Check for errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return borrows, nil
}

// GetPendingUserBorrows retrieves pending borrow requests for a specific user
func GetPendingUserBorrows(userID int) ([]*Borrow, error) {
	db := config.GetDB()

	// Execute query
	rows, err := db.Query(`
                SELECT b.id, b.user_id, b.book_id, b.status, b.borrow_date, b.due_date, b.return_date, 
                        b.approved_by, b.created_at, b.updated_at
                FROM borrows b
                WHERE b.user_id = $1 AND b.status = $2
                ORDER BY b.created_at DESC
        `, userID, BorrowStatusPending)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse rows
	var borrows []*Borrow
	for rows.Next() {
		borrow := &Borrow{}
		err := rows.Scan(
			&borrow.ID,
			&borrow.UserID,
			&borrow.BookID,
			&borrow.Status,
			&borrow.BorrowDate,
			&borrow.DueDate,
			&borrow.ReturnDate,
			&borrow.ApprovedBy,
			&borrow.CreatedAt,
			&borrow.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get related book
		borrow.Book, _ = GetBookByID(borrow.BookID)

		borrows = append(borrows, borrow)
	}

	// Check for errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return borrows, nil
}

// GetPastUserBorrows retrieves past (returned or rejected) borrows for a specific user
func GetPastUserBorrows(userID int) ([]*Borrow, error) {
	db := config.GetDB()

	// Execute query
	rows, err := db.Query(`
                SELECT b.id, b.user_id, b.book_id, b.status, b.borrow_date, b.due_date, b.return_date, 
                        b.approved_by, b.created_at, b.updated_at
                FROM borrows b
                WHERE b.user_id = $1 AND b.status IN ($2, $3)
                ORDER BY b.updated_at DESC
        `, userID, BorrowStatusReturned, BorrowStatusRejected)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse rows
	var borrows []*Borrow
	for rows.Next() {
		borrow := &Borrow{}
		err := rows.Scan(
			&borrow.ID,
			&borrow.UserID,
			&borrow.BookID,
			&borrow.Status,
			&borrow.BorrowDate,
			&borrow.DueDate,
			&borrow.ReturnDate,
			&borrow.ApprovedBy,
			&borrow.CreatedAt,
			&borrow.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get related book
		borrow.Book, _ = GetBookByID(borrow.BookID)

		borrows = append(borrows, borrow)
	}

	// Check for errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return borrows, nil
}

// GetBorrowsWithFilters retrieves borrows with search and filtering options
func GetBorrowsWithFilters(searchTerm, status string, page, itemsPerPage int) ([]*Borrow, int, error) {
	db := config.GetDB()

	// Base query for counting total items
	countQuery := `
                SELECT COUNT(*)
                FROM borrows b
                LEFT JOIN users u ON b.user_id = u.id
                LEFT JOIN books bk ON b.book_id = bk.id
                WHERE 1=1
        `

	// Base query for fetching borrows with relations
	query := `
                SELECT b.id, b.user_id, b.book_id, b.status, b.borrow_date, b.due_date, b.return_date,
                                b.approved_by, b.created_at, b.updated_at
                FROM borrows b
                LEFT JOIN users u ON b.user_id = u.id
                LEFT JOIN books bk ON b.book_id = bk.id
                WHERE 1=1
        `

	// Parameters for query
	params := []interface{}{}
	paramCount := 1

	// Add search filter if provided
	if searchTerm != "" {
		searchFilter := fmt.Sprintf(" AND (bk.title ILIKE $%d OR u.name ILIKE $%d OR u.email ILIKE $%d OR u.student_id ILIKE $%d)",
			paramCount, paramCount, paramCount, paramCount)
		query += searchFilter
		countQuery += searchFilter
		params = append(params, "%"+searchTerm+"%")
		paramCount++
	}

	// Add status filter if provided
	if status != "" {
		statusFilter := fmt.Sprintf(" AND b.status = $%d", paramCount)
		query += statusFilter
		countQuery += statusFilter
		params = append(params, status)
		paramCount++
	}

	// Get total count
	var totalItems int
	err := db.QueryRow(countQuery, params...).Scan(&totalItems)
	if err != nil {
		return nil, 0, err
	}

	// Add order and pagination
	query += " ORDER BY b.created_at DESC"
	if page > 0 && itemsPerPage > 0 {
		offset := (page - 1) * itemsPerPage
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", paramCount, paramCount+1)
		params = append(params, itemsPerPage, offset)
	}

	// Execute query
	rows, err := db.Query(query, params...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	// Parse rows
	var borrows []*Borrow
	for rows.Next() {
		borrow := &Borrow{}
		err := rows.Scan(
			&borrow.ID,
			&borrow.UserID,
			&borrow.BookID,
			&borrow.Status,
			&borrow.BorrowDate,
			&borrow.DueDate,
			&borrow.ReturnDate,
			&borrow.ApprovedBy,
			&borrow.CreatedAt,
			&borrow.UpdatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Get related user
		borrow.User, _ = GetUserByID(borrow.UserID)

		// Get related book
		borrow.Book, _ = GetBookByID(borrow.BookID)

		// Get approver if exists
		if borrow.ApprovedBy != nil {
			borrow.Approver, _ = GetUserByID(*borrow.ApprovedBy)
		}

		// Calculate if overdue
		if borrow.Status == BorrowStatusApproved && borrow.DueDate != nil {
			if time.Now().After(*borrow.DueDate) {
				borrow.IsOverdue = true
			}
		}

		borrows = append(borrows, borrow)
	}

	// Check for errors
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return borrows, totalItems, nil
}

// GetBorrowHistory retrieves all borrow history
func GetBorrowHistory() ([]*Borrow, error) {
	db := config.GetDB()

	// Execute query
	rows, err := db.Query(`
                SELECT b.id, b.user_id, b.book_id, b.status, b.borrow_date, b.due_date, b.return_date, 
                        b.approved_by, b.created_at, b.updated_at
                FROM borrows b
                ORDER BY b.updated_at DESC
                LIMIT 100
        `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse rows
	var borrows []*Borrow
	for rows.Next() {
		borrow := &Borrow{}
		err := rows.Scan(
			&borrow.ID,
			&borrow.UserID,
			&borrow.BookID,
			&borrow.Status,
			&borrow.BorrowDate,
			&borrow.DueDate,
			&borrow.ReturnDate,
			&borrow.ApprovedBy,
			&borrow.CreatedAt,
			&borrow.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Get related user
		borrow.User, _ = GetUserByID(borrow.UserID)

		// Get related book
		borrow.Book, _ = GetBookByID(borrow.BookID)

		// Get approver if exists
		if borrow.ApprovedBy != nil {
			borrow.Approver, _ = GetUserByID(*borrow.ApprovedBy)
		}

		borrows = append(borrows, borrow)
	}

	// Check for errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return borrows, nil
}

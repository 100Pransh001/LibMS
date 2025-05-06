package models

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"library-management-system/config"
	"time"
)

// NullTime represents a time.Time that may be null.
// NullTime implements the sql.Scanner interface so
// it can be used as a scan destination:
//
//	var nt NullTime
//	err := db.QueryRow("SELECT fulfilled_date FROM reservations WHERE id = ?", id).Scan(&nt)
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if Time is not NULL
}

// Scan implements the Scanner interface.
func (nt *NullTime) Scan(value interface{}) error {
	if value == nil {
		nt.Time, nt.Valid = time.Time{}, false
		return nil
	}
	nt.Valid = true
	switch v := value.(type) {
	case time.Time:
		nt.Time = v
	default:
		return fmt.Errorf("cannot scan %T into NullTime", value)
	}
	return nil
}

// Value implements the driver Valuer interface.
func (nt NullTime) Value() (driver.Value, error) {
	if !nt.Valid {
		return nil, nil
	}
	return nt.Time, nil
}

// Reservation status constants
const (
	ReservationStatusActive    = "active"
	ReservationStatusFulfilled = "fulfilled"
	ReservationStatusCancelled = "cancelled"
	ReservationStatusExpired   = "expired"
)

// Reservation represents a book reservation
type Reservation struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	BookID          int       `json:"book_id"`
	Status          string    `json:"status"`
	ReservationDate time.Time `json:"reservation_date"`
	ExpiryDate      time.Time `json:"expiry_date"`
	FulfilledDate   time.Time `json:"fulfilled_date,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Helper field to check if FulfilledDate is valid
	HasFulfilledDate bool `json:"-"`

	// Relationships (populated as needed)
	User *User `json:"user,omitempty"`
	Book *Book `json:"book,omitempty"`
}

// ReserveBook creates a new reservation for a book
func ReserveBook(userID, bookID int) error {
	db := config.GetDB()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if book exists and is unavailable
	var available int
	err = tx.QueryRow("SELECT available FROM books WHERE id = $1", bookID).Scan(&available)
	if err != nil {
		return err
	}

	if available > 0 {
		return errors.New("this book is currently available and can be borrowed directly")
	}

	// Check if user already has an active reservation for this book
	var count int
	err = tx.QueryRow(`
                SELECT COUNT(*) FROM reservations 
                WHERE user_id = $1 AND book_id = $2 AND status = $3
        `, userID, bookID, ReservationStatusActive).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("you already have an active reservation for this book")
	}

	// Check if user is currently borrowing this book
	err = tx.QueryRow(`
                SELECT COUNT(*) FROM borrows 
                WHERE user_id = $1 AND book_id = $2 AND status = $3
        `, userID, bookID, BorrowStatusApproved).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("you are currently borrowing this book")
	}

	// Create reservation with expiry date 14 days from now
	expiryDate := time.Now().AddDate(0, 0, 14)
	_, err = tx.Exec(`
                INSERT INTO reservations (user_id, book_id, status, reservation_date, expiry_date)
                VALUES ($1, $2, $3, CURRENT_TIMESTAMP, $4)
        `, userID, bookID, ReservationStatusActive, expiryDate)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// CancelReservation cancels a reservation
func CancelReservation(id, userID int) error {
	db := config.GetDB()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if reservation exists and belongs to user
	var status string
	err = tx.QueryRow(`
                SELECT status FROM reservations 
                WHERE id = $1 AND user_id = $2
        `, id, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("reservation not found")
		}
		return err
	}

	// Check if reservation can be cancelled
	if status != ReservationStatusActive {
		return errors.New("only active reservations can be cancelled")
	}

	// Update reservation status
	_, err = tx.Exec(`
                UPDATE reservations
                SET status = $1, updated_at = CURRENT_TIMESTAMP
                WHERE id = $2
        `, ReservationStatusCancelled, id)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// ProcessReservationsForBook processes reservations when a book becomes available
func ProcessReservationsForBook(bookID int) error {
	db := config.GetDB()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Check if book is available
	var available int
	err = tx.QueryRow("SELECT available FROM books WHERE id = $1", bookID).Scan(&available)
	if err != nil {
		return err
	}

	if available <= 0 {
		// No copies available, nothing to do
		return nil
	}

	// Get the oldest active reservation for this book
	var reservationID, userID int
	err = tx.QueryRow(`
                SELECT id, user_id FROM reservations
                WHERE book_id = $1 AND status = $2
                ORDER BY reservation_date ASC
                LIMIT 1
        `, bookID, ReservationStatusActive).Scan(&reservationID, &userID)
	if err != nil {
		if err == sql.ErrNoRows {
			// No active reservations for this book
			return nil
		}
		return err
	}

	// Update reservation status to fulfilled
	_, err = tx.Exec(`
                UPDATE reservations
                SET status = $1, fulfilled_date = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP
                WHERE id = $2
        `, ReservationStatusFulfilled, reservationID)
	if err != nil {
		return err
	}

	// Create a pending borrow request for the user
	borrowDate := time.Now()
	dueDate := borrowDate.AddDate(0, 0, 14) // Due in 14 days
	_, err = tx.Exec(`
                INSERT INTO borrows (user_id, book_id, status, borrow_date, due_date)
                VALUES ($1, $2, $3, $4, $5)
        `, userID, bookID, BorrowStatusPending, borrowDate, dueDate)
	if err != nil {
		return err
	}

	// Commit transaction
	return tx.Commit()
}

// GetUserReservations gets all reservations for a user
func GetUserReservations(userID int) ([]*Reservation, error) {
	db := config.GetDB()

	// Execute query to get reservations
	rows, err := db.Query(`
                SELECT r.id, r.user_id, r.book_id, r.status, r.reservation_date, r.expiry_date, 
                       r.fulfilled_date, r.created_at, r.updated_at
                FROM reservations r
                WHERE r.user_id = $1
                ORDER BY r.reservation_date DESC
        `, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Parse rows
	var reservations []*Reservation
	for rows.Next() {
		reservation := &Reservation{}

		// Temporary variable for fulfilled_date
		var fulfilledDate NullTime

		err := rows.Scan(
			&reservation.ID,
			&reservation.UserID,
			&reservation.BookID,
			&reservation.Status,
			&reservation.ReservationDate,
			&reservation.ExpiryDate,
			&fulfilledDate,
			&reservation.CreatedAt,
			&reservation.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Set fulfilled date if valid
		if fulfilledDate.Valid {
			reservation.FulfilledDate = fulfilledDate.Time
			reservation.HasFulfilledDate = true
		}

		// Get related book
		reservation.Book, _ = GetBookByID(reservation.BookID)
		reservations = append(reservations, reservation)
	}

	// Check for errors
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reservations, nil
}

// CleanExpiredReservations moves expired reservations to expired status
func CleanExpiredReservations() error {
	db := config.GetDB()

	// Update reservations that have passed their expiry date
	_, err := db.Exec(`
                UPDATE reservations
                SET status = $1, updated_at = CURRENT_TIMESTAMP
                WHERE status = $2 AND expiry_date < CURRENT_TIMESTAMP
        `, ReservationStatusExpired, ReservationStatusActive)

	return err
}

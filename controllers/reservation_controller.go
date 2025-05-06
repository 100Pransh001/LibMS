package controllers

import (
	"library-management-system/middleware"
	"library-management-system/models"
	"library-management-system/utils"
	"net/http"
	"strconv"
	"strings"
)

// ReserveBook creates a new reservation for a book
func ReserveBook(w http.ResponseWriter, r *http.Request) {
	// Only POST method is allowed
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Only students can reserve books
	if !user.IsStudent {
		utils.SetError(w, r, "Only students can reserve books")
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}

	// Extract book ID from URL
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/books/")
	idStr = strings.TrimSuffix(idStr, "/reserve")

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.SetError(w, r, "Invalid book ID")
		http.Redirect(w, r, "/books", http.StatusSeeOther)
		return
	}

	// Reserve the book
	err = models.ReserveBook(user.ID, id)
	if err != nil {
		utils.SetError(w, r, "Error reserving book: "+err.Error())
		http.Redirect(w, r, "/books/"+idStr, http.StatusSeeOther)
		return
	}

	// Set success message and redirect
	utils.SetFlash(w, r, "Book reserved successfully. You will be notified when it becomes available.")
	http.Redirect(w, r, "/books/"+idStr, http.StatusSeeOther)
}

// CancelReservation cancels a reservation
func CancelReservation(w http.ResponseWriter, r *http.Request) {
	// Only POST method is allowed
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user from context
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Extract reservation ID from URL
	path := r.URL.Path
	idStr := strings.TrimPrefix(path, "/reservations/")
	idStr = strings.TrimSuffix(idStr, "/cancel")

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		utils.SetError(w, r, "Invalid reservation ID")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Cancel the reservation
	err = models.CancelReservation(id, user.ID)
	if err != nil {
		utils.SetError(w, r, "Error cancelling reservation: "+err.Error())
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Set success message and redirect
	utils.SetFlash(w, r, "Reservation cancelled successfully")
	http.Redirect(w, r, "/profile", http.StatusSeeOther)
}

// UserReservations displays the user's reservations
func UserReservations(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Only students can view their reservations
	if !user.IsStudent {
		utils.SetError(w, r, "Only students can view reservations")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Get user's reservations
	reservations, err := models.GetUserReservations(user.ID)
	if err != nil {
		utils.SetError(w, r, "Error fetching reservations: "+err.Error())
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Run cleanup for expired reservations
	go models.CleanExpiredReservations()

	// Prepare data for template
	data := &utils.TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Title":         "My Reservations",
			"Reservations":  reservations,
		},
	}

	// Render template
	utils.RenderTemplate(w, r, "user_reservations.html", data)
}
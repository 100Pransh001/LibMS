package controllers

import (
	"net/http"

	"library-management-system/middleware"
	"library-management-system/models"
	"library-management-system/utils"
)

// Home handles the home page
func Home(w http.ResponseWriter, r *http.Request) {
	// Check if path is exactly "/"
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Get user if authenticated
	user := middleware.GetUserFromContext(r)

	// Prepare data for template
	data := &utils.TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Title": "Home",
		},
	}

	// Get recent books
	recentBooks, err := getRecentBooks(10)
	if err == nil {
		data.Data["RecentBooks"] = recentBooks
	}

	// If user is authenticated, get personalized data
	if user != nil {
		if user.IsLibrarian {
			// For librarians, get pending borrow requests and overdue books
			pendingRequests, err := models.GetAllPendingBorrows()
			if err == nil {
				data.Data["PendingRequests"] = pendingRequests
				data.Data["PendingCount"] = len(pendingRequests)
			}

			overdueBooks, err := models.GetOverdueBooks()
			if err == nil {
				data.Data["OverdueBooks"] = overdueBooks
				data.Data["OverdueCount"] = len(overdueBooks)
			}
		} else {
			// For students, get active and pending borrows
			activeBorrows, err := models.GetActiveUserBorrows(user.ID)
			if err == nil {
				data.Data["ActiveBorrows"] = activeBorrows
				data.Data["ActiveCount"] = len(activeBorrows)
			}

			pendingBorrows, err := models.GetPendingUserBorrows(user.ID)
			if err == nil {
				data.Data["PendingBorrows"] = pendingBorrows
				data.Data["PendingCount"] = len(pendingBorrows)
			}
		}
	}

	// Render home template
	utils.RenderTemplate(w, r, "home.html", data)
}

// getRecentBooks returns a list of recent books
func getRecentBooks(limit int) ([]*models.Book, error) {
	// In a real application, we would implement this to fetch recent books
	// For now, let's just return all books
	return models.GetAllBooks()
}
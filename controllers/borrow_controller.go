package controllers

import (
        "net/http"
        "strconv"
        "strings"
        "time"

        "library-management-system/middleware"
        "library-management-system/models"
        "library-management-system/utils"
)

// BorrowList displays the list of borrow requests for librarians
func BorrowList(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can view borrow list
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to view this page")
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Get query parameters
        query := r.URL.Query()
        searchTerm := query.Get("search")
        status := query.Get("status")
        
        // Get page number from query string
        pageStr := query.Get("page")
        page := 1
        if pageStr != "" {
                var err error
                page, err = strconv.Atoi(pageStr)
                if err != nil || page < 1 {
                        page = 1
                }
        }
        
        // Items per page
        const itemsPerPage = 10
        
        // Get filtered borrow requests
        borrows, totalItems, err := models.GetBorrowsWithFilters(searchTerm, status, page, itemsPerPage)
        if err != nil {
                utils.SetError(w, r, "Error fetching borrow requests: "+err.Error())
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Calculate total pages
        totalPages := (totalItems + itemsPerPage - 1) / itemsPerPage
        if totalPages < 1 {
                totalPages = 1
        }
        
        // Prepare data for template
        data := &utils.TemplateData{
                User: user,
                Data: map[string]interface{}{
                        "title":      "Borrow Requests",
                        "borrows":    borrows,
                        "search":     searchTerm,
                        "status":     status,
                        "page":       page,
                        "totalPages": totalPages,
                        "now":        time.Now(),
                },
        }
        
        // Render template
        utils.RenderTemplate(w, r, "borrow_list.html", data)
}

// BorrowAction handles actions on borrow requests (approve/reject)
func BorrowAction(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can perform actions on borrow requests
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to perform this action")
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Process form submission
        if r.Method != http.MethodPost {
                utils.SetError(w, r, "Invalid request method")
                http.Redirect(w, r, "/borrows", http.StatusSeeOther)
                return
        }
        
        // Parse form
        err := r.ParseForm()
        if err != nil {
                utils.SetError(w, r, "Error processing form")
                http.Redirect(w, r, "/borrows", http.StatusSeeOther)
                return
        }
        
        // Extract borrow ID from URL
        path := r.URL.Path
        borrowIDStr := strings.TrimPrefix(path, "/borrows/")
        borrowIDStr = strings.TrimSuffix(borrowIDStr, "/action")
        
        // Get form values
        action := r.FormValue("action")
        dueDateStr := r.FormValue("due_date")
        
        // Validate form
        borrowID, err := strconv.Atoi(borrowIDStr)
        if err != nil || borrowID <= 0 {
                utils.SetError(w, r, "Invalid borrow ID")
                http.Redirect(w, r, "/borrows", http.StatusSeeOther)
                return
        }
        
        if action != "approve" && action != "reject" {
                utils.SetError(w, r, "Invalid action")
                http.Redirect(w, r, "/borrows", http.StatusSeeOther)
                return
        }
        
        // Perform action
        if action == "approve" {
                // Validate due date
                if dueDateStr == "" {
                        utils.SetError(w, r, "Please provide a due date")
                        http.Redirect(w, r, "/borrows", http.StatusSeeOther)
                        return
                }
                
                // Parse due date
                dueDate, err := time.Parse("2006-01-02", dueDateStr)
                if err != nil {
                        utils.SetError(w, r, "Invalid due date format")
                        http.Redirect(w, r, "/borrows", http.StatusSeeOther)
                        return
                }
                
                // Check if due date is in the future
                if dueDate.Before(time.Now()) {
                        utils.SetError(w, r, "Due date must be in the future")
                        http.Redirect(w, r, "/borrows", http.StatusSeeOther)
                        return
                }
                
                // Approve borrow request
                err = models.ApproveBorrow(borrowID, user.ID, dueDate)
                if err != nil {
                        utils.SetError(w, r, "Error approving borrow request: "+err.Error())
                        http.Redirect(w, r, "/borrows", http.StatusSeeOther)
                        return
                }
                
                utils.SetFlash(w, r, "Borrow request approved successfully")
        } else {
                // Reject borrow request
                err = models.RejectBorrow(borrowID, user.ID)
                if err != nil {
                        utils.SetError(w, r, "Error rejecting borrow request: "+err.Error())
                        http.Redirect(w, r, "/borrows", http.StatusSeeOther)
                        return
                }
                
                utils.SetFlash(w, r, "Borrow request rejected successfully")
        }
        
        // Redirect back to borrow list
        http.Redirect(w, r, "/borrows", http.StatusSeeOther)
}

// ReturnBook handles returning a borrowed book
func ReturnBook(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Extract borrow ID from URL
        path := r.URL.Path
        idStr := strings.TrimPrefix(path, "/borrows/")
        idStr = strings.TrimSuffix(idStr, "/return")
        
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
                utils.SetError(w, r, "Invalid borrow ID")
                http.Redirect(w, r, "/profile", http.StatusSeeOther)
                return
        }
        
        // Get borrow record
        borrow, err := models.GetBorrowByID(id)
        if err != nil {
                utils.SetError(w, r, "Borrow record not found")
                http.Redirect(w, r, "/profile", http.StatusSeeOther)
                return
        }
        
        // Check if the current user is the borrower or a librarian
        if borrow.UserID != user.ID && !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to return this book")
                http.Redirect(w, r, "/profile", http.StatusSeeOther)
                return
        }
        
        // Check if the book is currently borrowed (status is approved)
        if borrow.Status != models.BorrowStatusApproved {
                utils.SetError(w, r, "This book is not currently borrowed")
                
                if user.IsLibrarian {
                        http.Redirect(w, r, "/borrows", http.StatusSeeOther)
                } else {
                        http.Redirect(w, r, "/profile", http.StatusSeeOther)
                }
                return
        }
        
        // Return the book
        err = models.ReturnBook(id)
        if err != nil {
                utils.SetError(w, r, "Error returning book: "+err.Error())
                
                if user.IsLibrarian {
                        http.Redirect(w, r, "/borrows", http.StatusSeeOther)
                } else {
                        http.Redirect(w, r, "/profile", http.StatusSeeOther)
                }
                return
        }
        
        utils.SetFlash(w, r, "Book returned successfully")
        
        // Redirect based on user role
        if user.IsLibrarian {
                http.Redirect(w, r, "/borrows", http.StatusSeeOther)
        } else {
                http.Redirect(w, r, "/profile", http.StatusSeeOther)
        }
}

// BorrowHistory displays the borrow history for librarians
func BorrowHistory(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can view borrow history
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to view this page")
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Get borrow history
        borrows, err := models.GetBorrowHistory()
        if err != nil {
                utils.SetError(w, r, "Error fetching borrow history: "+err.Error())
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Prepare data for template
        data := &utils.TemplateData{
                User: user,
                Data: map[string]interface{}{
                        "Title":   "Borrow History",
                        "Borrows": borrows,
                },
        }
        
        // Render template
        utils.RenderTemplate(w, r, "borrow_history.html", data)
}

// BorrowReport displays the borrow report for librarians
func BorrowReport(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can view reports
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to view this page")
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Get active borrows
        activeBorrows, err := models.GetActiveBorrows()
        if err != nil {
                utils.SetError(w, r, "Error fetching active borrows: "+err.Error())
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Get overdue borrows
        overdueBorrows, err := models.GetOverdueBooks()
        if err != nil {
                utils.SetError(w, r, "Error fetching overdue borrows: "+err.Error())
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Count pending requests
        pendingBorrows, err := models.GetAllPendingBorrows()
        if err != nil {
                utils.SetError(w, r, "Error fetching pending borrows: "+err.Error())
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Prepare data for template
        data := &utils.TemplateData{
                User: user,
                Data: map[string]interface{}{
                        "Title":             "Borrow Report",
                        "ActiveBorrows":     activeBorrows,
                        "ActiveCount":       len(activeBorrows),
                        "OverdueBorrows":    overdueBorrows,
                        "OverdueCount":      len(overdueBorrows),
                        "PendingBorrows":    pendingBorrows,
                        "PendingCount":      len(pendingBorrows),
                        "TotalActive":       len(activeBorrows),
                        "TotalOverdue":      len(overdueBorrows),
                        "OverduePercentage": calculatePercentage(len(overdueBorrows), len(activeBorrows)),
                },
        }
        
        // Render template
        utils.RenderTemplate(w, r, "borrow_report.html", data)
}

// calculatePercentage calculates percentage of part / total
func calculatePercentage(part, total int) float64 {
        if total == 0 {
                return 0
        }
        return float64(part) / float64(total) * 100
}
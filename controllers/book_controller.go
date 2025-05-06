package controllers

import (
        "database/sql"
        "net/http"
        "strconv"
        "strings"

        "library-management-system/middleware"
        "library-management-system/models"
        "library-management-system/utils"
)

// BookList displays the list of books
func BookList(w http.ResponseWriter, r *http.Request) {
        // Get the current user if authenticated
        user := middleware.GetUserFromContext(r)

        // Get search and filter parameters
        query := r.URL.Query()
        search := query.Get("search")
        searchBy := query.Get("searchBy")
        
        // Get pagination parameters
        page, err := strconv.Atoi(query.Get("page"))
        if err != nil || page < 1 {
                page = 1
        }
        
        // Get books based on search criteria
        books, err := models.GetBooks(search, searchBy, page)
        if err != nil {
                utils.SetError(w, r, "Error fetching books: "+err.Error())
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Get total books for pagination
        totalBooks, err := models.CountBooks(search, searchBy)
        if err != nil {
                utils.SetError(w, r, "Error counting books: "+err.Error())
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Calculate pagination
        pageSize := 10
        totalPages := (totalBooks + pageSize - 1) / pageSize
        
        // Prepare data for template
        data := &utils.TemplateData{
                User: user,
                Data: map[string]interface{}{
                        "Title":      "Book List",
                        "Books":      books,
                        "Page":       page,
                        "TotalPages": totalPages,
                        "Search":     search,
                        "SearchBy":   searchBy,
                },
        }
        
        // Render template
        utils.RenderTemplate(w, r, "book_list.html", data)
}

// BookDetail displays details of a specific book
func BookDetail(w http.ResponseWriter, r *http.Request) {
        // Get the current user if authenticated
        user := middleware.GetUserFromContext(r)
        
        // Extract book ID from URL
        path := r.URL.Path
        idStr := strings.TrimPrefix(path, "/books/")
        if strings.Contains(idStr, "/") {
                idStr = idStr[:strings.Index(idStr, "/")]
        }
        
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
                utils.SetError(w, r, "Invalid book ID")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        // Get book details
        book, err := models.GetBookByID(id)
        if err != nil {
                utils.SetError(w, r, "Book not found")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        data := &utils.TemplateData{
                User: user,
                Data: map[string]interface{}{
                        "Title": book.Title,
                        "Book":  book,
                },
        }
        
        // If user is authenticated, check borrow status
        if user != nil {
                // Check if user has a pending borrow request for this book
                hasPendingRequest, err := models.HasPendingBorrowRequest(user.ID, id)
                if err == nil {
                        data.Data["HasPendingRequest"] = hasPendingRequest
                }
                
                // Check if user is currently borrowing this book
                currentBorrow, err := models.GetCurrentBorrow(user.ID, id)
                if err == nil && currentBorrow != nil {
                        data.Data["IsCurrentlyBorrowing"] = true
                        data.Data["CurrentBorrow"] = currentBorrow
                } else {
                        data.Data["IsCurrentlyBorrowing"] = false
                }
                
                // Check if user has an active reservation for this book
                reservations, err := models.GetUserReservations(user.ID)
                if err == nil {
                        for _, reservation := range reservations {
                                if reservation.BookID == id && reservation.Status == models.ReservationStatusActive {
                                        data.Data["HasActiveReservation"] = true
                                        data.Data["Reservation"] = reservation
                                        break
                                }
                        }
                }
        }
        
        // Render template
        utils.RenderTemplate(w, r, "book_detail.html", data)
}

// AddBook displays the form for adding a new book (GET) or processes the form submission (POST)
func AddBook(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can add books
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to add books")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        // Process form submission
        if r.Method == http.MethodPost {
                // Parse form
                err := r.ParseForm()
                if err != nil {
                        utils.SetError(w, r, "Error processing form")
                        utils.RenderTemplate(w, r, "book_form.html", &utils.TemplateData{User: user})
                        return
                }
                
                // Get form values
                title := r.FormValue("title")
                author := r.FormValue("author")
                isbn := r.FormValue("isbn")
                publisher := r.FormValue("publisher")
                pubYearStr := r.FormValue("publication_year")
                category := r.FormValue("category")
                description := r.FormValue("description")
                quantityStr := r.FormValue("quantity")
                
                // Validate form
                if title == "" || author == "" || isbn == "" || quantityStr == "" {
                        utils.SetError(w, r, "Please fill in all required fields")
                        utils.RenderTemplate(w, r, "book_form.html", &utils.TemplateData{User: user})
                        return
                }
                
                // Convert numeric values
                pubYear, _ := strconv.Atoi(pubYearStr)
                quantity, err := strconv.Atoi(quantityStr)
                if err != nil || quantity <= 0 {
                        utils.SetError(w, r, "Quantity must be a positive number")
                        utils.RenderTemplate(w, r, "book_form.html", &utils.TemplateData{User: user})
                        return
                }
                
                // Check if ISBN already exists
                exists, err := models.IsbnExists(isbn)
                if err != nil {
                        utils.SetError(w, r, "Error checking ISBN: "+err.Error())
                        utils.RenderTemplate(w, r, "book_form.html", &utils.TemplateData{User: user})
                        return
                }
                if exists {
                        utils.SetError(w, r, "A book with this ISBN already exists")
                        utils.RenderTemplate(w, r, "book_form.html", &utils.TemplateData{User: user})
                        return
                }
                
                // Create new book
                book := &models.Book{
                        Title:           title,
                        Author:          author,
                        ISBN:            isbn,
                        Publisher:       publisher,
                        PublicationYear: pubYear,
                        Category:        category,
                        Description:     description,
                        Quantity:        quantity,
                        Available:       quantity,
                        AddedBy:         sql.NullInt64{Int64: int64(user.ID), Valid: true},
                }
                
                // Save book to database
                err = book.Create()
                if err != nil {
                        utils.SetError(w, r, "Error adding book: "+err.Error())
                        utils.RenderTemplate(w, r, "book_form.html", &utils.TemplateData{User: user})
                        return
                }
                
                // Set flash message and redirect
                utils.SetFlash(w, r, "Book added successfully")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        // Display form for GET request
        data := &utils.TemplateData{
                User: user,
                Data: map[string]interface{}{
                        "Title": "Add Book",
                },
        }
        
        // Render template
        utils.RenderTemplate(w, r, "book_form.html", data)
}

// EditBook displays the form for editing a book (GET) or processes the form submission (POST)
func EditBook(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can edit books
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to edit books")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        // Extract book ID from URL
        path := r.URL.Path
        idStr := strings.TrimPrefix(path, "/books/")
        idStr = strings.TrimSuffix(idStr, "/edit")
        
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
                utils.SetError(w, r, "Invalid book ID")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        // Get book details
        book, err := models.GetBookByID(id)
        if err != nil {
                utils.SetError(w, r, "Book not found")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        // Process form submission
        if r.Method == http.MethodPost {
                // Parse form
                err := r.ParseForm()
                if err != nil {
                        utils.SetError(w, r, "Error processing form")
                        data := &utils.TemplateData{
                                User: user,
                                Data: map[string]interface{}{
                                        "Title": "Edit Book",
                                        "Book":  book,
                                },
                        }
                        utils.RenderTemplate(w, r, "book_form.html", data)
                        return
                }
                
                // Get form values
                title := r.FormValue("title")
                author := r.FormValue("author")
                isbn := r.FormValue("isbn")
                publisher := r.FormValue("publisher")
                pubYearStr := r.FormValue("publication_year")
                category := r.FormValue("category")
                description := r.FormValue("description")
                quantityStr := r.FormValue("quantity")
                
                // Validate form
                if title == "" || author == "" || isbn == "" || quantityStr == "" {
                        utils.SetError(w, r, "Please fill in all required fields")
                        data := &utils.TemplateData{
                                User: user,
                                Data: map[string]interface{}{
                                        "Title": "Edit Book",
                                        "Book":  book,
                                },
                        }
                        utils.RenderTemplate(w, r, "book_form.html", data)
                        return
                }
                
                // Convert numeric values
                pubYear, _ := strconv.Atoi(pubYearStr)
                quantity, err := strconv.Atoi(quantityStr)
                if err != nil || quantity <= 0 {
                        utils.SetError(w, r, "Quantity must be a positive number")
                        data := &utils.TemplateData{
                                User: user,
                                Data: map[string]interface{}{
                                        "Title": "Edit Book",
                                        "Book":  book,
                                },
                        }
                        utils.RenderTemplate(w, r, "book_form.html", data)
                        return
                }
                
                // Check if ISBN already exists and belongs to a different book
                if isbn != book.ISBN {
                        exists, err := models.IsbnExistsExcept(isbn, id)
                        if err != nil {
                                utils.SetError(w, r, "Error checking ISBN: "+err.Error())
                                data := &utils.TemplateData{
                                        User: user,
                                        Data: map[string]interface{}{
                                                "Title": "Edit Book",
                                                "Book":  book,
                                        },
                                }
                                utils.RenderTemplate(w, r, "book_form.html", data)
                                return
                        }
                        if exists {
                                utils.SetError(w, r, "A book with this ISBN already exists")
                                data := &utils.TemplateData{
                                        User: user,
                                        Data: map[string]interface{}{
                                                "Title": "Edit Book",
                                                "Book":  book,
                                        },
                                }
                                utils.RenderTemplate(w, r, "book_form.html", data)
                                return
                        }
                }
                
                // Calculate new available copies
                availableDiff := quantity - book.Quantity
                newAvailable := book.Available + availableDiff
                if newAvailable < 0 {
                        newAvailable = 0
                }
                
                // Update book
                book.Title = title
                book.Author = author
                book.ISBN = isbn
                book.Publisher = publisher
                book.PublicationYear = pubYear
                book.Category = category
                book.Description = description
                book.Quantity = quantity
                book.Available = newAvailable
                
                // Save changes to database
                err = book.Update()
                if err != nil {
                        utils.SetError(w, r, "Error updating book: "+err.Error())
                        data := &utils.TemplateData{
                                User: user,
                                Data: map[string]interface{}{
                                        "Title": "Edit Book",
                                        "Book":  book,
                                },
                        }
                        utils.RenderTemplate(w, r, "book_form.html", data)
                        return
                }
                
                // Set flash message and redirect
                utils.SetFlash(w, r, "Book updated successfully")
                http.Redirect(w, r, "/books/"+strconv.Itoa(id), http.StatusSeeOther)
                return
        }
        
        // Display form for GET request
        data := &utils.TemplateData{
                User: user,
                Data: map[string]interface{}{
                        "Title": "Edit Book",
                        "Book":  book,
                },
        }
        
        // Render template
        utils.RenderTemplate(w, r, "book_form.html", data)
}

// DeleteBook handles deletion of a book
func DeleteBook(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can delete books
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to delete books")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        // Extract book ID from URL
        path := r.URL.Path
        idStr := strings.TrimPrefix(path, "/books/")
        idStr = strings.TrimSuffix(idStr, "/delete")
        
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
                utils.SetError(w, r, "Invalid book ID")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        // Get book details
        book, err := models.GetBookByID(id)
        if err != nil {
                utils.SetError(w, r, "Book not found")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        // Check if book is currently borrowed
        active, err := models.HasActiveOrPendingBorrows(id)
        if err != nil {
                utils.SetError(w, r, "Error checking borrow status: "+err.Error())
                http.Redirect(w, r, "/books/"+strconv.Itoa(id), http.StatusSeeOther)
                return
        }
        if active {
                utils.SetError(w, r, "Cannot delete book as it is currently borrowed or has pending borrow requests")
                http.Redirect(w, r, "/books/"+strconv.Itoa(id), http.StatusSeeOther)
                return
        }
        
        // Delete book
        err = book.Delete()
        if err != nil {
                utils.SetError(w, r, "Error deleting book: "+err.Error())
                http.Redirect(w, r, "/books/"+strconv.Itoa(id), http.StatusSeeOther)
                return
        }
        
        // Set flash message and redirect
        utils.SetFlash(w, r, "Book deleted successfully")
        http.Redirect(w, r, "/books", http.StatusSeeOther)
}

// BorrowBook handles borrowing a book
func BorrowBook(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Extract book ID from URL
        path := r.URL.Path
        idStr := strings.TrimPrefix(path, "/books/")
        idStr = strings.TrimSuffix(idStr, "/borrow")
        
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
                utils.SetError(w, r, "Invalid book ID")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        // Get book details
        book, err := models.GetBookByID(id)
        if err != nil {
                utils.SetError(w, r, "Book not found")
                http.Redirect(w, r, "/books", http.StatusSeeOther)
                return
        }
        
        // Check if book is available
        if book.Available <= 0 {
                utils.SetError(w, r, "No copies available for borrowing")
                http.Redirect(w, r, "/books/"+strconv.Itoa(id), http.StatusSeeOther)
                return
        }
        
        // Check if user has a pending request for this book
        hasPending, err := models.HasPendingBorrowRequest(user.ID, id)
        if err != nil {
                utils.SetError(w, r, "Error checking borrow status: "+err.Error())
                http.Redirect(w, r, "/books/"+strconv.Itoa(id), http.StatusSeeOther)
                return
        }
        if hasPending {
                utils.SetError(w, r, "You already have a pending request for this book")
                http.Redirect(w, r, "/books/"+strconv.Itoa(id), http.StatusSeeOther)
                return
        }
        
        // Check if user is currently borrowing this book
        isBorrowing, err := models.IsCurrentlyBorrowing(user.ID, id)
        if err != nil {
                utils.SetError(w, r, "Error checking borrow status: "+err.Error())
                http.Redirect(w, r, "/books/"+strconv.Itoa(id), http.StatusSeeOther)
                return
        }
        if isBorrowing {
                utils.SetError(w, r, "You are already borrowing this book")
                http.Redirect(w, r, "/books/"+strconv.Itoa(id), http.StatusSeeOther)
                return
        }
        
        // Create borrow request
        err = models.CreateBorrowRequest(user.ID, id)
        if err != nil {
                utils.SetError(w, r, "Error creating borrow request: "+err.Error())
                http.Redirect(w, r, "/books/"+strconv.Itoa(id), http.StatusSeeOther)
                return
        }
        
        // Set flash message and redirect
        utils.SetFlash(w, r, "Borrow request submitted successfully")
        http.Redirect(w, r, "/books/"+strconv.Itoa(id), http.StatusSeeOther)
}

// BookReport displays the book report for librarians
func BookReport(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can view reports
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to view reports")
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Get report data
        totalBooks, err := models.CountAllBooks()
        if err != nil {
                utils.SetError(w, r, "Error generating report: "+err.Error())
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        availableBooks, err := models.CountAvailableBooks()
        if err != nil {
                utils.SetError(w, r, "Error generating report: "+err.Error())
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        topBorrowedBooks, err := models.GetTopBorrowedBooks(10)
        if err != nil {
                utils.SetError(w, r, "Error generating report: "+err.Error())
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Prepare data for template
        data := &utils.TemplateData{
                User: user,
                Data: map[string]interface{}{
                        "Title":            "Book Report",
                        "TotalBooks":       totalBooks,
                        "AvailableBooks":   availableBooks,
                        "BorrowedBooks":    totalBooks - availableBooks,
                        "TopBorrowedBooks": topBorrowedBooks,
                },
        }
        
        // Render template
        utils.RenderTemplate(w, r, "book_report.html", data)
}
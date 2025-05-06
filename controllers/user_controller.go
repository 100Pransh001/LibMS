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

// UserList displays the list of users for librarians
func UserList(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can view user list
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to view this page")
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Get query parameters
        query := r.URL.Query()
        searchTerm := query.Get("search")
        role := query.Get("role")
        
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
        
        // Get users (TODO: implement filtered version with pagination)
        users, err := models.GetAllUsers()
        if err != nil {
                utils.SetError(w, r, "Error fetching users: "+err.Error())
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Calculate total pages (mock for now)
        totalPages := 1
        
        // Prepare data for template
        data := &utils.TemplateData{
                User: user,
                Data: map[string]interface{}{
                        "Title":      "User List",
                        "Users":      users,
                        "Search":     searchTerm,
                        "Role":       role,
                        "Page":       page,
                        "TotalPages": totalPages,
                },
        }
        
        // Render template
        utils.RenderTemplate(w, r, "user_list.html", data)
}

// AddUser displays the form to add a new user (GET) or processes the form (POST)
func AddUser(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can add users
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to add users")
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Process form submission
        if r.Method == http.MethodPost {
                // Parse form
                err := r.ParseForm()
                if err != nil {
                        utils.SetError(w, r, "Error processing form")
                        utils.RenderTemplate(w, r, "user_form.html", &utils.TemplateData{User: user})
                        return
                }
                
                // Get form values
                name := r.FormValue("name")
                email := r.FormValue("email")
                password := r.FormValue("password")
                confirmPassword := r.FormValue("confirm_password")
                role := r.FormValue("role")
                studentID := r.FormValue("student_id")
                phone := r.FormValue("phone")
                
                // Validate form
                if name == "" || email == "" || password == "" || role == "" {
                        utils.SetError(w, r, "Please fill in all required fields")
                        utils.RenderTemplate(w, r, "user_form.html", &utils.TemplateData{User: user})
                        return
                }
                
                // Validate password
                if password != confirmPassword {
                        utils.SetError(w, r, "Passwords do not match")
                        utils.RenderTemplate(w, r, "user_form.html", &utils.TemplateData{User: user})
                        return
                }
                
                // Validate role
                if role != "librarian" && role != "student" {
                        utils.SetError(w, r, "Invalid role")
                        utils.RenderTemplate(w, r, "user_form.html", &utils.TemplateData{User: user})
                        return
                }
                
                // Create new user
                newUser := &models.User{
                        Name:      name,
                        Email:     email,
                        Password:  password,
                        Role:      role,
                        StudentID: sql.NullString{String: studentID, Valid: studentID != ""},
                        Phone:     sql.NullString{String: phone, Valid: phone != ""},
                }
                
                // Save user to database
                err = newUser.Create()
                if err != nil {
                        if err == models.ErrDuplicateEmail {
                                utils.SetError(w, r, "Email already exists")
                        } else {
                                utils.SetError(w, r, "Error creating user: "+err.Error())
                        }
                        utils.RenderTemplate(w, r, "user_form.html", &utils.TemplateData{User: user})
                        return
                }
                
                // Set flash message and redirect
                utils.SetFlash(w, r, "User added successfully")
                http.Redirect(w, r, "/users", http.StatusSeeOther)
                return
        }
        
        // Display form for GET request
        data := &utils.TemplateData{
                User: user,
                Data: map[string]interface{}{
                        "Title": "Add User",
                },
        }
        
        // Render template
        utils.RenderTemplate(w, r, "user_form.html", data)
}

// EditUser displays the form to edit a user (GET) or processes the form (POST)
func EditUser(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can edit users
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to edit users")
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Extract user ID from URL
        path := r.URL.Path
        idStr := strings.TrimPrefix(path, "/users/edit/")
        
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
                utils.SetError(w, r, "Invalid user ID")
                http.Redirect(w, r, "/users", http.StatusSeeOther)
                return
        }
        
        // Get user to edit
        editUser, err := models.GetUserByID(id)
        if err != nil {
                utils.SetError(w, r, "User not found")
                http.Redirect(w, r, "/users", http.StatusSeeOther)
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
                                        "Title":    "Edit User",
                                        "EditUser": editUser,
                                },
                        }
                        utils.RenderTemplate(w, r, "user_form.html", data)
                        return
                }
                
                // Get form values
                name := r.FormValue("name")
                email := r.FormValue("email")
                role := r.FormValue("role")
                studentID := r.FormValue("student_id")
                phone := r.FormValue("phone")
                
                // Validate form
                if name == "" || email == "" || role == "" {
                        utils.SetError(w, r, "Please fill in all required fields")
                        data := &utils.TemplateData{
                                User: user,
                                Data: map[string]interface{}{
                                        "Title":    "Edit User",
                                        "EditUser": editUser,
                                },
                        }
                        utils.RenderTemplate(w, r, "user_form.html", data)
                        return
                }
                
                // Validate role
                if role != "librarian" && role != "student" {
                        utils.SetError(w, r, "Invalid role")
                        data := &utils.TemplateData{
                                User: user,
                                Data: map[string]interface{}{
                                        "Title":    "Edit User",
                                        "EditUser": editUser,
                                },
                        }
                        utils.RenderTemplate(w, r, "user_form.html", data)
                        return
                }
                
                // Update user
                editUser.Name = name
                editUser.Email = email
                editUser.Role = role
                editUser.StudentID = sql.NullString{String: studentID, Valid: studentID != ""}
                editUser.Phone = sql.NullString{String: phone, Valid: phone != ""}
                
                // Save changes to database
                err = editUser.Update()
                if err != nil {
                        if err == models.ErrDuplicateEmail {
                                utils.SetError(w, r, "Email already exists")
                        } else {
                                utils.SetError(w, r, "Error updating user: "+err.Error())
                        }
                        data := &utils.TemplateData{
                                User: user,
                                Data: map[string]interface{}{
                                        "Title":    "Edit User",
                                        "EditUser": editUser,
                                },
                        }
                        utils.RenderTemplate(w, r, "user_form.html", data)
                        return
                }
                
                // Set flash message and redirect
                utils.SetFlash(w, r, "User updated successfully")
                http.Redirect(w, r, "/users", http.StatusSeeOther)
                return
        }
        
        // Display form for GET request
        data := &utils.TemplateData{
                User: user,
                Data: map[string]interface{}{
                        "Title":    "Edit User",
                        "EditUser": editUser,
                },
        }
        
        // Render template
        utils.RenderTemplate(w, r, "user_form.html", data)
}

// DeleteUser handles deleting a user
func DeleteUser(w http.ResponseWriter, r *http.Request) {
        // Get user from context
        user := middleware.GetUserFromContext(r)
        if user == nil {
                http.Redirect(w, r, "/login", http.StatusSeeOther)
                return
        }
        
        // Only librarians can delete users
        if !user.IsLibrarian {
                utils.SetError(w, r, "You do not have permission to delete users")
                http.Redirect(w, r, "/", http.StatusSeeOther)
                return
        }
        
        // Extract user ID from URL
        path := r.URL.Path
        idStr := strings.TrimPrefix(path, "/users/delete/")
        
        id, err := strconv.Atoi(idStr)
        if err != nil || id <= 0 {
                utils.SetError(w, r, "Invalid user ID")
                http.Redirect(w, r, "/users", http.StatusSeeOther)
                return
        }
        
        // Get user to delete
        deleteUser, err := models.GetUserByID(id)
        if err != nil {
                utils.SetError(w, r, "User not found")
                http.Redirect(w, r, "/users", http.StatusSeeOther)
                return
        }
        
        // Prevent deleting self
        if deleteUser.ID == user.ID {
                utils.SetError(w, r, "You cannot delete your own account")
                http.Redirect(w, r, "/users", http.StatusSeeOther)
                return
        }
        
        // Check if there are active or pending borrows for this user
        activeBorrows, err := models.GetActiveUserBorrows(deleteUser.ID)
        if err != nil {
                utils.SetError(w, r, "Error checking user borrows: "+err.Error())
                http.Redirect(w, r, "/users", http.StatusSeeOther)
                return
        }
        
        if len(activeBorrows) > 0 {
                utils.SetError(w, r, "Cannot delete user with active borrows")
                http.Redirect(w, r, "/users", http.StatusSeeOther)
                return
        }
        
        pendingBorrows, err := models.GetPendingUserBorrows(deleteUser.ID)
        if err != nil {
                utils.SetError(w, r, "Error checking user borrows: "+err.Error())
                http.Redirect(w, r, "/users", http.StatusSeeOther)
                return
        }
        
        if len(pendingBorrows) > 0 {
                utils.SetError(w, r, "Cannot delete user with pending borrow requests")
                http.Redirect(w, r, "/users", http.StatusSeeOther)
                return
        }
        
        // Delete user
        err = deleteUser.Delete()
        if err != nil {
                utils.SetError(w, r, "Error deleting user: "+err.Error())
                http.Redirect(w, r, "/users", http.StatusSeeOther)
                return
        }
        
        // Set flash message and redirect
        utils.SetFlash(w, r, "User deleted successfully")
        http.Redirect(w, r, "/users", http.StatusSeeOther)
}
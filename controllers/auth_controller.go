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

// Login handles the main login selection page
func Login(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to home
	if userID := utils.GetSessionInt(r, "user_id"); userID > 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := &utils.TemplateData{
		Data: map[string]interface{}{
			"Title": "Login",
		},
	}

	// Get redirect parameter
	redirect := r.URL.Query().Get("redirect")
	if redirect != "" {
		data.Data["Redirect"] = redirect
	}

	// Render login selection template
	utils.RenderTemplate(w, r, "login_selection.html", data)
}

// StudentLogin handles student login
func StudentLogin(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to home
	if userID := utils.GetSessionInt(r, "user_id"); userID > 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := &utils.TemplateData{
		Data: map[string]interface{}{
			"Title": "Student Login",
		},
	}

	// Process login form submission
	if r.Method == http.MethodPost {
		// Parse form
		err := r.ParseForm()
		if err != nil {
			utils.SetError(w, r, "Error processing form")
			utils.RenderTemplate(w, r, "student_login.html", data)
			return
		}

		// Get form values
		email := r.FormValue("email")
		password := r.FormValue("password")
		redirect := r.FormValue("redirect")

		// Validate form
		if email == "" || password == "" {
			utils.SetError(w, r, "Please enter both email and password")
			utils.RenderTemplate(w, r, "student_login.html", data)
			return
		}

		// Authenticate user
		user, err := models.Authenticate(email, password)
		if err != nil {
			utils.SetError(w, r, "Invalid email or password")
			utils.RenderTemplate(w, r, "student_login.html", data)
			return
		}

		// Check if user is a student
		if !user.IsStudent {
			utils.SetError(w, r, "This login is for students only. Please use the librarian login.")
			utils.RenderTemplate(w, r, "student_login.html", data)
			return
		}

		// Set session
		utils.SetSession(w, r, "user_id", user.ID)
		utils.SetSession(w, r, "user_role", user.Role)
		utils.SetFlash(w, r, "You have successfully logged in")

		// Redirect to requested page or home
		if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		return
	}

	// Get redirect parameter
	redirect := r.URL.Query().Get("redirect")
	if redirect != "" {
		data.Data["Redirect"] = redirect
	}

	// Render student login template
	utils.RenderTemplate(w, r, "student_login.html", data)
}

// LibrarianLogin handles librarian login
func LibrarianLogin(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to home
	if userID := utils.GetSessionInt(r, "user_id"); userID > 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := &utils.TemplateData{
		Data: map[string]interface{}{
			"Title": "Librarian Login",
		},
	}

	// Process login form submission
	if r.Method == http.MethodPost {
		// Parse form
		err := r.ParseForm()
		if err != nil {
			utils.SetError(w, r, "Error processing form")
			utils.RenderTemplate(w, r, "librarian_login.html", data)
			return
		}

		// Get form values
		email := r.FormValue("email")
		password := r.FormValue("password")
		redirect := r.FormValue("redirect")

		// Validate form
		if email == "" || password == "" {
			utils.SetError(w, r, "Please enter both email and password")
			utils.RenderTemplate(w, r, "librarian_login.html", data)
			return
		}

		// Authenticate user
		user, err := models.Authenticate(email, password)
		if err != nil {
			utils.SetError(w, r, "Invalid email or password")
			utils.RenderTemplate(w, r, "librarian_login.html", data)
			return
		}

		// Check if user is a librarian
		if !user.IsLibrarian {
			utils.SetError(w, r, "This login is for librarians only. Please use the student login.")
			utils.RenderTemplate(w, r, "librarian_login.html", data)
			return
		}

		// Set session
		utils.SetSession(w, r, "user_id", user.ID)
		utils.SetSession(w, r, "user_role", user.Role)
		utils.SetFlash(w, r, "You have successfully logged in")

		// Redirect to requested page or home
		if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusSeeOther)
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		}
		return
	}

	// Get redirect parameter
	redirect := r.URL.Query().Get("redirect")
	if redirect != "" {
		data.Data["Redirect"] = redirect
	}

	// Render librarian login template
	utils.RenderTemplate(w, r, "librarian_login.html", data)
}

// Register handles user registration
func Register(w http.ResponseWriter, r *http.Request) {
	// If user is already logged in, redirect to home
	if userID := utils.GetSessionInt(r, "user_id"); userID > 0 {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := &utils.TemplateData{
		Data: map[string]interface{}{
			"Title": "Register",
		},
	}

	// Process registration form submission
	if r.Method == http.MethodPost {
		// Parse form
		err := r.ParseForm()
		if err != nil {
			utils.SetError(w, r, "Error processing form")
			utils.RenderTemplate(w, r, "register.html", data)
			return
		}

		// Get form values
		name := r.FormValue("name")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")
		studentID := r.FormValue("student_id")
		phone := r.FormValue("phone")

		// Validate form
		if name == "" || email == "" || password == "" {
			utils.SetError(w, r, "Please fill in all required fields")
			utils.RenderTemplate(w, r, "register.html", data)
			return
		}

		if password != confirmPassword {
			utils.SetError(w, r, "Passwords do not match")
			utils.RenderTemplate(w, r, "register.html", data)
			return
		}

		// Create new user
		user := &models.User{
			Name:      name,
			Email:     email,
			Password:  password,
			Role:      "student", // Default role is student
			StudentID: sql.NullString{String: studentID, Valid: studentID != ""},
			Phone:     sql.NullString{String: phone, Valid: phone != ""},
		}

		// Save user to database
		err = user.Create()
		if err != nil {
			if err == models.ErrDuplicateEmail {
				utils.SetError(w, r, "Email already exists")
			} else {
				utils.SetError(w, r, "Error creating account: "+err.Error())
			}
			utils.RenderTemplate(w, r, "register.html", data)
			return
		}

		// Set session
		utils.SetSession(w, r, "user_id", user.ID)
		utils.SetSession(w, r, "user_role", user.Role)
		utils.SetFlash(w, r, "Account created successfully! Welcome to the Library Management System")

		// Redirect to home
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Render register template
	utils.RenderTemplate(w, r, "register.html", data)
}

// Logout handles user logout
func Logout(w http.ResponseWriter, r *http.Request) {
	// Clear session
	utils.ClearSession(w, r)
	utils.SetFlash(w, r, "You have been logged out")

	// Redirect to login page
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// Profile displays the user profile
func Profile(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get profile user (either current user or requested user profile)
	profileUserID := user.ID

	// Safely extract user ID parameter if path is in format /profile/ID
	if strings.HasPrefix(r.URL.Path, "/profile/") && len(r.URL.Path) > len("/profile/") {
		idParam := r.URL.Path[len("/profile/"):]
		if idParam != "" && user.IsLibrarian {
			// Librarians can view other user profiles
			if id, err := strconv.Atoi(idParam); err == nil {
				profileUser, fetchErr := models.GetUserByID(id)
				if fetchErr == nil {
					profileUserID = profileUser.ID
				}
			}
		}
	}

	// Get user profile
	profileUser, err := models.GetUserByID(profileUserID)
	if err != nil {
		utils.SetError(w, r, "User not found")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Get active borrows
	activeBorrows, err := models.GetActiveUserBorrows(profileUserID)
	if err != nil {
		utils.SetError(w, r, "Error loading active borrows")
		utils.RenderTemplate(w, r, "user_profile.html", &utils.TemplateData{User: user})
		return
	}

	// Get pending borrows
	pendingBorrows, err := models.GetPendingUserBorrows(profileUserID)
	if err != nil {
		utils.SetError(w, r, "Error loading pending borrows")
		utils.RenderTemplate(w, r, "user_profile.html", &utils.TemplateData{User: user})
		return
	}

	// Get past borrows (returned or rejected)
	pastBorrows, err := models.GetPastUserBorrows(profileUserID)
	if err != nil {
		utils.SetError(w, r, "Error loading past borrows")
		utils.RenderTemplate(w, r, "user_profile.html", &utils.TemplateData{User: user})
		return
	}

	// Get user reservations
	var reservations []*models.Reservation
	if profileUser.IsStudent {
		reservations, err = models.GetUserReservations(profileUserID)
		if err != nil {
			utils.SetError(w, r, "Error loading reservations")
			utils.RenderTemplate(w, r, "user_profile.html", &utils.TemplateData{User: user})
			return
		}

		// Clean up expired reservations
		go models.CleanExpiredReservations()
	}

	// Prepare data for template
	data := &utils.TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Title":          "Profile",
			"profileUser":    profileUser,
			"activeBorrows":  activeBorrows,
			"pendingBorrows": pendingBorrows,
			"pastBorrows":    pastBorrows,
			"reservations":   reservations,
		},
	}

	// Render profile template
	utils.RenderTemplate(w, r, "user_profile.html", data)
}

// EditProfile handles editing user profile
func EditProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Process form submission
	if r.Method == http.MethodPost {
		// Parse form
		err := r.ParseForm()
		if err != nil {
			utils.SetError(w, r, "Error processing form")
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		}

		// Get form values
		name := r.FormValue("name")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		studentID := r.FormValue("student_id")

		// Update user
		user.Name = name
		user.Email = email
		user.Phone = sql.NullString{String: phone, Valid: phone != ""}
		user.StudentID = sql.NullString{String: studentID, Valid: studentID != ""}

		// Save changes to database
		err = user.Update()
		if err != nil {
			if err == models.ErrDuplicateEmail {
				utils.SetError(w, r, "Email already exists")
			} else {
				utils.SetError(w, r, "Error updating profile: "+err.Error())
			}
			http.Redirect(w, r, "/profile/edit", http.StatusSeeOther)
			return
		}

		// Set flash message and redirect
		utils.SetFlash(w, r, "Profile updated successfully")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Prepare data for template
	data := &utils.TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Title": "Edit Profile",
		},
	}

	// Render edit profile template
	utils.RenderTemplate(w, r, "edit_profile.html", data)
}

// ChangePassword handles changing user password
func ChangePassword(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Process form submission
	if r.Method == http.MethodPost {
		// Parse form
		err := r.ParseForm()
		if err != nil {
			utils.SetError(w, r, "Error processing form")
			http.Redirect(w, r, "/profile/password", http.StatusSeeOther)
			return
		}

		// Get form values
		currentPassword := r.FormValue("current_password")
		newPassword := r.FormValue("new_password")
		confirmPassword := r.FormValue("confirm_password")

		// Validate current password
		_, err = models.Authenticate(user.Email, currentPassword)
		if err != nil {
			utils.SetError(w, r, "Current password is incorrect")
			http.Redirect(w, r, "/profile/password", http.StatusSeeOther)
			return
		}

		// Validate new password
		if newPassword == "" {
			utils.SetError(w, r, "New password cannot be empty")
			http.Redirect(w, r, "/profile/password", http.StatusSeeOther)
			return
		}

		if newPassword != confirmPassword {
			utils.SetError(w, r, "New passwords do not match")
			http.Redirect(w, r, "/profile/password", http.StatusSeeOther)
			return
		}

		// Update password
		err = user.UpdatePassword(newPassword)
		if err != nil {
			utils.SetError(w, r, "Error updating password: "+err.Error())
			http.Redirect(w, r, "/profile/password", http.StatusSeeOther)
			return
		}

		// Set flash message and redirect
		utils.SetFlash(w, r, "Password changed successfully")
		http.Redirect(w, r, "/profile", http.StatusSeeOther)
		return
	}

	// Prepare data for template
	data := &utils.TemplateData{
		User: user,
		Data: map[string]interface{}{
			"Title": "Change Password",
		},
	}

	// Render change password template
	utils.RenderTemplate(w, r, "change_password.html", data)
}

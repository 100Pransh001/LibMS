package middleware

import (
	"context"
	"net/http"

	"library-management-system/models"
	"library-management-system/utils"
)

// userKey is the context key for the user
type userKey struct{}

// RequireAuth middleware checks if the user is authenticated
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from session
		userID := utils.GetSessionInt(r, "user_id")
		if userID <= 0 {
			// User is not authenticated, redirect to login with redirect back
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}

		// Get user from database
		user, err := models.GetUserByID(userID)
		if err != nil {
			// User not found in database, clear session and redirect to login
			utils.ClearSession(w, r)
			utils.SetError(w, r, "Session expired. Please login again.")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Add user to context
		ctx := context.WithValue(r.Context(), userKey{}, user)

		// Call the next handler with our modified context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireLibrarian middleware checks if the user is authenticated and has librarian role
func RequireLibrarian(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reuse RequireAuth middleware first
		authHandler := RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context
			user := GetUserFromContext(r)
			
			// Check if user is librarian
			if !user.IsLibrarian {
				utils.SetError(w, r, "Access denied. You need librarian privileges to access this page.")
				http.Redirect(w, r, "/", http.StatusSeeOther)
				return
			}
			
			// User is a librarian, proceed
			next.ServeHTTP(w, r)
		}))
		
		// Execute the auth handler
		authHandler.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves the user from the request context
func GetUserFromContext(r *http.Request) *models.User {
	user, ok := r.Context().Value(userKey{}).(*models.User)
	if !ok {
		return nil
	}
	return user
}

// LoadAuth middleware adds the authenticated user to the context if available
func LoadAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from session
		userID := utils.GetSessionInt(r, "user_id")
		if userID > 0 {
			// Get user from database
			user, err := models.GetUserByID(userID)
			if err == nil {
				// Add user to context
				ctx := context.WithValue(r.Context(), userKey{}, user)
				r = r.WithContext(ctx)
			} else {
				// User not found in database, clear session
				utils.ClearSession(w, r)
			}
		}
		
		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
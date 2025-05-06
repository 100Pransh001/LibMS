package routes

import (
        "net/http"

        "library-management-system/controllers"
        "library-management-system/middleware"
)

// SetupRoutes initializes all the routes for the application
func SetupRoutes() {
        // Public routes (with auth context loaded for personalization)
        http.Handle("/", middleware.LoadAuth(http.HandlerFunc(controllers.Home)))
        http.HandleFunc("/login", controllers.Login)
        http.HandleFunc("/login/student", controllers.StudentLogin)
        http.HandleFunc("/login/librarian", controllers.LibrarianLogin)
        http.HandleFunc("/register", controllers.Register)
        http.HandleFunc("/logout", controllers.Logout)
        
        // Book catalog (public, but with different functionality for authenticated users)
        http.Handle("/books", middleware.LoadAuth(http.HandlerFunc(controllers.BookList)))
        // Book detail is handled by bookHandler()
        
        // Protected routes (require authentication)
        authRoutes := http.NewServeMux()
        
        // Profile routes
        authRoutes.HandleFunc("/profile", controllers.Profile)
        authRoutes.HandleFunc("/profile/edit", controllers.EditProfile)
        authRoutes.HandleFunc("/profile/password", controllers.ChangePassword)
        
        // Borrowing routes for students
        authRoutes.HandleFunc("/books/*/borrow", controllers.BorrowBook)
        authRoutes.HandleFunc("/books/*/reserve", controllers.ReserveBook)
        authRoutes.HandleFunc("/borrows/*/return", controllers.ReturnBook)
        authRoutes.HandleFunc("/reservations", controllers.UserReservations)
        authRoutes.HandleFunc("/reservations/*/cancel", controllers.CancelReservation)
        
        // Librarian routes
        librarianRoutes := http.NewServeMux()
        
        // Book management
        librarianRoutes.HandleFunc("/books/add", controllers.AddBook)
        librarianRoutes.HandleFunc("/books/*/edit", controllers.EditBook)
        librarianRoutes.HandleFunc("/books/*/delete", controllers.DeleteBook)
        
        // User management
        librarianRoutes.HandleFunc("/users", controllers.UserList)
        librarianRoutes.HandleFunc("/users/add", controllers.AddUser)
        librarianRoutes.HandleFunc("/users/*/edit", controllers.EditUser)
        librarianRoutes.HandleFunc("/users/*/delete", controllers.DeleteUser)
        
        // Borrow management
        librarianRoutes.HandleFunc("/borrows", controllers.BorrowList)
        librarianRoutes.HandleFunc("/borrows/*/action", controllers.BorrowAction)
        librarianRoutes.HandleFunc("/borrow-history", controllers.BorrowHistory)
        
        // Reports
        librarianRoutes.HandleFunc("/borrow-report", controllers.BorrowReport)
        librarianRoutes.HandleFunc("/book-report", controllers.BookReport)
        
        // Apply middleware for authentication and authorization
        http.Handle("/profile", middleware.RequireAuth(http.HandlerFunc(controllers.Profile)))
        http.Handle("/profile/", middleware.RequireAuth(profileHandler()))
        http.Handle("/profile/edit", middleware.RequireAuth(http.HandlerFunc(controllers.EditProfile)))
        http.Handle("/profile/password", middleware.RequireAuth(http.HandlerFunc(controllers.ChangePassword)))
        
        // Book borrow/return routes
        http.Handle("/books/", bookHandler())
        http.Handle("/borrows/", borrowHandler())
        http.Handle("/reservations/", reservationHandler())
        http.Handle("/reservations", middleware.RequireAuth(http.HandlerFunc(controllers.UserReservations)))
        
        // Librarian routes with authorization middleware
        http.Handle("/users", middleware.RequireLibrarian(http.HandlerFunc(controllers.UserList)))
        http.Handle("/users/add", middleware.RequireLibrarian(http.HandlerFunc(controllers.AddUser)))
        http.Handle("/users/edit/", middleware.RequireLibrarian(userEditHandler()))
        http.Handle("/users/delete/", middleware.RequireLibrarian(userDeleteHandler()))
        
        // Borrow routes for librarians
        http.Handle("/borrows", middleware.RequireLibrarian(http.HandlerFunc(controllers.BorrowList)))
        
        // Report routes
        http.Handle("/borrow-report", middleware.RequireLibrarian(http.HandlerFunc(controllers.BorrowReport)))
        http.Handle("/book-report", middleware.RequireLibrarian(http.HandlerFunc(controllers.BookReport)))
        http.Handle("/borrow-history", middleware.RequireLibrarian(http.HandlerFunc(controllers.BorrowHistory)))
}

// Helper handler for book routes
func bookHandler() http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                path := r.URL.Path
                
                // Check if it's a borrow request
                if len(path) > 7 && path[len(path)-7:] == "/borrow" {
                        middleware.RequireAuth(http.HandlerFunc(controllers.BorrowBook)).ServeHTTP(w, r)
                        return
                }
                
                // Check if it's a reserve request
                if len(path) > 8 && path[len(path)-8:] == "/reserve" {
                        middleware.RequireAuth(http.HandlerFunc(controllers.ReserveBook)).ServeHTTP(w, r)
                        return
                }
                
                // Regular book detail with auth context loaded
                middleware.LoadAuth(http.HandlerFunc(controllers.BookDetail)).ServeHTTP(w, r)
        })
}

// Helper handler for borrow routes
func borrowHandler() http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                path := r.URL.Path
                
                // Check if it's a return request
                if len(path) > 7 && path[len(path)-7:] == "/return" {
                        middleware.RequireAuth(http.HandlerFunc(controllers.ReturnBook)).ServeHTTP(w, r)
                        return
                }
                
                // Check if it's an action request (approve/reject)
                if len(path) > 7 && path[len(path)-7:] == "/action" {
                        middleware.RequireLibrarian(http.HandlerFunc(controllers.BorrowAction)).ServeHTTP(w, r)
                        return
                }
                
                // Fallback to 404
                http.NotFound(w, r)
        })
}

// Helper handler for user edit routes
func userEditHandler() http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                controllers.EditUser(w, r)
        })
}

// Helper handler for user delete routes
func userDeleteHandler() http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                controllers.DeleteUser(w, r)
        })
}

// Helper handler for profile routes
func profileHandler() http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                // Extract ID from URL and pass to Profile controller
                controllers.Profile(w, r)
        })
}

// Helper handler for reservation routes
func reservationHandler() http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                path := r.URL.Path
                
                // Check if it's a cancel request
                if len(path) > 7 && path[len(path)-7:] == "/cancel" {
                        middleware.RequireAuth(http.HandlerFunc(controllers.CancelReservation)).ServeHTTP(w, r)
                        return
                }
                
                // Fallback to 404
                http.NotFound(w, r)
        })
}
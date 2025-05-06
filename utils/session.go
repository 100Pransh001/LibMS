package utils

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	
	"library-management-system/config"
)

var store *sessions.CookieStore

// InitSession initializes the session store
func InitSession() {
	store = sessions.NewCookieStore([]byte(config.AppConfig.Session.Secret))
	
	// Configure session cookie
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(config.AppConfig.Session.Lifetime / time.Second),
		HttpOnly: config.AppConfig.Session.HttpOnly,
		Secure:   config.AppConfig.Session.Secure,
	}
}

// GetSession returns the session for the current request
func GetSession(r *http.Request) *sessions.Session {
	session, _ := store.Get(r, config.AppConfig.Session.Name)
	return session
}

// SetSession sets a session value
func SetSession(w http.ResponseWriter, r *http.Request, key string, value interface{}) error {
	session := GetSession(r)
	session.Values[key] = value
	return session.Save(r, w)
}

// GetSessionString gets a string value from the session
func GetSessionString(r *http.Request, key string) string {
	session := GetSession(r)
	if val, ok := session.Values[key].(string); ok {
		return val
	}
	return ""
}

// GetSessionInt gets an int value from the session
func GetSessionInt(r *http.Request, key string) int {
	session := GetSession(r)
	
	// Try as int
	if val, ok := session.Values[key].(int); ok {
		return val
	}
	
	// Try as string and convert to int
	if val, ok := session.Values[key].(string); ok {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	
	return 0
}

// GetSessionBool gets a bool value from the session
func GetSessionBool(r *http.Request, key string) bool {
	session := GetSession(r)
	if val, ok := session.Values[key].(bool); ok {
		return val
	}
	return false
}

// ClearSession removes all session values
func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session := GetSession(r)
	session.Values = make(map[interface{}]interface{})
	return session.Save(r, w)
}
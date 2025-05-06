package utils

import (
        "fmt"
        "html/template"
        "net/http"
        "path/filepath"
        "strconv"
        "time"

        "library-management-system/config"
        "library-management-system/models"
)

// TemplateData holds data that will be passed to templates
type TemplateData struct {
        User           *models.User
        Data           map[string]interface{}
        Flash          string
        Error          string
        Now            time.Time
        CurrentPageURL string
}

// Template cache for parsed templates
var templateCache map[string]*template.Template

// InitTemplates initializes and caches the templates
func InitTemplates() error {
        // Initialize template cache
        if config.AppConfig.Template.CacheParsedTemplates {
                templateCache = make(map[string]*template.Template)
                
                // Get all page templates
                pages, err := filepath.Glob(filepath.Join(config.AppConfig.Template.TemplatesDir, "*.html"))
                if err != nil {
                        return err
                }
                
                // Create a template for each page
                for _, page := range pages {
                        // Get the filename
                        name := filepath.Base(page)
                        
                        // Parse the base layout with the page
                        ts, err := template.New("layout.html").Funcs(templateFuncs()).ParseFiles(
                                filepath.Join(config.AppConfig.Template.TemplatesDir, "layout.html"),
                                page,
                        )
                        
                        if err != nil {
                                return err
                        }
                        
                        // Add to cache
                        templateCache[name] = ts
                }
        }
        
        return nil
}

// RenderTemplate renders a template with the given data
func RenderTemplate(w http.ResponseWriter, r *http.Request, tmpl string, data *TemplateData) error {
        var ts *template.Template
        var err error
        
        // Set content type header
        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        
        // If caching is enabled, get template from cache
        if config.AppConfig.Template.CacheParsedTemplates {
                // Get the template from cache
                var ok bool
                ts, ok = templateCache[tmpl]
                if !ok {
                        http.Error(w, "Template not found: "+tmpl, http.StatusInternalServerError)
                        return fmt.Errorf("template not found: %s", tmpl)
                }
        } else {
                // Parse the template every time (development mode)
                ts, err = template.New("layout.html").Funcs(templateFuncs()).ParseFiles(
                        filepath.Join(config.AppConfig.Template.TemplatesDir, "layout.html"),
                        filepath.Join(config.AppConfig.Template.TemplatesDir, tmpl),
                )
                if err != nil {
                        http.Error(w, "Error parsing template: "+err.Error(), http.StatusInternalServerError)
                        return err
                }
        }
        
        // Set standard data
        if data == nil {
                data = &TemplateData{}
        }
        
        // Set current time
        data.Now = time.Now()
        
        // Set flash messages from session
        if flash := GetSessionString(r, "flash"); flash != "" {
                data.Flash = flash
                SetSession(w, r, "flash", "")
        }
        
        // Set error messages from session
        if err := GetSessionString(r, "error"); err != "" {
                data.Error = err
                SetSession(w, r, "error", "")
        }
        
        // Store current URL for post-login redirects
        data.CurrentPageURL = r.URL.Path
        
        // Execute the template
        err = ts.ExecuteTemplate(w, "layout", data)
        if err != nil {
                http.Error(w, "Error executing template: "+err.Error(), http.StatusInternalServerError)
                return err
        }
        
        return nil
}

// SetFlash sets a flash message in the session
func SetFlash(w http.ResponseWriter, r *http.Request, message string) {
        SetSession(w, r, "flash", message)
}

// SetError sets an error message in the session
func SetError(w http.ResponseWriter, r *http.Request, message string) {
        SetSession(w, r, "error", message)
}

// templateFuncs returns a map of functions that can be used in templates
func templateFuncs() template.FuncMap {
        return template.FuncMap{
                // Math functions
                "add": func(a, b int) int {
                        return a + b
                },
                "sub": func(a, b int) int {
                        return a - b
                },
                "mul": func(a, b interface{}) float64 {
                        // Type assertions for flexibility
                        var x, y float64
                        switch v := a.(type) {
                        case int:
                                x = float64(v)
                        case float64:
                                x = v
                        }
                        switch v := b.(type) {
                        case int:
                                y = float64(v)
                        case float64:
                                y = v
                        }
                        return x * y
                },
                "div": func(a, b interface{}) float64 {
                        // Type assertions for flexibility
                        var x, y float64
                        switch v := a.(type) {
                        case int:
                                x = float64(v)
                        case float64:
                                x = v
                        }
                        switch v := b.(type) {
                        case int:
                                y = float64(v)
                        case float64:
                                y = v
                        }
                        return x / y
                },
                "mod": func(a, b int) int {
                        return a % b
                },
                "float64": func(a interface{}) float64 {
                        switch v := a.(type) {
                        case int:
                                return float64(v)
                        case float64:
                                return v
                        }
                        return 0
                },
                "int": func(a interface{}) int {
                        switch v := a.(type) {
                        case int:
                                return v
                        case int64:
                                return int(v)
                        case float64:
                                return int(v)
                        case string:
                                i, _ := strconv.Atoi(v)
                                return i
                        }
                        return 0
                },
                // Date functions
                "formatDate": func(t time.Time, layout ...string) string {
                        dateFormat := "2006-01-02" // Default format for HTML date inputs
                        if len(layout) > 0 && layout[0] != "" {
                                dateFormat = layout[0]
                        }
                        return t.Format(dateFormat)
                },
                "now": func() time.Time {
                        return time.Now()
                },
                // Array/slice functions
                "eq": func(a, b interface{}) bool {
                        return a == b
                },
                "neq": func(a, b interface{}) bool {
                        return a != b
                },
                "gt": func(a, b interface{}) bool {
                        switch v := a.(type) {
                        case int:
                                if bInt, ok := b.(int); ok {
                                        return v > bInt
                                }
                        case float64:
                                if bFloat, ok := b.(float64); ok {
                                        return v > bFloat
                                }
                        }
                        return false
                },
                "lt": func(a, b interface{}) bool {
                        switch v := a.(type) {
                        case int:
                                if bInt, ok := b.(int); ok {
                                        return v < bInt
                                }
                        case float64:
                                if bFloat, ok := b.(float64); ok {
                                        return v < bFloat
                                }
                        }
                        return false
                },
                "seq": func(start, end int) []int {
                        var result []int
                        for i := start; i <= end; i++ {
                                result = append(result, i)
                        }
                        return result
                },
                // String functions
                "lower": func(s string) string {
                        return s
                },
                "upper": func(s string) string {
                        return s
                },
        }
}
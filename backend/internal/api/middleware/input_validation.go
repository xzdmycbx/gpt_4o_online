package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	// SQL injection patterns
	sqlInjectionPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(union\s+select|select\s+.*\s+from|insert\s+into|delete\s+from|drop\s+table|update\s+.*\s+set)`),
		regexp.MustCompile(`(?i)(or\s+1\s*=\s*1|and\s+1\s*=\s*1|'\s*or\s*'.*?'\s*=\s*')`),
		regexp.MustCompile(`(?i)(exec\s*\(|execute\s*\(|script\s*>)`),
	}

	// XSS patterns
	xssPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`),
		regexp.MustCompile(`(?i)javascript:`),
		regexp.MustCompile(`(?i)on\w+\s*=`), // onclick, onerror, etc.
		regexp.MustCompile(`(?i)<iframe`),
	}

	// Path traversal patterns
	pathTraversalPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\.\./`),
		regexp.MustCompile(`\.\.\\`),
	}
)

// InputValidationMiddleware validates input to prevent injection attacks
func InputValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check query parameters
		for key, values := range c.Request.URL.Query() {
			for _, value := range values {
				if isSuspicious(value) {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "Invalid input detected",
					})
					c.Abort()
					return
				}
			}
		}

		// Check form data (if exists)
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			if err := c.Request.ParseForm(); err == nil {
				for key, values := range c.Request.PostForm {
					for _, value := range values {
						if isSuspicious(value) {
							c.JSON(http.StatusBadRequest, gin.H{
								"error": "Invalid input detected",
							})
							c.Abort()
							return
						}
					}
				}
			}
		}

		c.Next()
	}
}

// isSuspicious checks if input contains suspicious patterns
func isSuspicious(input string) bool {
	input = strings.ToLower(input)

	// Check SQL injection patterns
	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}

	// Check XSS patterns
	for _, pattern := range xssPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}

	// Check path traversal patterns
	for _, pattern := range pathTraversalPatterns {
		if pattern.MatchString(input) {
			return true
		}
	}

	return false
}

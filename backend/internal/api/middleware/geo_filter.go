package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ai-chat/backend/internal/pkg/geo"
)

// GeoFilterMiddleware blocks requests from specific countries
func GeoFilterMiddleware(ipChecker *geo.IPChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP using Gin's ClientIP() which respects trusted proxies
		clientIP := c.ClientIP()

		// Check if IP should be blocked
		blocked, country, err := ipChecker.CheckIP(clientIP)
		if err != nil {
			// Log error but don't block on error
			c.Next()
			return
		}

		if blocked {
			// Always show Chinese message as per requirements
			message := "根据您所在地区法律法规，我们暂时不能为您提供服务，敬请谅解"

			c.JSON(http.StatusForbidden, gin.H{
				"error":   message,
				"country": country,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// GeoInfo middleware adds geographical information to context
func GeoInfo(ipChecker *geo.IPChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP using Gin's ClientIP() which respects trusted proxies
		clientIP := c.ClientIP()

		countryCode, countryName, err := ipChecker.GetCountry(clientIP)
		if err == nil {
			c.Set("country_code", countryCode)
			c.Set("country_name", countryName)
			c.Set("client_ip", clientIP)
		}

		c.Next()
	}
}

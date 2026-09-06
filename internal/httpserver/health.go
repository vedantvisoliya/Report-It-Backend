package httpserver

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// health godoc
//
// @Summary      Health check
// @Description  Reports whether the API server is running and returns the current UTC time.
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "ok, message and server time"
// @Router       /health [get]
func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"message": "server is running",
		"time":    time.Now().UTC(),
	})
}

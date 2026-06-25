package handler

import (
	"net/http"

	"github.com/Devlaner/devlane/api/internal/middleware"
	"github.com/Devlaner/devlane/api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SearchHandler exposes cross-entity workspace search.
type SearchHandler struct {
	Svc *service.SearchService
}

// Search handles GET /api/workspaces/:slug/search/?query=...&project_id=...
func (h *SearchHandler) Search(c *gin.Context) {
	user := middleware.GetUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}
	slug := c.Param("slug")

	var projectID *uuid.UUID
	if raw := c.Query("project_id"); raw != "" {
		pid, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}
		projectID = &pid
	}

	results, err := h.Svc.Search(c.Request.Context(), slug, c.Query("query"), projectID, user.ID)
	if err != nil {
		if err == service.ErrProjectForbidden || err == service.ErrProjectNotFound {
			c.JSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Search failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"results": results})
}

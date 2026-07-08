package handler

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnalyticsHandler struct {
	DB  *gorm.DB
	Log *slog.Logger
}

type AnalyticsResponse struct {
	ByState    map[string]int64 `json:"by_state"`
	ByPriority map[string]int64 `json:"by_priority"`
}

func (h *AnalyticsHandler) GetWorkspaceAnalytics(c *gin.Context) {
	slug := c.Param("slug")

	var stateResults []struct {
		State string
		Count int64
	}
	if err := h.DB.Table("issues").Select("state, count(*) as count").Where("workspace_slug = ?", slug).Group("state").Scan(&stateResults).Error; err != nil {
		h.Log.Error("failed to fetch workspace state analytics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	byState := make(map[string]int64)
	for _, r := range stateResults {
		byState[r.State] = r.Count
	}

	c.JSON(http.StatusOK, AnalyticsResponse{
		ByState:    byState,
		ByPriority: map[string]int64{"high": 3, "medium": 8, "low": 12}, // Exemple statique à lier à votre DB
	})
}

func (h *AnalyticsHandler) GetProjectAnalytics(c *gin.Context) {
	projectID := c.Param("projectID")

	var stateResults []struct {
		State string
		Count int64
	}
	if err := h.DB.Table("issues").Select("state, count(*) as count").Where("project_id = ?", projectID).Group("state").Scan(&stateResults).Error; err != nil {
		h.Log.Error("failed to fetch project analytics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	byState := make(map[string]int64)
	for _, r := range stateResults {
		byState[r.State] = r.Count
	}

	c.JSON(http.StatusOK, AnalyticsResponse{ByState: byState})
}

func (h *AnalyticsHandler) ExportWorkspaceCSV(c *gin.Context) {
	slug := c.Param("slug")
	filename := fmt.Sprintf("workspace-%s-analytics-%s.csv", slug, time.Now().Format("2006-01-02"))

	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Transfer-Encoding", "binary")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{"Issue ID", "Title", "State", "Priority"})

	_ = writer.Write([]string{"ISSUE-1", "Fix login bug", "In Progress", "High"})
	_ = writer.Write([]string{"ISSUE-2", "Setup Docker environment", "Done", "Medium"})
}

func (h *AnalyticsHandler) ExportProjectCSV(c *gin.Context) {
	projectID := c.Param("projectID")
	filename := fmt.Sprintf("project-%s-analytics-%s.csv", projectID, time.Now().Format("2006-01-02"))

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/csv")

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{"Project Issue ID", "Title", "State"})
	_ = writer.Write([]string{"PROJ-101", "Implement analytics endpoints", "In Progress"})
}

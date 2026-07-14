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

	err := h.DB.Table("issues").
		Select("states.name as state, count(issues.id) as count").
		Joins("INNER JOIN states ON issues.state_id = states.id").
		Joins("INNER JOIN workspaces ON issues.workspace_id = workspaces.id").
		Where("workspaces.slug = ? AND issues.deleted_at IS NULL AND workspaces.deleted_at IS NULL", slug).
		Group("states.name").
		Scan(&stateResults).Error

	if err != nil {
		h.Log.Error("failed to fetch workspace state analytics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	byState := make(map[string]int64)
	for _, r := range stateResults {
		byState[r.State] = r.Count
	}

	var priorityResults []struct {
		Priority string
		Count    int64
	}
	err = h.DB.Table("issues").
		Select("issues.priority, count(issues.id) as count").
		Joins("INNER JOIN workspaces ON issues.workspace_id = workspaces.id").
		Where("workspaces.slug = ? AND issues.deleted_at IS NULL AND workspaces.deleted_at IS NULL AND issues.priority IS NOT NULL AND issues.priority != ''", slug).
		Group("issues.priority").
		Scan(&priorityResults).Error

	if err != nil {
		h.Log.Error("failed to fetch workspace priority analytics", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	byPriority := make(map[string]int64)
	for _, r := range priorityResults {
		byPriority[r.Priority] = r.Count
	}

	c.JSON(http.StatusOK, AnalyticsResponse{
		ByState:    byState,
		ByPriority: byPriority,
	})
}

func (h *AnalyticsHandler) GetProjectAnalytics(c *gin.Context) {
	projectID := c.Param("projectID")

	var stateResults []struct {
		State string
		Count int64
	}

	err := h.DB.Table("issues").
		Select("states.name as state, count(issues.id) as count").
		Joins("INNER JOIN states ON issues.state_id = states.id").
		Where("issues.project_id = ? AND issues.deleted_at IS NULL", projectID).
		Group("states.name").
		Scan(&stateResults).Error

	if err != nil {
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

	var issues []struct {
		ID       string
		Name     string // CHANGED: Changed Title to Name to match schema column 'name'
		State    string
		Priority string
	}

	// CHANGED: Querying issues.name instead of issues.title
	err := h.DB.Table("issues").
		Select("issues.id, issues.name, states.name as state, issues.priority").
		Joins("INNER JOIN states ON issues.state_id = states.id").
		Joins("INNER JOIN workspaces ON issues.workspace_id = workspaces.id").
		Where("workspaces.slug = ? AND issues.deleted_at IS NULL AND workspaces.deleted_at IS NULL", slug).
		Scan(&issues).Error

	if err != nil {
		h.Log.Error("failed to fetch workspace issues for CSV export", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSV"})
		return
	}

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{"Issue ID", "Title", "State", "Priority"})

	for _, issue := range issues {
		_ = writer.Write([]string{
			issue.ID,
			issue.Name,
			issue.State,
			issue.Priority,
		})
	}
}

func (h *AnalyticsHandler) ExportProjectCSV(c *gin.Context) {
	projectID := c.Param("projectID")
	filename := fmt.Sprintf("project-%s-analytics-%s.csv", projectID, time.Now().Format("2006-01-02"))

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "text/csv")

	var issues []struct {
		ID    string
		Name  string // CHANGED: Changed Title to Name to match schema column 'name'
		State string
	}

	// CHANGED: Querying issues.name instead of issues.title
	err := h.DB.Table("issues").
		Select("issues.id, issues.name, states.name as state").
		Joins("INNER JOIN states ON issues.state_id = states.id").
		Where("issues.project_id = ? AND issues.deleted_at IS NULL", projectID).
		Scan(&issues).Error

	if err != nil {
		h.Log.Error("failed to fetch project issues for CSV export", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate CSV"})
		return
	}

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{"Project Issue ID", "Title", "State"})

	for _, issue := range issues {
		_ = writer.Write([]string{
			issue.ID,
			issue.Name,
			issue.State,
		})
	}
}
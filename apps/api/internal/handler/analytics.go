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
	ByAssignee map[string]int64 `json:"by_assignee"`
	ByLabel    map[string]int64 `json:"by_label"`
}

func (h *AnalyticsHandler) GetWorkspaceAnalytics(c *gin.Context) {
	slug := c.Param("slug")

	// 1. Counts by State
	var stateResults []struct {
		State string
		Count int64
	}
	err := h.DB.Table("issues").
		Select("states.name as state, count(issues.id) as count").
		Joins("INNER JOIN states ON issues.state_id = states.id").
		Joins("INNER JOIN workspaces ON issues.workspace_id = workspaces.id").
		Where("workspaces.slug = ? AND issues.deleted_at IS NULL AND workspaces.deleted_at IS NULL AND states.deleted_at IS NULL", slug).
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

	// 2. Counts by Priority
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

	// 3. Counts by Assignee (Many-to-Many via issue_assignees)
	var assigneeResults []struct {
		Email string
		Count int64
	}
	err = h.DB.Table("issue_assignees").
		Select("users.email as email, count(issue_assignees.issue_id) as count").
		Joins("INNER JOIN users ON issue_assignees.user_id = users.id").
		Joins("INNER JOIN issues ON issue_assignees.issue_id = issues.id").
		Joins("INNER JOIN workspaces ON issues.workspace_id = workspaces.id").
		Where("workspaces.slug = ? AND issues.deleted_at IS NULL AND workspaces.deleted_at IS NULL", slug).
		Group("users.email").
		Scan(&assigneeResults).Error

	byAssignee := make(map[string]int64)
	if err == nil {
		for _, r := range assigneeResults {
			byAssignee[r.Email] = r.Count
		}
	} else {
		h.Log.Warn("failed to fetch workspace assignee analytics", "error", err)
	}

	// 4. Counts by Label (Many-to-Many via issue_labels)
	var labelResults []struct {
		Label string
		Count int64
	}
	err = h.DB.Table("issue_labels").
		Select("labels.name as label, count(issue_labels.issue_id) as count").
		Joins("INNER JOIN labels ON issue_labels.label_id = labels.id").
		Joins("INNER JOIN issues ON issue_labels.issue_id = issues.id").
		Joins("INNER JOIN workspaces ON issues.workspace_id = workspaces.id").
		Where("workspaces.slug = ? AND issues.deleted_at IS NULL AND workspaces.deleted_at IS NULL AND labels.deleted_at IS NULL", slug).
		Group("labels.name").
		Scan(&labelResults).Error

	byLabel := make(map[string]int64)
	if err == nil {
		for _, r := range labelResults {
			byLabel[r.Label] = r.Count
		}
	} else {
		h.Log.Warn("failed to fetch workspace label analytics", "error", err)
	}

	c.JSON(http.StatusOK, AnalyticsResponse{
		ByState:    byState,
		ByPriority: byPriority,
		ByAssignee: byAssignee,
		ByLabel:    byLabel,
	})
}

func (h *AnalyticsHandler) GetProjectAnalytics(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		projectID = c.Param("projectID")
	}

	// 1. Project Counts by State
	var stateResults []struct {
		State string
		Count int64
	}
	err := h.DB.Table("issues").
		Select("states.name as state, count(issues.id) as count").
		Joins("INNER JOIN states ON issues.state_id = states.id").
		Where("issues.project_id = ? AND issues.deleted_at IS NULL AND states.deleted_at IS NULL", projectID).
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

	// 2. Project Counts by Priority
	var priorityResults []struct {
		Priority string
		Count    int64
	}
	err = h.DB.Table("issues").
		Select("issues.priority, count(issues.id) as count").
		Where("issues.project_id = ? AND issues.deleted_at IS NULL AND issues.priority IS NOT NULL AND issues.priority != ''", projectID).
		Group("issues.priority").
		Scan(&priorityResults).Error

	byPriority := make(map[string]int64)
	if err == nil {
		for _, r := range priorityResults {
			byPriority[r.Priority] = r.Count
		}
	}

	// 3. Project Counts by Assignee
	var assigneeResults []struct {
		Email string
		Count int64
	}
	err = h.DB.Table("issue_assignees").
		Select("users.email as email, count(issue_assignees.issue_id) as count").
		Joins("INNER JOIN users ON issue_assignees.user_id = users.id").
		Joins("INNER JOIN issues ON issue_assignees.issue_id = issues.id").
		Where("issues.project_id = ? AND issues.deleted_at IS NULL", projectID).
		Group("users.email").
		Scan(&assigneeResults).Error

	byAssignee := make(map[string]int64)
	if err == nil {
		for _, r := range assigneeResults {
			byAssignee[r.Email] = r.Count
		}
	}

	// 4. Project Counts by Label
	var labelResults []struct {
		Label string
		Count int64
	}
	err = h.DB.Table("issue_labels").
		Select("labels.name as label, count(issue_labels.issue_id) as count").
		Joins("INNER JOIN labels ON issue_labels.label_id = labels.id").
		Joins("INNER JOIN issues ON issue_labels.issue_id = issues.id").
		Where("issues.project_id = ? AND issues.deleted_at IS NULL AND labels.deleted_at IS NULL", projectID).
		Group("labels.name").
		Scan(&labelResults).Error

	byLabel := make(map[string]int64)
	if err == nil {
		for _, r := range labelResults {
			byLabel[r.Label] = r.Count
		}
	}

	c.JSON(http.StatusOK, AnalyticsResponse{
		ByState:    byState,
		ByPriority: byPriority,
		ByAssignee: byAssignee,
		ByLabel:    byLabel,
	})
}

func (h *AnalyticsHandler) ExportWorkspaceCSV(c *gin.Context) {
	slug := c.Param("slug")

	var issues []struct {
		ID       string
		Name     string
		State    string
		Priority string
	}

	err := h.DB.Table("issues").
		Select("issues.id, issues.name, states.name as state, issues.priority").
		Joins("INNER JOIN states ON issues.state_id = states.id").
		Joins("INNER JOIN workspaces ON issues.workspace_id = workspaces.id").
		Where("workspaces.slug = ? AND issues.deleted_at IS NULL AND workspaces.deleted_at IS NULL AND states.deleted_at IS NULL", slug).
		Scan(&issues).Error

	if err != nil {
		h.Log.Error("failed to fetch workspace issues for CSV export", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch workspace data"})
		return
	}

	filename := fmt.Sprintf("workspace-%s-analytics-%s.csv", slug, time.Now().Format("2006-01-02"))
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Transfer-Encoding", "binary")

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
	projectID := c.Param("projectId")
	if projectID == "" {
		projectID = c.Param("projectID")
	}

	var issues []struct {
		ID    string
		Name  string
		State string
	}

	err := h.DB.Table("issues").
		Select("issues.id, issues.name, states.name as state").
		Joins("INNER JOIN states ON issues.state_id = states.id").
		Where("issues.project_id = ? AND issues.deleted_at IS NULL AND states.deleted_at IS NULL", projectID).
		Scan(&issues).Error

	if err != nil {
		h.Log.Error("failed to fetch project issues for CSV export", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch project data"})
		return
	}

	filename := fmt.Sprintf("project-%s-analytics-%s.csv", projectID, time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	c.Header("Content-Type", "text/csv")

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
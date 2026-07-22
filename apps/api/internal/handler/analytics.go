package handler

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnalyticsHandler struct {
	DB  *gorm.DB
	Log *slog.Logger
}

// TrendPoint is one day's worth of created-vs-resolved counts for the
// created/resolved trend chart.
type TrendPoint struct {
	Date     string `json:"date"`
	Created  int64  `json:"created"`
	Resolved int64  `json:"resolved"`
}

type AnalyticsResponse struct {
	ByState    map[string]int64 `json:"by_state"`
	ByPriority map[string]int64 `json:"by_priority"`
	ByAssignee map[string]int64 `json:"by_assignee"`
	ByLabel    map[string]int64 `json:"by_label"`
	Trend      []TrendPoint     `json:"trend"`
	// PartialError is true when one or more secondary aggregates (assignee,
	// label, trend) failed to load. ByState/ByPriority failures still return
	// a 500 below since those are required for the page to render at all.
	PartialError bool     `json:"partial_error,omitempty"`
	Warnings     []string `json:"warnings,omitempty"`
}

// sanitizeCSVField neutralizes spreadsheet-formula injection: if a
// user-controlled cell starts with a character that Excel/Sheets/LibreOffice
// interpret as a formula trigger, prefix it with a single quote so it's
// imported as plain text instead of being evaluated.
func sanitizeCSVField(v string) string {
	if v == "" {
		return v
	}
	switch v[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + v
	default:
		return v
	}
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

	var warnings []string
	partialError := false

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
		partialError = true
		warnings = append(warnings, "assignee breakdown unavailable")
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
		partialError = true
		warnings = append(warnings, "label breakdown unavailable")
	}

	// 5. Created vs Resolved trend, last 30 days.
	trend, err := h.workspaceTrend(slug)
	if err != nil {
		h.Log.Warn("failed to fetch workspace created/resolved trend", "error", err)
		partialError = true
		warnings = append(warnings, "created/resolved trend unavailable")
	}

	c.JSON(http.StatusOK, AnalyticsResponse{
		ByState:      byState,
		ByPriority:   byPriority,
		ByAssignee:   byAssignee,
		ByLabel:      byLabel,
		Trend:        trend,
		PartialError: partialError,
		Warnings:     warnings,
	})
}

// workspaceTrend returns per-day created and resolved issue counts for the
// last 30 days for the given workspace. "Resolved" is approximated as the
// last-updated date of issues currently sitting in a "completed" state
// group, since the schema doesn't carry a dedicated completed-at timestamp.
func (h *AnalyticsHandler) workspaceTrend(slug string) ([]TrendPoint, error) {
	var createdRows []struct {
		Day   time.Time
		Count int64
	}
	if err := h.DB.Table("issues").
		Select("date_trunc('day', issues.created_at) as day, count(issues.id) as count").
		Joins("INNER JOIN workspaces ON issues.workspace_id = workspaces.id").
		Where("workspaces.slug = ? AND issues.deleted_at IS NULL AND workspaces.deleted_at IS NULL AND issues.created_at >= ?", slug, time.Now().AddDate(0, 0, -30)).
		Group("day").
		Scan(&createdRows).Error; err != nil {
		return nil, err
	}

	var resolvedRows []struct {
		Day   time.Time
		Count int64
	}
	if err := h.DB.Table("issues").
		Select("date_trunc('day', issues.updated_at) as day, count(issues.id) as count").
		Joins("INNER JOIN states ON issues.state_id = states.id").
		Joins("INNER JOIN workspaces ON issues.workspace_id = workspaces.id").
		Where(`workspaces.slug = ? AND issues.deleted_at IS NULL AND workspaces.deleted_at IS NULL AND states.deleted_at IS NULL AND states."group" = ? AND issues.updated_at >= ?`, slug, "completed", time.Now().AddDate(0, 0, -30)).
		Group("day").
		Scan(&resolvedRows).Error; err != nil {
		return nil, err
	}

	byDay := make(map[string]*TrendPoint)
	order := make([]string, 0, 30)
	get := func(day time.Time) *TrendPoint {
		key := day.Format("2006-01-02")
		if tp, ok := byDay[key]; ok {
			return tp
		}
		tp := &TrendPoint{Date: key}
		byDay[key] = tp
		order = append(order, key)
		return tp
	}
	for _, r := range createdRows {
		get(r.Day).Created = r.Count
	}
	for _, r := range resolvedRows {
		get(r.Day).Resolved = r.Count
	}

	sort.Strings(order)
	trend := make([]TrendPoint, 0, len(order))
	for _, key := range order {
		trend = append(trend, *byDay[key])
	}
	return trend, nil
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

	var warnings []string
	partialError := false

	byPriority := make(map[string]int64)
	if err == nil {
		for _, r := range priorityResults {
			byPriority[r.Priority] = r.Count
		}
	} else {
		h.Log.Warn("failed to fetch project priority analytics", "error", err, "project_id", projectID)
		partialError = true
		warnings = append(warnings, "priority breakdown unavailable")
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
	} else {
		h.Log.Warn("failed to fetch project assignee analytics", "error", err, "project_id", projectID)
		partialError = true
		warnings = append(warnings, "assignee breakdown unavailable")
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
	} else {
		h.Log.Warn("failed to fetch project label analytics", "error", err, "project_id", projectID)
		partialError = true
		warnings = append(warnings, "label breakdown unavailable")
	}

	c.JSON(http.StatusOK, AnalyticsResponse{
		ByState:      byState,
		ByPriority:   byPriority,
		ByAssignee:   byAssignee,
		ByLabel:      byLabel,
		PartialError: partialError,
		Warnings:     warnings,
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

	if err := writer.Write([]string{"Issue ID", "Title", "State", "Priority"}); err != nil {
		h.Log.Error("failed to write CSV header for workspace export", "error", err, "slug", slug)
		return
	}

	for _, issue := range issues {
		row := []string{
			issue.ID,
			sanitizeCSVField(issue.Name),
			sanitizeCSVField(issue.State),
			sanitizeCSVField(issue.Priority),
		}
		if err := writer.Write(row); err != nil {
			h.Log.Error("failed to write CSV row for workspace export", "error", err, "slug", slug)
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		h.Log.Error("failed to flush CSV for workspace export", "error", err, "slug", slug)
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

	if err := writer.Write([]string{"Project Issue ID", "Title", "State"}); err != nil {
		h.Log.Error("failed to write CSV header for project export", "error", err, "project_id", projectID)
		return
	}

	for _, issue := range issues {
		row := []string{
			issue.ID,
			sanitizeCSVField(issue.Name),
			sanitizeCSVField(issue.State),
		}
		if err := writer.Write(row); err != nil {
			h.Log.Error("failed to write CSV row for project export", "error", err, "project_id", projectID)
			return
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		h.Log.Error("failed to flush CSV for project export", "error", err, "project_id", projectID)
	}
}
package mail

import (
	"strings"
	"testing"
)

func TestBuildNotificationEmail_Assigned(t *testing.T) {
	data := NotificationEmailData{
		ReceiverName:  "Alice",
		ActorName:     "Bob",
		IssueRef:      "DEV-123",
		IssueTitle:    "Fix login bug",
		IssueURL:      "https://app.devlane.io/issue/abc-123",
		WorkspaceName: "Engineering",
	}

	subject, body := BuildNotificationEmail("assigned", data)

	expectedSubject := "Bob assigned you to DEV-123"
	if subject != expectedSubject {
		t.Errorf("subject mismatch: got %q, want %q", subject, expectedSubject)
	}

	if !strings.Contains(body, "Hi Alice") {
		t.Error("body should contain greeting")
	}
	if !strings.Contains(body, "Bob assigned you to DEV-123: Fix login bug") {
		t.Error("body should contain assignment message")
	}
	if !strings.Contains(body, data.IssueURL) {
		t.Error("body should contain issue URL")
	}
	if !strings.Contains(body, "Engineering") {
		t.Error("body should contain workspace name")
	}
}

func TestBuildNotificationEmail_Mentioned(t *testing.T) {
	data := NotificationEmailData{
		ReceiverName:  "Charlie",
		ActorName:     "Dana",
		IssueRef:      "PROD-456",
		IssueTitle:    "Deploy to production",
		IssueURL:      "https://app.devlane.io/issue/def-456",
		WorkspaceName: "Operations",
	}

	subject, body := BuildNotificationEmail("mentioned", data)

	expectedSubject := "Dana mentioned you in PROD-456"
	if subject != expectedSubject {
		t.Errorf("subject mismatch: got %q, want %q", subject, expectedSubject)
	}

	if !strings.Contains(body, "Dana mentioned you in PROD-456") {
		t.Error("body should contain mention message")
	}
}

func TestBuildNotificationEmail_Commented(t *testing.T) {
	data := NotificationEmailData{
		ReceiverName:   "Eve",
		ActorName:      "Frank",
		IssueRef:       "BUG-789",
		IssueTitle:     "Performance issue",
		IssueURL:       "https://app.devlane.io/issue/ghi-789",
		WorkspaceName:  "Backend",
		CommentPreview: "I think we should optimize the database query here",
	}

	subject, body := BuildNotificationEmail("commented", data)

	expectedSubject := "Frank commented on BUG-789"
	if subject != expectedSubject {
		t.Errorf("subject mismatch: got %q, want %q", subject, expectedSubject)
	}

	if !strings.Contains(body, "Frank commented on BUG-789") {
		t.Error("body should contain comment message")
	}
	if !strings.Contains(body, "Comment preview:") {
		t.Error("body should contain comment preview label")
	}
	if !strings.Contains(body, data.CommentPreview) {
		t.Error("body should contain comment preview text")
	}
}

func TestBuildNotificationEmail_StateChanged(t *testing.T) {
	data := NotificationEmailData{
		ReceiverName:  "Grace",
		ActorName:     "Henry",
		IssueRef:      "FEAT-111",
		IssueTitle:    "Add dark mode",
		IssueURL:      "https://app.devlane.io/issue/jkl-111",
		WorkspaceName: "Frontend",
		OldValue:      "In Progress",
		NewValue:      "Done",
	}

	subject, body := BuildNotificationEmail("state_changed", data)

	expectedSubject := "Henry moved FEAT-111 from In Progress to Done"
	if subject != expectedSubject {
		t.Errorf("subject mismatch: got %q, want %q", subject, expectedSubject)
	}

	if !strings.Contains(body, "From: In Progress") {
		t.Error("body should contain old state")
	}
	if !strings.Contains(body, "To: Done") {
		t.Error("body should contain new state")
	}
}

func TestBuildNotificationEmail_Subscribed(t *testing.T) {
	data := NotificationEmailData{
		ReceiverName:  "Ivan",
		ActorName:     "Jane",
		IssueRef:      "TASK-222",
		IssueTitle:    "Update documentation",
		IssueURL:      "https://app.devlane.io/issue/mno-222",
		WorkspaceName: "Documentation",
		FieldName:     "priority",
		OldValue:      "Low",
		NewValue:      "High",
	}

	subject, body := BuildNotificationEmail("subscribed", data)

	expectedSubject := "Jane updated priority on TASK-222"
	if subject != expectedSubject {
		t.Errorf("subject mismatch: got %q, want %q", subject, expectedSubject)
	}

	if !strings.Contains(body, "Jane updated priority on TASK-222") {
		t.Error("body should contain field change message")
	}
	if !strings.Contains(body, "From: Low") {
		t.Error("body should contain old value")
	}
	if !strings.Contains(body, "To: High") {
		t.Error("body should contain new value")
	}
}

func TestBuildNotificationEmail_CommentedWithoutPreview(t *testing.T) {
	data := NotificationEmailData{
		ReceiverName:  "Kevin",
		ActorName:     "Laura",
		IssueRef:      "FIX-333",
		IssueTitle:    "Memory leak",
		IssueURL:      "https://app.devlane.io/issue/pqr-333",
		WorkspaceName: "Infrastructure",
	}

	_, body := BuildNotificationEmail("commented", data)

	if strings.Contains(body, "Comment preview:") {
		t.Error("body should not contain comment preview label when no preview provided")
	}
}

func TestBuildNotificationEmail_StateChangedWithoutOldValue(t *testing.T) {
	data := NotificationEmailData{
		ReceiverName:  "Mike",
		ActorName:     "Nancy",
		IssueRef:      "ISSUE-444",
		IssueTitle:    "Test coverage",
		IssueURL:      "https://app.devlane.io/issue/stu-444",
		WorkspaceName: "QA",
		NewValue:      "In Review",
	}

	subject, body := BuildNotificationEmail("state_changed", data)

	expectedSubject := "Nancy changed the state of ISSUE-444"
	if subject != expectedSubject {
		t.Errorf("subject mismatch: got %q, want %q", subject, expectedSubject)
	}

	if !strings.Contains(body, "To: In Review") {
		t.Error("body should contain new state")
	}
	if strings.Contains(body, "From:") {
		t.Error("body should not contain 'From:' when old value is empty")
	}
}

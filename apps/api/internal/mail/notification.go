package mail

import "fmt"

// NotificationEmailData holds data for building notification email content.
type NotificationEmailData struct {
	ReceiverName   string
	ActorName      string
	IssueRef       string
	IssueTitle     string
	IssueURL       string
	WorkspaceName  string
	CommentPreview string
	FieldName      string
	OldValue       string
	NewValue       string
}

// BuildNotificationEmail returns subject and body for a notification email.
// sender is one of: "assigned", "mentioned", "commented", "state_changed", "subscribed"
func BuildNotificationEmail(sender string, data NotificationEmailData) (subject, body string) {
	subject = buildNotificationSubject(sender, data)
	body = buildNotificationBody(sender, data)
	return subject, body
}

func buildNotificationSubject(sender string, data NotificationEmailData) string {
	switch sender {
	case "assigned":
		return fmt.Sprintf("%s assigned you to %s", data.ActorName, data.IssueRef)
	case "mentioned":
		return fmt.Sprintf("%s mentioned you in %s", data.ActorName, data.IssueRef)
	case "commented":
		return fmt.Sprintf("%s commented on %s", data.ActorName, data.IssueRef)
	case "state_changed":
		if data.OldValue != "" && data.NewValue != "" {
			return fmt.Sprintf("%s moved %s from %s to %s", data.ActorName, data.IssueRef, data.OldValue, data.NewValue)
		}
		return fmt.Sprintf("%s changed the state of %s", data.ActorName, data.IssueRef)
	case "subscribed":
		if data.FieldName != "" {
			return fmt.Sprintf("%s updated %s on %s", data.ActorName, data.FieldName, data.IssueRef)
		}
		return fmt.Sprintf("%s updated %s", data.ActorName, data.IssueRef)
	default:
		return fmt.Sprintf("Update on %s", data.IssueRef)
	}
}

func buildNotificationBody(sender string, data NotificationEmailData) string {
	greeting := fmt.Sprintf("Hi %s,\n\n", data.ReceiverName)
	footer := fmt.Sprintf("\n\nView issue: %s\n\nWorkspace: %s\n\n---\nYou're receiving this because you're watching this issue.", data.IssueURL, data.WorkspaceName)

	switch sender {
	case "assigned":
		return greeting + fmt.Sprintf("%s assigned you to %s: %s", data.ActorName, data.IssueRef, data.IssueTitle) + footer

	case "mentioned":
		return greeting + fmt.Sprintf("%s mentioned you in %s: %s", data.ActorName, data.IssueRef, data.IssueTitle) + footer

	case "commented":
		msg := fmt.Sprintf("%s commented on %s: %s", data.ActorName, data.IssueRef, data.IssueTitle)
		if data.CommentPreview != "" {
			msg += fmt.Sprintf("\n\nComment preview:\n%s", data.CommentPreview)
		}
		return greeting + msg + footer

	case "state_changed":
		msg := fmt.Sprintf("%s changed the state of %s: %s", data.ActorName, data.IssueRef, data.IssueTitle)
		if data.OldValue != "" && data.NewValue != "" {
			msg += fmt.Sprintf("\nFrom: %s\nTo: %s", data.OldValue, data.NewValue)
		} else if data.NewValue != "" {
			msg += fmt.Sprintf("\nTo: %s", data.NewValue)
		}
		return greeting + msg + footer

	case "subscribed":
		var msg string
		if data.FieldName != "" {
			msg = fmt.Sprintf("%s updated %s on %s: %s", data.ActorName, data.FieldName, data.IssueRef, data.IssueTitle)
		} else {
			msg = fmt.Sprintf("%s updated %s: %s", data.ActorName, data.IssueRef, data.IssueTitle)
		}
		if data.OldValue != "" && data.NewValue != "" {
			msg += fmt.Sprintf("\nFrom: %s\nTo: %s", data.OldValue, data.NewValue)
		} else if data.NewValue != "" {
			msg += fmt.Sprintf("\nTo: %s", data.NewValue)
		}
		return greeting + msg + footer

	default:
		return greeting + fmt.Sprintf("%s updated %s: %s", data.ActorName, data.IssueRef, data.IssueTitle) + footer
	}
}

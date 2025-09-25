package models

// NotificationType can be "sms" or "email"
type Notification struct {
	ID      int64		`json:"id"`
	UserID  int64		`json:"user_id"`
	Type    string      `json:"type"` // "sms" or "email"
	To      string      `json:"to,omitempty"`      // phone or email
	Subject string      `json:"subject,omitempty"` // for emails
	Message string      `json:"message"`
	
}
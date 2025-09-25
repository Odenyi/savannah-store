package models
// NotificationType can be "sms" or "email"
type Notification struct {
	Type    string      `json:"type"` // "sms" or "email"
	To      string      `json:"to,omitempty"`      // phone or email
	Subject string      `json:"subject,omitempty"` // for emails
	Message string      `json:"message"`
	
}
type SMSPayload struct {
	Message     string `json:"message"`
	Msisdn      string `json:"msisdn"`
	SenderID    string `json:"sender_id"`
	CallbackURL string `json:"callback_url,omitempty"`
}
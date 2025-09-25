package queue

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"savannah-store/notification-service/internal/library"
	"savannah-store/notification-service/internal/models"

	"github.com/sirupsen/logrus"
	"bytes"
)

// processNotificationQueue handles notification messages (SMS or Email)
func (q *Queue) processNotificationQueue(body []byte) error {
	var notif models.Notification
	if err := json.Unmarshal(body, &notif); err != nil {
		logrus.Errorf("Failed to parse notification body: %v", err)
		return err
	}

	switch notif.Type {
	case "email":
		return sendEmail(notif)
	case "sms":
		return SendSMS(notif)
	default:
		logrus.Warnf("Unknown notification type: %s", notif.Type)
		return fmt.Errorf("unknown notification type: %s", notif.Type)
	}
}

// sendEmail sends an email using SMTP
func sendEmail(notif models.Notification) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("SMTP configuration is missing")
	}

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	to := []string{notif.To}
	msg := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", notif.Subject, notif.Message))

	err := smtp.SendMail(fmt.Sprintf("%s:%s", smtpHost, smtpPort), auth, smtpUser, to, msg)
	if err != nil {
		logrus.Errorf("Failed to send email to %s: %v", notif.To, err)
		return err
	}

	logrus.Infof("Email sent to %s successfully", notif.To)
	return nil
}


// SendSMS sends SMS via Intouch VAS transactional API
func SendSMS(notif models.Notification) error {
	token, err := library.GetSMSToken()
	if err != nil {
		log.Printf("failed to get SMS token: %v", err)
		return err
	}

	if notif.To == "" || notif.Message == "" {
		return fmt.Errorf("phone number or message is empty")
	}

	payload := models.SMSPayload{
		Message:     notif.Message,
		Msisdn:      notif.To,
		SenderID:    os.Getenv("SENDER_ID"),
	}

	data, _ := json.Marshal(payload)
	smsendpoint := os.Getenv("SMS_URL")

	req, err := http.NewRequest("POST", smsendpoint, bytes.NewBuffer(data))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", token) // use the token from GetSMSToken()

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to send SMS, status: %s", resp.Status)
	}

	log.Printf("SMS sent to %s successfully", notif.To)
	return nil
}


package models

import "time"

type User struct {
	ID            int64     `db:"id" json:"id"`
	Email         string    `db:"email" json:"email"`
	EmailVerified bool      `db:"email_verified" json:"email_verified"`
	Phone         string    `db:"phone" json:"phone,omitempty"`
	FullName      string    `db:"full_name" json:"full_name"`
	PasswordHash  string    `db:"password_hash" json:"-"`
	OTP           string    `db:"otp" json:"-"`
	Role          string    `db:"role" json:"role"`
	OIDCProvider  string    `db:"oidc_provider" json:"oidc_provider,omitempty"`
	OIDCSub       string    `db:"oidc_sub" json:"oidc_sub,omitempty"`
	Created       time.Time `db:"created" json:"created"`
	Updated       time.Time `db:"updated" json:"updated"`
}

//the payload for starting Google OAuth
type AuthRequest struct {
    Phone    string `json:"phone" example:"0712345678"`
    Usertype string `json:"usertype" example:"customer"`
}
type UserSignupRequest struct {
    Email    string `json:"email" example:"test@example.com"`
    Phone    string `json:"phone" example:"0712345678"`
    Password string `json:"password" example:"StrongPass123"`
    RoleID   int64  `json:"role_id" example:"2"`
}

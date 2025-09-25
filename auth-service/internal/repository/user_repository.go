package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"savannah-store/auth-service/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) SaveUser(ctx context.Context, u *models.User) error {
	q := `INSERT INTO user (email, email_verified, phone, full_name, password_hash, otp, role, oidc_provider, oidc_sub, created, updated)
	      VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := r.db.ExecContext(ctx, q,
		u.Email, u.EmailVerified, u.Phone, u.FullName, u.PasswordHash, u.OTP, u.Role, u.OIDCProvider, u.OIDCSub, time.Now(), time.Now(),
	)
	if err != nil {
		return err
	}
	id, _ := res.LastInsertId()
	u.ID = id
	return nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	q := `SELECT id,email,email_verified,phone,full_name,password_hash,otp,role,oidc_provider,oidc_sub,created,updated FROM user WHERE email=? LIMIT 1`
	row := r.db.QueryRowContext(ctx, q, email)
	u := &models.User{}
	err := row.Scan(&u.ID, &u.Email, &u.EmailVerified, &u.Phone, &u.FullName, &u.PasswordHash, &u.OTP, &u.Role, &u.OIDCProvider, &u.OIDCSub, &u.Created, &u.Updated)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	q := `SELECT id,email,email_verified,phone,full_name,password_hash,otp,role,oidc_provider,oidc_sub,created,updated FROM user WHERE id=? LIMIT 1`
	row := r.db.QueryRowContext(ctx, q, id)
	u := &models.User{}
	if err := row.Scan(&u.ID, &u.Email, &u.EmailVerified, &u.Phone, &u.FullName, &u.PasswordHash, &u.OTP, &u.Role, &u.OIDCProvider, &u.OIDCSub, &u.Created, &u.Updated); err != nil {
		return nil, err
	}
	return u, nil
}

package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

func HashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

// Seed admin user if not exists
func (s *Service) EnsureAdmin(ctx context.Context, email, password string) error {
	var exists bool
	if err := s.pool.QueryRow(ctx, `select exists(select 1 from admin_users where email=$1)`, email).Scan(&exists); err != nil {
		return fmt.Errorf("check admin exists: %w", err)
	}
	if exists {
		return nil
	}
	hash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = s.pool.Exec(ctx, `insert into admin_users(email, password_hash) values($1,$2)`, email, hash)
	if err != nil {
		return fmt.Errorf("insert admin: %w", err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, email, password string) (string, error) {
	var id int64
	var hash string
	err := s.pool.QueryRow(ctx, `select id, password_hash from admin_users where email=$1`, email).Scan(&id, &hash)
	if err != nil {
		return "", fmt.Errorf("invalid_credentials")
	}
	if !CheckPassword(hash, password) {
		return "", fmt.Errorf("invalid_credentials")
	}

	token := newToken()
	expires := time.Now().Add(14 * 24 * time.Hour)

	_, err = s.pool.Exec(ctx, `
		insert into admin_sessions(session_token, admin_user_id, expires_at)
		values($1,$2,$3)
	`, token, id, expires)
	if err != nil {
		return "", fmt.Errorf("create_session_failed")
	}
	return token, nil
}

func (s *Service) ValidateSession(ctx context.Context, token string) (bool, error) {
	var ok bool
	err := s.pool.QueryRow(ctx, `
		select exists(
			select 1
			from admin_sessions
			where session_token=$1 and expires_at > now()
		)
	`, token).Scan(&ok)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func newToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

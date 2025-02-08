// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: refresh_token.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createRefreshToken = `-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (
    token, 
    created_at, 
    updated_at, 
    user_id, 
    expires_at, 
    revoked_at
) VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    NULL
)
`

type CreateRefreshTokenParams struct {
	Token     string    `json:"token"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) error {
	_, err := q.db.ExecContext(ctx, createRefreshToken, arg.Token, arg.UserID, arg.ExpiresAt)
	return err
}

const getRefreshTokenById = `-- name: GetRefreshTokenById :one
SELECT token, created_at, updated_at, user_id, expires_at, revoked_at from refresh_tokens
WHERE token = $1
`

func (q *Queries) GetRefreshTokenById(ctx context.Context, token string) (RefreshToken, error) {
	row := q.db.QueryRowContext(ctx, getRefreshTokenById, token)
	var i RefreshToken
	err := row.Scan(
		&i.Token,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.UserID,
		&i.ExpiresAt,
		&i.RevokedAt,
	)
	return i, err
}

-- name: CreateRefreshToken :exec
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
);

-- name: GetRefreshTokenById :one
SELECT * from refresh_tokens
WHERE token = $1;

-- name: GetUsernameByRefreshToken :one
SELECT user_id FROM refresh_tokens
WHERE token = $1;

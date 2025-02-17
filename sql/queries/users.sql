-- name: CreateUser :one
INSERT INTO users(id, created_at, updated_at, email, hashed_password)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: DeleteUsers :execresult
DELETE FROM users *;

-- name: GetUserByEmail :one
SELECT * from users
WHERE email = $1;

-- name: GetUserByID :one
SELECT * from users
WHERE ID = $1;

-- name: UpdateUsernamePassword :one
UPDATE users
SET email = $2, hashed_password = $3, updated_at = NOW()
WHERE id = $1
RETURNING id, created_at, updated_at, email;

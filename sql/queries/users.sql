-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, hashed_password, email)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
) 
RETURNING *;

-- name: DeleteAllUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users 
WHERE email=$1;

-- name: GetUserFromRefreshToken :one
SELECT * FROM users 
WHERE users.id = (
    SELECT user_id FROM refresh_tokens
    WHERE refresh_tokens.token = $1
    AND refresh_tokens.expires_at > NOW()
    AND refresh_tokens.revoked_at IS NULL
);

-- name: UpdateUserPasswordById :exec
UPDATE users
SET
email = $2,
hashed_password = $3,
updated_at = NOW()
WHERE id= $1;

-- name: GetUserById :one
SELECT * FROM users
WHERE id=$1;

-- name: AlterChirpyRed :exec
UPDATE users 
SET 
is_chirpy_red = $2
WHERE id=$1;
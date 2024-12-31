-- name: CreateChirpy :one
INSERT INTO chirpy(id, created_at, updated_at, body, user_id)
 VALUES (
      gen_random_uuid(),
      NOW(),
      NOW(),
      $1, 
      $2
  
 ) RETURNING *;

-- name: GetAllChirps :many
 SELECT * FROM chirpy
 ORDER BY created_at ASC;

-- name: GetChirpById :one
 SELECT * FROM chirpy
 WHERE id=$1;

-- name: DeleteChirpyById :exec
 DELETE FROM chirpy
 WHERE id=$1;
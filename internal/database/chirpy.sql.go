// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: chirpy.sql

package database

import (
	"context"

	"github.com/google/uuid"
)

const createChirpy = `-- name: CreateChirpy :one
INSERT INTO chirpy(id, created_at, updated_at, body, user_id)
 VALUES (
      gen_random_uuid(),
      NOW(),
      NOW(),
      $1, 
      $2
  
 ) RETURNING id, created_at, updated_at, body, user_id
`

type CreateChirpyParams struct {
	Body   string
	UserID uuid.UUID
}

func (q *Queries) CreateChirpy(ctx context.Context, arg CreateChirpyParams) (Chirpy, error) {
	row := q.db.QueryRowContext(ctx, createChirpy, arg.Body, arg.UserID)
	var i Chirpy
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Body,
		&i.UserID,
	)
	return i, err
}

const deleteChirpyById = `-- name: DeleteChirpyById :exec
 DELETE FROM chirpy
 WHERE id=$1
`

func (q *Queries) DeleteChirpyById(ctx context.Context, id uuid.UUID) error {
	_, err := q.db.ExecContext(ctx, deleteChirpyById, id)
	return err
}

const getAllChirps = `-- name: GetAllChirps :many
 SELECT id, created_at, updated_at, body, user_id FROM chirpy
 ORDER BY created_at ASC
`

func (q *Queries) GetAllChirps(ctx context.Context) ([]Chirpy, error) {
	rows, err := q.db.QueryContext(ctx, getAllChirps)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []Chirpy
	for rows.Next() {
		var i Chirpy
		if err := rows.Scan(
			&i.ID,
			&i.CreatedAt,
			&i.UpdatedAt,
			&i.Body,
			&i.UserID,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getChirpById = `-- name: GetChirpById :one
 SELECT id, created_at, updated_at, body, user_id FROM chirpy
 WHERE id=$1
`

func (q *Queries) GetChirpById(ctx context.Context, id uuid.UUID) (Chirpy, error) {
	row := q.db.QueryRowContext(ctx, getChirpById, id)
	var i Chirpy
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Body,
		&i.UserID,
	)
	return i, err
}

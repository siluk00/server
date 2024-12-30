package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/siluk00/server/internal/auth"
	"github.com/siluk00/server/internal/database"
)

func (cfg *apiConfig) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	type userRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req userRequest

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)

	if err != nil {
		respondWithError(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	var params database.CreateUserParams
	params.Email = req.Email
	params.HashedPassword, err = auth.HashPassword(req.Password)

	if err != nil {
		respondWithError(w, "error hashing password: "+err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), params)
	if err != nil {
		respondWithError(w, "Error querying database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type userResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	var res userResponse
	res.CreatedAt = user.CreatedAt
	res.Email = user.Email
	res.ID = user.ID
	res.UpdatedAt = user.UpdatedAt

	w.WriteHeader(http.StatusCreated)
	marshalAndWrite(w, res)

}

func (cfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	type userRequest struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	var req userRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)

	if err != nil {
		respondWithError(w, "error decoding request: "+err.Error(), http.StatusBadRequest)
		return
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), req.Email)

	if err != nil {
		respondWithError(w, "email not found: "+err.Error(), http.StatusUnauthorized)
		return
	}

	err = auth.CheckPasswordHash(req.Password, user.HashedPassword)
	if err != nil {
		respondWithError(w, "unmatched password: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var expires time.Duration
	exp := time.Duration(req.ExpiresInSeconds) * time.Second

	if exp > time.Hour || exp == time.Duration(0) {
		expires = time.Hour
	} else {
		expires = exp
	}

	jwtToken, err := auth.MakeJWT(user.ID, os.Getenv("SECRET"), expires)

	if err != nil {
		respondWithError(w, "error generating jwt:"+err.Error(), http.StatusInternalServerError)
		return
	}

	var tokenParams database.CreateRefreshTokenParams
	tokenParams.UserID = user.ID
	tokenParams.ExpiresAt = time.Now().Add(60 * 24 * time.Hour)
	tokenParams.Token, err = auth.MakeRefreshToken()

	if err != nil {
		respondWithError(w, "error generating refres_token: "+err.Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, err := cfg.dbQueries.CreateRefreshToken(r.Context(), tokenParams)

	if err != nil {
		respondWithError(w, "Error creating token in database: "+err.Error(), http.StatusInternalServerError)
	}

	type userResponse struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}

	var res userResponse
	res.CreatedAt = user.CreatedAt
	res.Email = user.Email
	res.ID = user.ID
	res.UpdatedAt = user.UpdatedAt
	res.RefreshToken = refreshToken.Token
	res.Token = jwtToken

	w.WriteHeader(http.StatusOK)
	marshalAndWrite(w, res)

}

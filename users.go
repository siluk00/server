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
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	var res userResponse
	res.CreatedAt = user.CreatedAt
	res.Email = user.Email
	res.ID = user.ID
	res.UpdatedAt = user.UpdatedAt
	res.IsChirpyRed = user.IsChirpyRed

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
		IsChirpyRed  bool      `json:"is_chirpy_red"`
	}

	var res userResponse
	res.CreatedAt = user.CreatedAt
	res.Email = user.Email
	res.ID = user.ID
	res.UpdatedAt = user.UpdatedAt
	res.RefreshToken = refreshToken.Token
	res.Token = jwtToken
	res.IsChirpyRed = user.IsChirpyRed

	w.WriteHeader(http.StatusOK)
	marshalAndWrite(w, res)
}

func (cfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	type userRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var req userRequest
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)

	if err != nil {
		respondWithError(w, "error decoding request", http.StatusBadRequest)
		return
	}

	defer r.Body.Close()

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, "error getting token", http.StatusUnauthorized)
		return
	}

	id, err := auth.ValidateJWT(token, os.Getenv("SECRET"))

	if err != nil {
		respondWithError(w, "error validating jwt", http.StatusUnauthorized)
		return
	}

	user, err := cfg.dbQueries.GetUserById(r.Context(), id)

	if err != nil {
		respondWithError(w, "errror getting user: "+err.Error(), http.StatusInternalServerError)
	}

	hashed, err := auth.HashPassword(req.Password)

	if err != nil {
		respondWithError(w, "error hashing password", http.StatusInternalServerError)
		return
	}

	var params database.UpdateUserPasswordByIdParams
	params.Email = req.Email
	params.HashedPassword = hashed
	params.ID = user.ID

	err = cfg.dbQueries.UpdateUserPasswordById(r.Context(), params)

	if err != nil {
		respondWithError(w, "error updating database: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type userResponse struct {
		ID          uuid.UUID `json:"id"`
		CreatedAt   time.Time `json:"created_at"`
		UpdatedAt   time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	var res userResponse
	res.CreatedAt = user.CreatedAt
	res.Email = req.Email
	res.ID = user.ID
	res.UpdatedAt = user.UpdatedAt
	res.IsChirpyRed = user.IsChirpyRed

	w.WriteHeader(http.StatusOK)
	marshalAndWrite(w, &res)

}

func (cfg *apiConfig) handlePolkaWebhooks(w http.ResponseWriter, r *http.Request) {
	auth, err := auth.GetAPIKey(r.Header)

	if err != nil {
		respondWithError(w, "Wrong Header: "+err.Error(), http.StatusUnauthorized)
	}

	if auth != os.Getenv("POLKA_KEY") {
		respondWithError(w, "Wrong Api Key: "+err.Error(), http.StatusUnauthorized)
	}

	type requestBody struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	var req requestBody
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&req)

	if err != nil {
		respondWithError(w, "couldnt parse request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	id, err := uuid.Parse(req.Data.UserID)
	if err != nil {
		respondWithError(w, "could not parse id: "+err.Error(), http.StatusBadRequest)
	}

	if _, err = cfg.dbQueries.GetUserById(r.Context(), id); err != nil {
		respondWithError(w, "couldn't find user: "+err.Error(), http.StatusNotFound)
		return
	}

	var params database.AlterChirpyRedParams
	params.ID = id
	params.IsChirpyRed = true

	cfg.dbQueries.AlterChirpyRed(r.Context(), params)

	w.WriteHeader(http.StatusNoContent)
}

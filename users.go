package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {
	type userRequest struct {
		Email string `json:"email"`
	}

	var req userRequest

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&req)

	if err != nil {
		respondWithError(w, "Error decoding request", http.StatusBadRequest)
		return
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), req.Email)
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

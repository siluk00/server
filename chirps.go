package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/siluk00/server/internal/auth"
	"github.com/siluk00/server/internal/database"
)

func (cfg *apiConfig) handleChirps(w http.ResponseWriter, r *http.Request) {
	type requestBody struct {
		Body string `json:"body"`
	}

	var req requestBody
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	err := decoder.Decode(&req)

	if err != nil {
		respondWithError(w, "Error decoding body request", http.StatusBadRequest)
		return
	}

	if len(req.Body) > 140 {
		respondWithError(w, "Chirpy is too long", http.StatusBadRequest)
		return
	} else if len(req.Body) == 0 {
		respondWithError(w, "Chirpy is null", http.StatusBadRequest)
		return
	}

	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, "Bearer token problem:"+err.Error(), http.StatusInternalServerError)
		return
	}

	userId, err := auth.ValidateJWT(token, os.Getenv("SECRET"))

	if err != nil {
		respondWithError(w, "Jwt invalid: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var res responseBody
	responseMessage := validateWords(req.Body)
	res.Body = responseMessage

	var params database.CreateChirpyParams
	params.Body = req.Body
	params.UserID = userId

	chirpy, err := cfg.dbQueries.CreateChirpy(r.Context(), params)

	if err != nil {
		respondWithError(w, "Error creating chirpy: "+err.Error(), http.StatusBadRequest)
	}

	res.Body = chirpy.Body
	res.CreatedAt = chirpy.CreatedAt
	res.UpdatedAt = chirpy.UpdatedAt
	res.ID = chirpy.ID
	res.UserID = userId

	w.WriteHeader(http.StatusCreated)
	marshalAndWrite(w, &res)
}

type responseBody struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
}

func (cfg *apiConfig) handleGetChirps(w http.ResponseWriter, r *http.Request) {
	var chirpies []database.Chirpy
	var err error
	authorId := r.URL.Query().Get("author_id")
	sort := r.URL.Query().Get("sort")
	if sort != "asc" && sort != "desc" {
		sort = "asc"
	}

	if authorId == "" {
		if sort == "asc" {
			chirpies, err = cfg.dbQueries.GetAllChirps(r.Context())
		} else {
			chirpies, err = cfg.dbQueries.GetAllChirpsDesc(r.Context())

		}
		if err != nil {
			respondWithError(w, "Error getting chirps: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		authorIdUUID, err := uuid.Parse(authorId)

		if err != nil {
			respondWithError(w, "error parsing UUID", http.StatusNotFound)
			return
		}
		if sort == "asc" {
			chirpies, err = cfg.dbQueries.GetChirpsByUserId(r.Context(), authorIdUUID)
		} else {
			chirpies, err = cfg.dbQueries.GetChirpsByUserIdDesc(r.Context(), authorIdUUID)
		}

		if err != nil {
			respondWithError(w, "Error getting chirps: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	res := make([]responseBody, len(chirpies))

	for i := range chirpies {
		res[i].Body = chirpies[i].Body
		res[i].CreatedAt = chirpies[i].CreatedAt
		res[i].ID = chirpies[i].ID
		res[i].UpdatedAt = chirpies[i].UpdatedAt
		res[i].UserID = chirpies[i].UserID
	}

	w.WriteHeader(http.StatusOK)
	marshalAndWrite(w, res)
}

func validateWords(s string) string {
	badWords := make([]string, 3)
	badWords[0] = "kerfuffle"
	badWords[1] = "sharbert"
	badWords[2] = "fornax"

	lower := strings.ToLower(s)
	lowerSlice := strings.Split(lower, " ")
	sSlice := strings.Split(s, " ")

	for i := 0; i < len(lowerSlice); i++ {
		for _, badword := range badWords {
			if lowerSlice[i] == badword {
				sSlice[i] = "****"
			}
		}
	}

	return strings.Join(sSlice, " ")
}

func (cfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	chirpy, err := cfg.dbQueries.GetChirpById(r.Context(), uuid.MustParse(r.PathValue("chirp_id")))

	if err != nil {
		respondWithError(w, "Couldn't find chirp: "+err.Error(), http.StatusNotFound)
	}

	var res responseBody
	res.Body = chirpy.Body
	res.CreatedAt = chirpy.CreatedAt
	res.ID = chirpy.ID
	res.UserID = chirpy.UserID
	res.UpdatedAt = chirpy.UpdatedAt
	marshalAndWrite(w, &res)

}

func (cfg *apiConfig) handleDeleteChirpById(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, "error getting token: "+err.Error(), http.StatusUnauthorized)
		return
	}

	idUser, err := auth.ValidateJWT(token, os.Getenv("SECRET"))

	if err != nil {
		respondWithError(w, "error validating: "+err.Error(), http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("chirp_id")
	fmt.Println(idStr)
	idChirpy, err := uuid.Parse(idStr)

	if err != nil {
		respondWithError(w, "Invalid chirpy id: "+err.Error(), http.StatusForbidden)
		return
	}

	chirpy, err := cfg.dbQueries.GetChirpById(r.Context(), idChirpy)

	if err != nil {
		respondWithError(w, "Couln't find chirpy on database"+err.Error(), http.StatusNotFound)
		return
	}

	if chirpy.UserID != idUser {
		respondWithError(w, "Not authorized: ", http.StatusForbidden)
		return
	}

	err = cfg.dbQueries.DeleteChirpyById(r.Context(), idChirpy)
	if err != nil {
		respondWithError(w, "couldn't delete chirp: "+err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

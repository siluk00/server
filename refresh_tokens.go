package main

import (
	"net/http"
	"os"
	"time"

	"github.com/siluk00/server/internal/auth"
)

func (cfg *apiConfig) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, "error getting token: "+err.Error(), http.StatusUnauthorized)
	}

	user, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), token)

	if err != nil {
		respondWithError(w, "error getting token from database: "+err.Error(), http.StatusUnauthorized)
	}

	jwtToken, err := auth.MakeJWT(user.ID, os.Getenv("SECRET"), time.Hour)

	if err != nil {
		respondWithError(w, "Error creating jwt: "+err.Error(), http.StatusUnauthorized)
	}

	type ResponseBody struct {
		Token string `json:"token"`
	}

	var res ResponseBody
	res.Token = jwtToken
	w.WriteHeader(200)
	marshalAndWrite(w, &res)
}

func (cfg *apiConfig) handleRevokeToken(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		respondWithError(w, "error getting token: "+err.Error(), http.StatusUnauthorized)
	}

	err = cfg.dbQueries.RevokeToken(r.Context(), token)

	if err != nil {
		respondWithError(w, "Error revoking token: "+err.Error(), http.StatusInternalServerError)
	}

	w.WriteHeader(http.StatusNoContent)
}

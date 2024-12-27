package main

import (
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/siluk00/server/internal/database"
)

type apiConfig struct {
	fileServerHits atomic.Int32
	dbQueries      *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Now you have access to w and r!
		cfg.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {

	if os.Getenv("PLATFORM") != "dev" {
		respondWithError(w, "Non-local user", http.StatusForbidden)
		return
	}

	cfg.fileServerHits.Store(0)

	err := cfg.dbQueries.DeleteAllUsers(r.Context())
	if err != nil {
		respondWithError(w, "Error Accessing database", http.StatusInternalServerError)
		return
	}
}

func (cfg *apiConfig) handleAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	_, err := w.Write([]byte(fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileServerHits.Load())))
	if err != nil {
		fmt.Printf("Error writing response: %s\n", err)
		return
	}
}

package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/siluk00/server/internal/database"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading environment variable: %s\n", err)
	}

	dbUrl := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbUrl)
	dbQueries := database.New(db)

	if err != nil {
		log.Fatalf("Error opening connection to database: %s\n", err)
	}

	var server http.Server
	server.Addr = ":8080"
	servemux := http.NewServeMux()
	server.Handler = servemux
	var cfg apiConfig
	cfg.fileServerHits.Store(0)
	cfg.dbQueries = dbQueries

	servemux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./app")))))
	servemux.Handle("/assets/", cfg.middlewareMetricsInc(http.StripPrefix("/assets", http.FileServer(http.Dir("./app/assets")))))
	servemux.HandleFunc("GET /api/healthz", handleHealthz)
	servemux.HandleFunc("POST /admin/reset", cfg.handleReset)
	servemux.HandleFunc("GET /admin/metrics", cfg.handleAdmin)
	servemux.HandleFunc("POST /api/users", cfg.handleCreateUser)
	servemux.HandleFunc("POST /api/chirps", cfg.handleChirps)
	servemux.HandleFunc("GET /api/chirps", cfg.handleGetChirps)
	servemux.HandleFunc("GET /api/chirps/{chirp_id}", cfg.handleGetChirp)
	servemux.HandleFunc("POST /api/login", cfg.handleLogin)
	servemux.HandleFunc("POST /api/refresh", cfg.handleRefreshToken)
	servemux.HandleFunc("POST /api/revoke", cfg.handleRevokeToken)
	servemux.HandleFunc("PUT /api/users", cfg.handleUpdateUser)
	servemux.HandleFunc("DELETE /api/chirps/{chirp_id}", cfg.handleDeleteChirpById)
	servemux.HandleFunc("POST /api/polka/webhooks", cfg.handlePolkaWebhooks)

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error serving: %s\n", err)
	}

}

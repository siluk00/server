package main

import (
	"log"
	"net/http"
)

func main() {
	var server http.Server
	server.Addr = ":8080"
	servemux := http.NewServeMux()
	server.Handler = servemux
	var cfg apiConfig
	cfg.fileServerHits.Store(0)

	servemux.Handle("/app/", cfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./app")))))
	servemux.Handle("/assets/", cfg.middlewareMetricsInc(http.StripPrefix("/assets", http.FileServer(http.Dir("./app/assets")))))
	servemux.HandleFunc("GET /api/healthz", handleHealthz)
	servemux.HandleFunc("POST /admin/reset", cfg.handleReset)
	servemux.HandleFunc("GET /admin/metrics", cfg.handleAdmin)

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error serving: %s\n", err)
	}

}

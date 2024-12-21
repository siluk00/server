package main

import "net/http"

func main() {
	var server http.Server
	server.Addr = ":8080"
	servemux := http.NewServeMux()
	server.Handler = servemux

	servemux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("./app"))))
	servemux.Handle("/assets/", http.StripPrefix("/assets", http.FileServer(http.Dir("./app/assets"))))
	servemux.HandleFunc("/healthz", handleHealthz)

	server.ListenAndServe()

}

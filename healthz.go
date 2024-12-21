package main

import (
	"fmt"
	"net/http"
)

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)

	_, err := w.Write([]byte("OK"))
	if err != nil {
		fmt.Printf("Error writing body response: %s\n", err)
	}

}

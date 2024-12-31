package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func marshalAndWrite(w http.ResponseWriter, res any) {
	w.Header().Set("Content-Type", "application/json")
	body, err := json.Marshal(res)

	if err != nil {
		fmt.Printf("Error Marshaling response: %s\n", err)
		return
	}

	_, err = w.Write(body)

	if err != nil {
		fmt.Printf("Error writing body")
		return
	}
}

func respondWithError(w http.ResponseWriter, message string, errorCode int) {

	type responseError struct {
		Error string `json:"error"`
	}

	var res responseError
	res.Error = message
	w.WriteHeader(errorCode)
	marshalAndWrite(w, &res)
}

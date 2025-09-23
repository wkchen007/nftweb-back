package main

import (
	"fmt"
	"log"
	"net/http"
)

// /healthz handler
func (app *application) healthzHandler(w http.ResponseWriter, r *http.Request) {
	//服務本身活著就回200
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "ok")
	log.Printf("[http] health check ok")
}

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	var payload = struct {
		Status  string `json:"status"`
		Message string `json:"message"`
		Version string `json:"version"`
	}{
		Status:  "active",
		Message: "Nft web demo",
		Version: "1.0.0",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

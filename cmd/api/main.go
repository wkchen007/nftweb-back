package main

import (
	"flag"
	"log"
	"net/http"
)

type application struct {
	httpAddr string
}

func main() {
	var app application

	flag.StringVar(&app.httpAddr, "httpAddr", ":8080", "HTTP network address")
	flag.Parse()

	// 啟動 HTTP server
	log.Printf("HTTP server listening on %s ...", app.httpAddr)
	if err := http.ListenAndServe(app.httpAddr, app.routes()); err != nil {
		log.Fatal(err)
	}
}

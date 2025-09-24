package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (app *application) routes() http.Handler {
	// create a router mux
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	mux.Get("/", app.Home)
	mux.Get("/healthz", app.healthzHandler)

	mux.Route("/wallet", func(mux chi.Router) {
		mux.Get("/address", app.GetWalletAddress)
		mux.Get("/balance", app.GetWalletBalance)
		mux.Post("/transfer", app.PostWalletTransfer)
	})

	mux.Route("/nft", func(mux chi.Router) {
		mux.Post("/OwnerOf", app.nft.OwnerOf)
		mux.Post("/Mint", app.nft.Mint)
	})

	return mux
}

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
	mux.Get("/demo", app.AllNFTs)

	mux.Route("/wallet", func(mux chi.Router) {
		mux.Post("/useSigner", app.PostWalletUseSigner)
		mux.Get("/address", app.GetWalletAddress)
		mux.Get("/balance", app.GetWalletBalance)
		mux.Post("/transfer", app.PostWalletTransfer)
	})

	mux.Route("/nft", func(mux chi.Router) {
		mux.Get("/owner", app.nft.Owner)
		mux.Post("/ownerOf", app.nft.OwnerOf)
		mux.Post("/mint", app.nft.Mint)
		mux.Post("/tokensOfOwner", app.nft.TokensOfOwner)
		mux.Get("/openBlindBox", app.nft.OpenBlindBox)
		mux.Get("/tokenURI/{id}", app.nft.TokenURI)
		mux.Get("/withdraw", app.nft.Withdraw)
		mux.Get("/balance", app.nft.Balance)
	})

	return mux
}

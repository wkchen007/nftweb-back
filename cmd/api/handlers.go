package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/wkchen007/nftweb-back/internal/ethcli"
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

func (app *application) AllNFTs(w http.ResponseWriter, r *http.Request) {
	nfts, err := app.DB.AllNFTs()
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, nfts)
}

func (app *application) GetWalletAddress(w http.ResponseWriter, r *http.Request) {
	fromAddr := app.ethClient.GetAddress()

	_ = app.writeJSON(w, http.StatusOK, fromAddr)
}

func (app *application) GetWalletBalance(w http.ResponseWriter, r *http.Request) {
	wallet, err := app.ethClient.GetBalance()
	if err != nil {
		app.errorJSON(w, err, http.StatusBadGateway)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, wallet)
}

func (app *application) PostWalletTransfer(w http.ResponseWriter, r *http.Request) {
	var req ethcli.TransferRequest
	err := app.readJSON(w, r, &req)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Printf("[http] transfer request: %+v", req)

	txRes, err := app.ethClient.TransferETH(req)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadGateway)
		return
	}

	_ = app.writeJSON(w, http.StatusOK, txRes)
}

func (app *application) PostWalletUseSigner(w http.ResponseWriter, r *http.Request) {
	var req ethcli.UseSignerRequest
	err := app.readJSON(w, r, &req)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Printf("[http] useSigner request: %+v", req)

	// 切換 signer（會更新 from/chainID/network）
	if err := app.ethClient.UseSigner(req); err != nil {
		app.errorJSON(w, err, http.StatusBadGateway)
		return
	}

	resp := ethcli.UseSignerResponse{
		Address: app.ethClient.From().Hex(),
		Network: app.ethClient.Network(),
	}

	_ = app.writeJSON(w, http.StatusOK, resp)
}

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

func (app *application) authenticate(w http.ResponseWriter, r *http.Request) {
	// read json payload
	var requestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	// validate user against database
	user, err := app.DB.GetUserByEmail(requestPayload.Email)
	if err != nil {
		app.errorJSON(w, fmt.Errorf("invalid credentials"), http.StatusBadRequest)
		return
	}

	// check password
	valid, err := user.PasswordMatches(requestPayload.Password)
	if err != nil || !valid {
		app.errorJSON(w, fmt.Errorf("invalid email or password"), http.StatusBadRequest)
		return
	}

	//寫入 eth client 的 signer
	req := ethcli.UseSignerRequest{
		PrivateKey: os.Getenv("PRIVATE_KEY"),
	}

	if err := app.ethClient.UseSigner(req); err != nil {
		app.errorJSON(w, err, http.StatusBadGateway)
		return
	}

	// create a jwt user
	u := jwtUser{
		ID:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	// generate tokens
	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		app.errorJSON(w, err)
		return
	}

	//log.Println(tokens.Token)
	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)
	http.SetCookie(w, refreshCookie)

	app.writeJSON(w, http.StatusAccepted, tokens)
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, app.auth.GetExpiredRefreshCookie())
	w.WriteHeader(http.StatusAccepted)
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

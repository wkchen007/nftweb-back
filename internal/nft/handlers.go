package nft

import (
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/wkchen007/nftweb-back/internal/models"
)

type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// OwnerOf 查詢 NFT 擁有者
type OwnerOfRequest struct {
	Contract string `json:"contract"`
	TokenID  string `json:"tokenId"`
}

type OwnerOfResponse struct {
	Contract string `json:"contract"`
	TokenID  string `json:"tokenId"`
	Owner    string `json:"owner"`
}

func (h *Handlers) OwnerOf(w http.ResponseWriter, r *http.Request) {
	var req OwnerOfRequest
	err := h.readJSON(w, r, &req)
	if err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Printf("[nft] OwnerOf request: %+v", req)

	resp, err := h.svc.OwnerOf(req)
	if err != nil {
		h.errorJSON(w, fmt.Errorf("ownerOf failed: %w", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

// Mint 鑄造 NFT
type MintRequest struct {
	Amount   string `json:"amount"`
	ValueETH string `json:"valueETH,omitempty"`
}

type MintResponse struct {
	TxHash string `json:"txHash"`
	From   string `json:"from"`
}

func (h *Handlers) Mint(w http.ResponseWriter, r *http.Request) {
	var req MintRequest
	err := h.readJSON(w, r, &req)
	if err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	if req.ValueETH == "" {
		num, err := strconv.Atoi(req.Amount)
		if err != nil {
			h.errorJSON(w, fmt.Errorf("invalid amount: %w", err), http.StatusBadRequest)
			return
		}
		req.ValueETH = fmt.Sprintf("%f", float64(num)*0.01) // 單價 0.01 ETH
	}
	log.Printf("[nft] Mint request: %+v", req)

	resp, err := h.svc.Mint(req)
	if err != nil {
		h.errorJSON(w, fmt.Errorf("mint failed: %w", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

type OwnerResponse struct {
	Contract string `json:"contract"`
	Owner    string `json:"owner"`
}

func (h *Handlers) Owner(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.ConCreator()
	if err != nil {
		h.errorJSON(w, fmt.Errorf("contract creator failed: %w", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

type TokensOfOwnerRequest struct {
	Owner           string `json:"owner"`
	IncludeTokenURI bool   `json:"includeTokenURI,omitempty"` // 是否同時查 tokenURI
}

type TokensOfOwnerResponse struct {
	Owner  string             `json:"owner"`
	Count  int                `json:"count"`
	Tokens []models.TokenItem `json:"tokens"`
}

func (h *Handlers) TokensOfOwner(w http.ResponseWriter, r *http.Request) {
	var req TokensOfOwnerRequest
	err := h.readJSON(w, r, &req)
	if err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Printf("[nft] TokensOfOwner request: %+v", req)

	resp, err := h.svc.TokensOfOwner(req)
	if err != nil {
		h.errorJSON(w, fmt.Errorf("tokensOfOwner failed: %w", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

type ConResponse struct {
	TxHash   string `json:"txHash"`
	Contract string `json:"contract"`
}

func (h *Handlers) OpenBlindBox(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.OpenBlindBox()
	if err != nil {
		h.errorJSON(w, fmt.Errorf("openBlindBox failed: %w", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) Withdraw(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.Withdraw()
	if err != nil {
		h.errorJSON(w, fmt.Errorf("withdraw failed: %w", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

func (h *Handlers) TokenURI(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.errorJSON(w, fmt.Errorf("tokenId is required"), http.StatusBadRequest)
		return
	}

	bigID := new(big.Int)
	_, ok := bigID.SetString(id, 10)
	if !ok {
		h.errorJSON(w, fmt.Errorf("invalid tokenId: %s", id), http.StatusBadRequest)
		return
	}
	uri, err := h.svc.TokenURI(bigID)
	if err != nil {
		h.errorJSON(w, fmt.Errorf("tokenURI failed: %w", err), http.StatusInternalServerError)
		return
	}

	resp := struct {
		TokenID  int    `json:"tokenId"`
		TokenURI string `json:"tokenURI"`
	}{
		TokenID:  int(bigID.Int64()),
		TokenURI: uri,
	}

	h.writeJSON(w, http.StatusOK, resp)
}

type BalanceResponse struct {
	Contract   string `json:"contract"`
	Balance    string `json:"balance"`    // wei
	BalanceETH string `json:"balanceETH"` // ETH
}

func (h *Handlers) Balance(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.Balance()
	if err != nil {
		h.errorJSON(w, fmt.Errorf("balance failed: %w", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

type CountResponse struct {
	Count int `json:"count"`
	Total int `json:"total"`
}

func (h *Handlers) Count(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.Count()
	if err != nil {
		h.errorJSON(w, fmt.Errorf("counter failed: %w", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

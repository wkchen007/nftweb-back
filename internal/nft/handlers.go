package nft

import (
	"fmt"
	"log"
	"net/http"
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
	Contract string `json:"contract"`
	To       string `json:"to"`
	Amount   string `json:"amount"`
	ValueETH string `json:"valueEth"`
}

type MintResponse struct {
	TxHash   string `json:"txHash"`
	From     string `json:"from"`
	Contract string `json:"contract"`
}

func (h *Handlers) Mint(w http.ResponseWriter, r *http.Request) {
	var req MintRequest
	err := h.readJSON(w, r, &req)
	if err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	log.Printf("[nft] Mint request: %+v", req)

	resp, err := h.svc.Mint(req)
	if err != nil {
		h.errorJSON(w, fmt.Errorf("mint failed: %w", err), http.StatusInternalServerError)
		return
	}

	h.writeJSON(w, http.StatusOK, resp)
}

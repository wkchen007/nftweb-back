package nft

import (
	"fmt"
	"log"
	"net/http"

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

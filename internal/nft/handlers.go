package nft

import (
	"fmt"
	"math/big"
	"net/http"
)

type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

// OwnerOfHandler 查詢 NFT 擁有者
func (h *Handlers) OwnerOfHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TokenID string `json:"tokenId"`
	}
	err := h.readJSON(w, r, &req)
	if err != nil {
		h.errorJSON(w, err, http.StatusBadRequest)
		return
	}
	_, ok := new(big.Int).SetString(req.TokenID, 10)
	if !ok {
		h.errorJSON(w, fmt.Errorf("invalid tokenId"), http.StatusBadRequest)
		return
	}

	//todo: call service
}

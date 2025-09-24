package ethcli

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
)

type Address struct {
	Address string `json:"address"`
	Network string `json:"network"`
}

type Wallet struct {
	Address    string `json:"address"`
	Balance    string `json:"balance"`
	BalanceEth string `json:"balanceEth"`
	Network    string `json:"network"`
}

type TransferRequest struct {
	To          string `json:"to"`        // 目標地址
	AmountEther string `json:"amountEth"` // 十進位字串，例如 "0.01"
}

type TransferResponse struct {
	From        string `json:"from"`
	To          string `json:"to"`
	ValueWei    string `json:"valueWei"`
	ValueEther  string `json:"valueEther"`
	TxHash      string `json:"txHash"`
	Network     string `json:"network"`
	ExplorerUrl string `json:"explorerUrl"`
}

func (c *Client) BuildTxURL(hash string) string {
	base := "https://etherscan.io"
	switch c.chainID.Int64() {
	case 1:
		base = "https://etherscan.io"
	case 5:
		base = "https://goerli.etherscan.io"
	case 11155111:
		base = "https://sepolia.etherscan.io"
	case 137:
		base = "https://polygonscan.com"
	}
	return fmt.Sprintf("%s/tx/%s", base, hash)
}

func (c *Client) GetAddress() Address {

	return Address{
		Address: c.from.Hex(),
		Network: c.network,
	}
}

// GetBalance 取得最新區塊的 ETH 餘額（wei）
func (c *Client) GetBalance() (Wallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	balWei, err := c.backend.BalanceAt(ctx, c.from, nil)
	if err != nil {
		return Wallet{}, err
	}

	wallet := Wallet{
		Address:    c.from.Hex(),
		Balance:    balWei.String(),
		BalanceEth: c.WeiToEtherString(balWei),
		Network:    c.network,
	}

	return wallet, nil
}

func (c *Client) TransferETH(req TransferRequest) (TransferResponse, error) {
	toStr := strings.TrimSpace(req.To)
	if !c.IsHexAddress(toStr) {
		return TransferResponse{}, fmt.Errorf("invalid 'to' address")
	}
	to := c.GethHexToAddress(toStr)

	amountWei, err := c.AmountToWei(strings.TrimSpace(req.AmountEther))
	if err != nil {
		return TransferResponse{}, err
	}
	if amountWei.Cmp(big.NewInt(0)) <= 0 {
		return TransferResponse{}, fmt.Errorf("amountEth must be > 0")
	}
	/*
		if c.from == to {
			return TransferResponse{}, fmt.Errorf("cannot transfer to self")
		}
	*/
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	nonce, err := c.backend.PendingNonceAt(ctx, c.from)
	if err != nil {
		return TransferResponse{}, fmt.Errorf("get nonce: %w", err)
	}

	// EIP-1559 建議費率
	tip, err := c.backend.SuggestGasTipCap(ctx)
	if err != nil {
		return TransferResponse{}, fmt.Errorf("suggest tip: %w", err)
	}
	head, err := c.backend.HeaderByNumber(ctx, nil)
	if err != nil {
		return TransferResponse{}, fmt.Errorf("get head: %w", err)
	}
	baseFee := big.NewInt(0)
	if head.BaseFee != nil {
		baseFee = new(big.Int).Set(head.BaseFee)
	}
	// 常見做法：feeCap = baseFee*2 + tip
	feeCap := new(big.Int).Add(new(big.Int).Mul(baseFee, big.NewInt(2)), tip)

	// 簡單轉帳 gas 上限固定 21000
	const gasLimit = uint64(21000)

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   new(big.Int).Set(c.chainID),
		Nonce:     nonce,
		GasTipCap: tip,
		GasFeeCap: feeCap,
		Gas:       gasLimit,
		To:        &to,
		Value:     new(big.Int).Set(amountWei),
		Data:      nil,
	})

	signer := types.LatestSignerForChainID(c.chainID)
	signed, err := types.SignTx(tx, signer, c.privKey)
	if err != nil {
		return TransferResponse{}, fmt.Errorf("sign tx: %w", err)
	}

	if err := c.backend.SendTransaction(ctx, signed); err != nil {
		return TransferResponse{}, fmt.Errorf("send tx: %w", err)
	}

	log.Printf("[ethcli] sent tx %s: %s wei from %s to %s", signed.Hash().Hex(), amountWei.String(), c.from.Hex(), to.Hex())

	return TransferResponse{
		From:        c.from.Hex(),
		To:          to.Hex(),
		ValueWei:    amountWei.String(),
		ValueEther:  c.WeiToEtherString(amountWei),
		TxHash:      signed.Hash().Hex(),
		Network:     c.network,
		ExplorerUrl: c.BuildTxURL(signed.Hash().Hex()),
	}, nil
}

package ethcli

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client 封裝 geth ethclient.Client
type Client struct {
	rpcURL  string
	backend *ethclient.Client
	// 簽名（可為 nil；nil 代表唯讀、不支援轉帳）
	privKey *ecdsa.PrivateKey
	from    gethcommon.Address
	chainID *big.Int
	network string
}

// New 建立並連線到指定 RPC（Sepolia）
func New(rpcURL string) (*Client, error) {
	if rpcURL == "" {
		return nil, fmt.Errorf("rpcURL is empty")
	}
	// ethclient.DialContext 會在 ctx 取消時中止連線嘗試
	backend, err := ethclient.DialContext(context.Background(), rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to dial rpc: %w", err)
	}
	log.Printf("[ethcli] connected to %s", rpcURL)
	return &Client{rpcURL: rpcURL, backend: backend}, nil
}

// 查詢 + 簽名轉帳用（從環境傳入私鑰）
func NewWithSigner(rpcURL, privKeyHex string) (*Client, error) {
	c, err := New(rpcURL)
	if err != nil {
		return nil, err
	}
	if privKeyHex == "" {
		return nil, fmt.Errorf("privKeyHex is empty")
	}
	pk, err := crypto.HexToECDSA(privKeyHex)
	if err != nil {
		return nil, fmt.Errorf("invalid private key: %w", err)
	}
	c.privKey = pk
	c.from = crypto.PubkeyToAddress(pk.PublicKey)
	// 取得 chainID
	chainID, err := c.backend.NetworkID(context.Background())
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("get chainID: %w", err)
	}
	c.chainID = chainID

	// 取得網路名稱
	switch c.chainID.Uint64() {
	case 1:
		c.network = "Mainnet"
	case 10:
		c.network = "Optimism"
	case 137:
		c.network = "Polygon"
	case 11155111:
		c.network = "Sepolia"
	default:
		c.network = "unknown"
	}

	log.Printf("[ethcli] signer address: %s", c.from.Hex())
	log.Printf("[ethcli] network chainID: %s", c.chainID.String())

	return c, nil
}

// Close 關閉底層連線
func (c *Client) Close() error {
	if c.backend != nil {
		c.backend.Close()
	}
	return nil
}

func (c *Client) NewTransactor(ctx context.Context) (*bind.TransactOpts, error) {
	opts, err := bind.NewKeyedTransactorWithChainID(c.privKey, c.chainID)
	if err != nil {
		return nil, err
	}
	// 極簡策略：
	// - 不手動設 Nonce（讓 bind 自動抓 pending nonce）
	// - 不手動設 GasPrice / GasTipCap / GasFeeCap（讓 bind 自動 Suggest）
	// - 不手動設 GasLimit（讓 bind 自動 EstimateGas）
	opts.Context = ctx
	return opts, nil
}

func (c *Client) Backend() *ethclient.Client {
	return c.backend
}

func (c *Client) WeiToEtherString(wei *big.Int) string {
	// ether = wei / 1e18，以 18 位小數輸出
	num := new(big.Rat).SetInt(wei)
	den := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	return new(big.Rat).Quo(num, den).FloatString(18)
}

func (c *Client) IsHexAddress(s string) bool {
	return gethcommon.IsHexAddress(s)
}

func (c *Client) GethHexToAddress(s string) gethcommon.Address {
	return gethcommon.HexToAddress(s)
}

func (c *Client) AmountToWei(amountEthStr string) (*big.Int, error) {
	oneEthWei := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)

	if amountEthStr != "" {
		r, ok := new(big.Rat).SetString(amountEthStr)
		if !ok {
			return nil, fmt.Errorf("invalid amount_ether")
		}
		weiRat := new(big.Rat).Mul(r, new(big.Rat).SetInt(oneEthWei))
		wei := new(big.Int)
		wei.Div(weiRat.Num(), weiRat.Denom()) // 取整（捨去小數）
		return wei, nil
	}
	return nil, fmt.Errorf("amount_ether is required")
}

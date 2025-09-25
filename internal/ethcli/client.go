package ethcli

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Client 封裝 geth ethclient.Client
type Client struct {
	rpcURL  string
	backend *ethclient.Client

	mu      sync.RWMutex
	privKey *ecdsa.PrivateKey
	from    gethcommon.Address
	chainID *big.Int
	network string
}

// New 只建立連線（唯讀）；之後可用 UseSigner 設定/切換錢包
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

// UseSigner 以新的私鑰切換目前 signer（thread-safe）
func (c *Client) UseSigner(req UseSignerRequest) error {
	if strings.TrimSpace(req.PrivateKey) == "" {
		return fmt.Errorf("privKeyHex is empty")
	}
	pk, err := crypto.HexToECDSA(trim0x(req.PrivateKey))
	if err != nil {
		return fmt.Errorf("invalid private key: %w", err)
	}

	// 先查 network/chainID（不鎖；讀 backend thread-safe）
	chainID, err := c.backend.NetworkID(context.Background())
	if err != nil {
		return fmt.Errorf("get chainID: %w", err)
	}

	network := networkName(chainID)

	// 寫入受保護欄位
	c.mu.Lock()
	c.privKey = pk
	c.from = crypto.PubkeyToAddress(pk.PublicKey)
	c.chainID = new(big.Int).Set(chainID)
	c.network = network
	c.mu.Unlock()

	log.Printf("[ethcli] signer address: %s", c.From().Hex())
	log.Printf("[ethcli] network chainID: %s (%s)", c.ChainID().String(), c.Network())
	return nil
}

// Close 關閉底層連線
func (c *Client) Close() error {
	if c.backend != nil {
		c.backend.Close()
	}
	return nil
}

func (c *Client) Backend() bind.ContractBackend { return c.backend }

func (c *Client) HasSigner() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.privKey != nil && c.chainID != nil
}

func (c *Client) From() gethcommon.Address {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.from
}

func (c *Client) ChainID() *big.Int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.chainID == nil {
		return nil
	}
	return new(big.Int).Set(c.chainID)
}

func (c *Client) Network() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.network
}

// NewTransactor 依目前 signer 產生帶 context 的 TransactOpts
func (c *Client) NewTransactor(ctx context.Context) (*bind.TransactOpts, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.privKey == nil || c.chainID == nil {
		return nil, fmt.Errorf("client has no signer")
	}
	opts, err := bind.NewKeyedTransactorWithChainID(c.privKey, c.chainID)
	if err != nil {
		return nil, err
	}
	opts.Context = ctx
	return opts, nil
}

func WeiToEtherString(wei *big.Int) string {
	num := new(big.Rat).SetInt(wei)
	den := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	return new(big.Rat).Quo(num, den).FloatString(18)
}

func AmountToWei(amountEthStr string) (*big.Int, error) {
	if strings.TrimSpace(amountEthStr) == "" {
		return nil, fmt.Errorf("amount_ether is required")
	}
	oneEthWei := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	r, ok := new(big.Rat).SetString(amountEthStr)
	if !ok {
		return nil, fmt.Errorf("invalid amount_ether")
	}
	weiRat := new(big.Rat).Mul(r, new(big.Rat).SetInt(oneEthWei))
	wei := new(big.Int)
	wei.Div(weiRat.Num(), weiRat.Denom()) // 取整（捨去小數）
	return wei, nil
}

func IsHexAddress(s string) bool                   { return gethcommon.IsHexAddress(s) }
func GethHexToAddress(s string) gethcommon.Address { return gethcommon.HexToAddress(s) }

func trim0x(s string) string {
	if len(s) >= 2 && (s[:2] == "0x" || s[:2] == "0X") {
		return s[2:]
	}
	return s
}

func networkName(id *big.Int) string {
	switch id.Uint64() {
	case 1:
		return "Mainnet"
	case 10:
		return "Optimism"
	case 137:
		return "Polygon"
	case 11155111:
		return "Sepolia"
	default:
		return "unknown"
	}
}

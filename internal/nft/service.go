package nft

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/wkchen007/nftweb-back/internal/ethcli"
	"github.com/wkchen007/nftweb-back/internal/models"
	"github.com/wkchen007/nftweb-back/internal/repository"
)

type Service struct {
	client    *ethcli.Client
	abi       abi.ABI
	contract  gethcommon.Address
	conTxHash gethcommon.Hash
	config    *Config
	DB        repository.DatabaseRepo
}

func loadABIFromFile(path string) (abi.ABI, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return abi.ABI{}, err
	}
	return abi.JSON(strings.NewReader(string(b)))
}

func NewServiceFromConfig(client *ethcli.Client, cfg *Config) (*Service, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil config")
	}
	return NewServiceWithABIPath(client, cfg.NFT.ABIPath, cfg.NFT.ContractAddress, cfg.NFT.ContractTxHash, cfg)
}

// NewServiceWithABIPath：顯式指定 ABI 路徑與合約地址（方便測試）
func NewServiceWithABIPath(client *ethcli.Client, abiPath, contractAddr string, conTxHash string, cfg *Config) (*Service, error) {
	if strings.TrimSpace(abiPath) == "" {
		return nil, fmt.Errorf("abiPath is empty")
	}
	if !gethcommon.IsHexAddress(contractAddr) {
		return nil, fmt.Errorf("invalid contract address")
	}
	if !client.IsTxHex(conTxHash) {
		return nil, fmt.Errorf("invalid contract tx hash")
	}
	parsed, err := loadABIFromFile(abiPath)
	if err != nil {
		return nil, err
	}
	return &Service{
		client:    client,
		abi:       parsed,
		contract:  gethcommon.HexToAddress(contractAddr),
		conTxHash: gethcommon.HexToHash(conTxHash),
		config:    cfg,
	}, nil
}

func (s *Service) ConCreator() (OwnerResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	//查看交易內容
	backend := s.client.Backend()
	tx, _, err := backend.TransactionByHash(ctx, s.conTxHash)
	if err != nil {
		return OwnerResponse{}, fmt.Errorf("failed to get contract creation tx: %w", err)
	}

	// 從交易裡還原 sender
	signer := types.LatestSignerForChainID(s.client.ChainID())
	creator, err := types.Sender(signer, tx)
	if err != nil {
		return OwnerResponse{}, fmt.Errorf("failed to recover sender")
	}

	return OwnerResponse{
		Contract: s.contract.Hex(),
		Owner:    creator.Hex(),
	}, nil
}

func (s *Service) OwnerOf(req OwnerOfRequest) (OwnerOfResponse, error) {
	contract := gethcommon.HexToAddress(req.Contract)
	if contract != s.contract {
		return OwnerOfResponse{}, fmt.Errorf("unsupported contract address")
	}

	tokenId, ok := new(big.Int).SetString(req.TokenID, 10)
	if !ok {
		return OwnerOfResponse{}, fmt.Errorf("invalid tokenId")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	bound := bind.NewBoundContract(s.contract, s.abi, s.client.ConBackend(), nil, nil)
	var out []interface{}
	if err := bound.Call(&bind.CallOpts{Context: ctx}, &out, "ownerOf", tokenId); err != nil {
		return OwnerOfResponse{}, err
	}
	if len(out) == 0 {
		return OwnerOfResponse{}, fmt.Errorf("no result")
	}
	addr, ok := out[0].(gethcommon.Address)
	if !ok {
		return OwnerOfResponse{}, fmt.Errorf("invalid return type")
	}
	return OwnerOfResponse{
		Contract: req.Contract,
		TokenID:  req.TokenID,
		Owner:    addr.Hex(),
	}, nil
}

// 打包、估算、簽名並送出 EIP-1559 交易（使用 newTransactor 設好的 tip/feecap）
func (s *Service) sendTx(ctx context.Context, method string, args ...interface{}) (txHash *gethcommon.Hash, err error) {
	// calldata
	data, err := s.abi.Pack(method, args...)
	if err != nil {
		return nil, err
	}

	// opts（已含 Nonce / GasTipCap / GasFeeCap）
	backend := s.client.ConBackend()
	opts, err := s.client.NewTransactor(ctx)
	if err != nil {
		return nil, err
	}

	// 估 gasLimit，使用 EIP-1559 欄位
	call := ethereum.CallMsg{From: opts.From, To: &s.contract, Data: data, Value: big.NewInt(0), GasFeeCap: opts.GasFeeCap, GasTipCap: opts.GasTipCap}
	gasLimit, gasErr := backend.EstimateGas(ctx, call)
	if gasErr != nil {
		return nil, fmt.Errorf("estimate gas: %w", gasErr)
	}
	opts.GasLimit = gasLimit

	bound := bind.NewBoundContract(s.contract, s.abi, backend, backend, backend)

	tx, err := bound.Transact(opts, method, args...)
	if err != nil {
		return nil, err
	}

	log.Printf("[nft] %s tx sent: %s contract: %s from: %s", method, tx.Hash().Hex(), s.contract.Hex(), opts.From.Hex())

	hash := tx.Hash()
	return &hash, nil
}

func (s *Service) OpenBlindBox() (ConResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	hash, err := s.sendTx(ctx, "openBlindBox")
	if err != nil {
		return ConResponse{}, fmt.Errorf("openBlindBox failed: %w", err)
	}

	return ConResponse{
		TxHash:   hash.Hex(),
		Contract: s.contract.Hex(),
	}, nil
}

func (s *Service) Mint(req MintRequest) (MintResponse, error) {
	contract := gethcommon.HexToAddress(req.Contract)
	if contract != s.contract {
		return MintResponse{}, fmt.Errorf("unsupported contract address")
	}

	ok := ethcli.IsHexAddress(req.To)
	if !ok {
		return MintResponse{}, fmt.Errorf("invalid to address")
	}
	to := ethcli.GethHexToAddress(req.To)
	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		return MintResponse{}, fmt.Errorf("invalid amount")
	}

	valueWei, err := ethcli.AmountToWei(req.ValueETH)
	if err != nil {
		return MintResponse{}, fmt.Errorf("invalid valueEth: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts, err := s.client.NewTransactor(ctx)
	if err != nil {
		return MintResponse{}, err
	}
	opts.Value = valueWei

	backend := s.client.ConBackend()
	bound := bind.NewBoundContract(s.contract, s.abi, backend, backend, backend)

	if data, packErr := s.abi.Pack("mint", to, amount); packErr == nil {
		call := ethereum.CallMsg{From: opts.From, To: &s.contract, Data: data, Value: valueWei}
		_, _ = backend.EstimateGas(ctx, call)
	}

	tx, err := bound.Transact(opts, "mint", to, amount)
	if err != nil {
		return MintResponse{}, err
	}

	log.Printf("[nft] Mint tx sent: %s contract: %s from: %s", tx.Hash().Hex(), req.Contract, opts.From.Hex())
	return MintResponse{
		TxHash:   tx.Hash().Hex(),
		From:     opts.From.Hex(),
		Contract: req.Contract,
	}, nil
}

// Counter 讀取合約的 counter()，通常代表已鑄出的數量上限（或已經 mint 的總數）
func (s *Service) Counter() (*big.Int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	backend := s.client.ConBackend()
	bound := bind.NewBoundContract(s.contract, s.abi, backend, nil, nil)

	var out []interface{}
	if err := bound.Call(&bind.CallOpts{Context: ctx}, &out, "counter"); err != nil {
		return nil, fmt.Errorf("counter call failed: %w", err)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("counter returned no result")
	}

	switch v := out[0].(type) {
	case *big.Int:
		return v, nil
	case big.Int:
		return new(big.Int).Set(&v), nil
	default:
		return nil, fmt.Errorf("unexpected counter return type: %T", out[0])
	}
}

// TokenURI 讀取合約指定 tokenId 的 tokenURI
func (s *Service) TokenURI(tokenID *big.Int) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	backend := s.client.ConBackend()
	bound := bind.NewBoundContract(s.contract, s.abi, backend, nil, nil)

	var out []interface{}
	if err := bound.Call(&bind.CallOpts{Context: ctx}, &out, "tokenURI", tokenID); err != nil {
		return "", fmt.Errorf("tokenURI call failed: %w", err)
	}
	if len(out) == 0 {
		return "", fmt.Errorf("tokenURI returned no result")
	}

	uri, ok := out[0].(string)
	if !ok {
		return "", fmt.Errorf("unexpected tokenURI return type: %T", out[0])
	}
	return uri, nil
}

// TokensOfOwner 線性掃描 ownerOf 取得某地址擁有的 tokenIds（因合約未提供 Enumerable）。
// maxScan<=0 時，優先使用 config.NFT.MaxScanTokenID；若也未設定，預設 1000。
func (s *Service) TokensOfOwner(req TokensOfOwnerRequest) (TokensOfOwnerResponse, error) {
	if !gethcommon.IsHexAddress(req.Owner) {
		return TokensOfOwnerResponse{}, fmt.Errorf("invalid address")
	}
	owner := gethcommon.HexToAddress(req.Owner)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	backend := s.client.ConBackend()
	bound := bind.NewBoundContract(s.contract, s.abi, backend, nil, nil)

	// 1. 先問合約目前的 counter
	total, err := s.Counter()
	if err != nil {
		return TokensOfOwnerResponse{}, fmt.Errorf("get counter: %w", err)
	}

	// 2. 線性掃描 ownerOf
	ids := []*big.Int{}
	for i := int64(0); i <= total.Int64(); i++ {
		tokenID := big.NewInt(i)

		var out []interface{}
		if err := bound.Call(&bind.CallOpts{Context: ctx}, &out, "ownerOf", tokenID); err != nil {
			// token 尚未被 mint，會 revert → 忽略錯誤繼續
			continue
		}
		if len(out) == 0 {
			continue
		}
		addr, ok := out[0].(gethcommon.Address)
		if !ok {
			continue
		}
		if addr == owner {
			ids = append(ids, tokenID)
		}
	}
	log.Printf("[nft] TokensOfOwner found tokens: %+v owned by %s", ids, owner.Hex())

	//3.找尋TokenURI（如果需要）
	intIDs := make([]int, 0, len(ids))
	for _, id := range ids {
		intIDs = append(intIDs, int(id.Int64()))
	}
	items := make([]models.TokenItem, 0, len(intIDs))
	if req.IncludeTokenURI {
		now := time.Now()
		deadline := time.Date(2025, 10, 1, 0, 0, 0, 0, time.Local) // 2025年10月1日
		if now.Before(deadline) {
			box, err := s.DB.GetBoxItem()
			if err != nil {
				return TokensOfOwnerResponse{}, fmt.Errorf("GetBoxItem: %w", err)
			}
			for _, id := range intIDs {
				items = append(items, models.TokenItem{TokenID: strconv.Itoa(id), TokenURI: box.TokenURI, ImageURI: box.ImageURI})
			}
		} else {
			items, err = s.DB.GetTokenItem(intIDs)
			if err != nil {
				return TokensOfOwnerResponse{}, fmt.Errorf("GetTokenItem: %w", err)
			}
		}
	} else {
		for _, id := range intIDs {
			items = append(items, models.TokenItem{TokenID: strconv.Itoa(id)})
		}
	}

	resp := TokensOfOwnerResponse{
		Owner:  owner.Hex(),
		Count:  len(items),
		Tokens: items,
	}
	return resp, nil
}

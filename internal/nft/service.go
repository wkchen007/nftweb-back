package nft

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/v2"
	gethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/wkchen007/nftweb-back/internal/ethcli"
)

type Service struct {
	client   *ethcli.Client
	abi      abi.ABI
	contract gethcommon.Address
	config   *Config
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
	return NewServiceWithABIPath(client, cfg.NFT.ABIPath, cfg.NFT.ContractAddress, cfg)
}

// NewServiceWithABIPath：顯式指定 ABI 路徑與合約地址（方便測試）
func NewServiceWithABIPath(client *ethcli.Client, abiPath, contractAddr string, cfg *Config) (*Service, error) {
	if strings.TrimSpace(abiPath) == "" {
		return nil, fmt.Errorf("abiPath is empty")
	}
	if !gethcommon.IsHexAddress(contractAddr) {
		return nil, fmt.Errorf("invalid contract address")
	}
	parsed, err := loadABIFromFile(abiPath)
	if err != nil {
		return nil, err
	}
	return &Service{
		client:   client,
		abi:      parsed,
		contract: gethcommon.HexToAddress(contractAddr),
		config:   cfg,
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

	bound := bind.NewBoundContract(s.contract, s.abi, s.client.Backend(), nil, nil)
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

func (s *Service) Mint(req MintRequest) (MintResponse, error) {
	contract := gethcommon.HexToAddress(req.Contract)
	if contract != s.contract {
		return MintResponse{}, fmt.Errorf("unsupported contract address")
	}

	ok := s.client.IsHexAddress(req.To)
	if !ok {
		return MintResponse{}, fmt.Errorf("invalid to address")
	}
	to := s.client.GethHexToAddress(req.To)
	amount, ok := new(big.Int).SetString(req.Amount, 10)
	if !ok {
		return MintResponse{}, fmt.Errorf("invalid amount")
	}

	valueWei, err := s.client.AmountToWei(req.ValueETH)
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

	backend := s.client.Backend()
	bound := bind.NewBoundContract(s.contract, s.abi, backend, backend, backend)

	// 自動估 gas
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

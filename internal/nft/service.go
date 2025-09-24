package nft

import (
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
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

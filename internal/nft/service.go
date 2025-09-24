package nft

import (
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
}

func loadABIFromFile(path string) (abi.ABI, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return abi.ABI{}, err
	}
	return abi.JSON(strings.NewReader(string(b)))
}

func NewService(client *ethcli.Client, abiPath, contractAddr string) (*Service, error) {
	parsed, err := loadABIFromFile(abiPath)
	if err != nil {
		return nil, err
	}
	return &Service{
		client:   client,
		abi:      parsed,
		contract: gethcommon.HexToAddress(contractAddr),
	}, nil
}

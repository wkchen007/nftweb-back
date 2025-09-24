package nft

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	NFT struct {
		ContractAddress string `yaml:"contractAddress"`
		ABIPath         string `yaml:"abiPath"`
		MaxScanTokenID  int64  `yaml:"maxScanTokenID"`
	} `yaml:"nft"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

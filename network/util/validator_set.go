package util

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
)

type Validator struct {
	NodeID     string `toml:"node_id" json:"node_id"`
	Stake      string `toml:"stake" json:"stake"`
	CertPubkey string `toml:"cert_pubkey" json:"cert_pubkey"`
}

type ValidatorSet struct {
	Epoch      int64       `toml:"epoch" json:"epoch"`
	Validators []Validator `toml:"validators" json:"validators"`
}

type ValidatorsConfig struct {
	ValidatorSets []ValidatorSet `toml:"validator_sets"`
}

func LoadValidatorsConfig() (*ValidatorsConfig, error) {
	if err := godotenv.Load(); err != nil {
	}

	filePath := os.Getenv("VALIDATORS_FILE")
	if filePath == "" {
		return nil, fmt.Errorf("VALIDATORS_FILE environment variable not set")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read validators file %s: %w", filePath, err)
	}

	var config ValidatorsConfig
	if err := toml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal TOML data from %s: %w", filePath, err)
	}

	return &config, nil
}

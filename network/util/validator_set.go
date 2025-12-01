package util

import (
	"fmt"
	"math/big"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/joho/godotenv"
)

type Stake struct {
	*big.Int
}

func (s *Stake) UnmarshalText(text []byte) error {
	stakeStr := string(text)

	if s.Int == nil {
		s.Int = new(big.Int)
	}

	stakeBigInt, success := s.Int.SetString(stakeStr, 0)
	if !success {
		return fmt.Errorf("invalid hex stake value: %s", stakeStr)
	}

	if stakeBigInt.Sign() < 0 {
		return fmt.Errorf("stake value is negative: %s", stakeStr)
	}

	return nil
}

type Validator struct {
	NodeID     string `toml:"node_id" json:"node_id"`
	Stake      Stake  `toml:"stake" json:"stake"`
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

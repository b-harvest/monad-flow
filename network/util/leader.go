package util

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"
	"math/bits"
	"sort"

	"golang.org/x/crypto/chacha20"
)

type weightedValidator struct {
	Original Validator
	Stake    *big.Int
}

const (
	PCG_MUL = 6364136223846793005
	PCG_INC = 11634580027462260723
)

func pcg32(state *uint64) []byte {
	*state = (*state * PCG_MUL) + PCG_INC
	currentState := *state

	xorshifted := uint32(((currentState >> 18) ^ currentState) >> 27)
	rot := uint32(currentState >> 59)
	x := bits.RotateLeft32(xorshifted, -int(rot))

	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, x)
	return bytes
}

type RustRng struct {
	cipher *chacha20.Cipher
}

func NewRustRngFromRound(round uint64) *RustRng {
	state := round
	key := make([]byte, 32)
	for i := 0; i < 32; i += 4 {
		chunk := pcg32(&state)
		copy(key[i:], chunk)
	}

	nonce := make([]byte, 12)
	cipher, err := chacha20.NewUnauthenticatedCipher(key, nonce)
	if err != nil {
		panic(fmt.Sprintf("failed to create cipher: %v", err))
	}

	return &RustRng{cipher: cipher}
}

func (r *RustRng) GenU256() *big.Int {
	buf := make([]byte, 32)
	r.cipher.XORKeyStream(buf, buf)
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	return new(big.Int).SetBytes(buf)
}

var MaxUint256 = new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 256), big.NewInt(1))

func randomize256WithRng(rng *RustRng, m *big.Int) *big.Int {
	if m.Sign() <= 0 {
		return big.NewInt(0)
	}

	temp := new(big.Int).Sub(MaxUint256, m)
	temp.Add(temp, big.NewInt(1))
	rem := new(big.Int).Mod(temp, m)
	maxBound := new(big.Int).Sub(MaxUint256, rem)

	for {
		r := rng.GenU256()
		if r.Cmp(maxBound) <= 0 {
			return new(big.Int).Mod(r, m)
		}
	}
}

func GetLeader(round uint64, validators []Validator) (Validator, error) {
	if len(validators) == 0 {
		return Validator{}, errors.New("validator set is empty")
	}

	validList := make([]weightedValidator, 0, len(validators))
	totalStake := new(big.Int)

	for _, v := range validators {
		stakeInt := new(big.Int)

		if _, ok := stakeInt.SetString(v.Stake, 0); !ok {
			return Validator{}, fmt.Errorf("invalid stake format for node %s: %s", v.NodeID, v.Stake)
		}

		if stakeInt.Sign() > 0 {
			validList = append(validList, weightedValidator{
				Original: v,
				Stake:    stakeInt,
			})
			totalStake.Add(totalStake, stakeInt)
		}
	}

	if len(validList) == 0 {
		return Validator{}, errors.New("no validators with positive stake")
	}

	sort.Slice(validList, func(i, j int) bool {
		return validList[i].Original.NodeID < validList[j].Original.NodeID
	})

	rng := NewRustRngFromRound(round)
	targetVal := randomize256WithRng(rng, totalStake)

	currentSum := new(big.Int)
	for _, wv := range validList {
		currentSum.Add(currentSum, wv.Stake)

		if currentSum.Cmp(targetVal) > 0 {
			return wv.Original, nil
		}
	}

	return validList[len(validList)-1].Original, nil
}

package util

import (
	"encoding/hex"
	"errors"
	"fmt"

	"monad-flow/model"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/zeebo/blake3"
)

type SenderInfo struct {
	PubKey []byte
	NodeID string
}

func RecoverSenderHybrid(chunk *model.MonadChunkPacket, rawData []byte) (*SenderInfo, error) {
	if len(rawData) < HeaderFullLen {
		return nil, errors.New("packet too short")
	}
	headerBody := rawData[SignatureSize:HeaderFullLen]
	proofCount := int(chunk.MerkleTreeDepth) - 1
	proofLen := proofCount * MerkleHashLen
	chunkStart := HeaderFullLen + proofLen

	if len(rawData) < chunkStart {
		return nil, errors.New("packet too short for proof")
	}
	chunkPayload := rawData[chunkStart:]
	leafHashFull := blake3.Sum256(chunkPayload)
	currentHash := leafHashFull[:MerkleHashLen]
	if chunk.MerkleTreeDepth == 0 {
		return nil, errors.New("invalid merkle depth 0")
	}
	numLeaves := 1 << (chunk.MerkleTreeDepth - 1)
	currentTreeIdx := (numLeaves - 1) + int(chunk.MerkleLeafIdx)

	proofs := chunk.MerkleProof
	if len(proofs) != proofCount {
		return nil, fmt.Errorf("proof count mismatch: expected %d, got %d", proofCount, len(proofs))
	}

	for i := len(proofs) - 1; i >= 0; i-- {
		sibling := proofs[i]

		hasher := blake3.New()
		if currentTreeIdx%2 == 1 {
			hasher.Write(currentHash)
			hasher.Write(sibling)
		} else {
			hasher.Write(sibling)
			hasher.Write(currentHash)
		}

		digest := hasher.Sum(nil)
		currentHash = digest[:MerkleHashLen]

		currentTreeIdx = (currentTreeIdx - 1) / 2
	}
	computedRoot := currentHash
	msgPayload := make([]byte, 0, len(headerBody)+len(computedRoot))
	msgPayload = append(msgPayload, headerBody...)
	msgPayload = append(msgPayload, computedRoot...)

	signingInput := make([]byte, 0, len(MonadSigningPrefix)+len(msgPayload))
	signingInput = append(signingInput, []byte(MonadSigningPrefix)...)
	signingInput = append(signingInput, msgPayload...)

	sighash := blake3.Sum256(signingInput)

	signature := make([]byte, 65)
	copy(signature, chunk.Signature[:])

	if signature[64] >= 27 {
		signature[64] -= 27
	}
	pubKey, err := crypto.SigToPub(sighash[:], signature)
	if err != nil {
		return nil, fmt.Errorf("signature recovery failed: %w", err)
	}
	compressedPubKey := crypto.CompressPubkey(pubKey)

	return &SenderInfo{
		PubKey: compressedPubKey,
		NodeID: hex.EncodeToString(compressedPubKey),
	}, nil
}

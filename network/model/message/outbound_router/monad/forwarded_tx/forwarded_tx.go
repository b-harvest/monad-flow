package forwarded_tx

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type ForwardedTxMessage []*types.Transaction

func DecodeForwardedTxMessage(b []byte) (*ForwardedTxMessage, error) {
	s := rlp.NewStream(bytes.NewReader(b), uint64(len(b)))
	if _, err := s.List(); err != nil {
		return nil, fmt.Errorf("data is not an RLP list (expected [tx, ...]): %w", err)
	}

	var txs ForwardedTxMessage
	for {
		kind, _, err := s.Kind()
		if err != nil {
			if errors.Is(err, rlp.EOL) {
				break
			}
			return nil, fmt.Errorf("failed to peek element kind: %w", err)
		}
		var txData []byte
		if kind == rlp.List {
			txData, err = s.Raw()
		} else {
			txData, err = s.Bytes()
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read list item: %w", err)
		}

		realTxData, err := unwrapRLPBytesRecursively(txData)
		if err != nil {
			return nil, fmt.Errorf("failed to unwrap tx layers: %w", err)
		}

		var tx types.Transaction
		if err := tx.UnmarshalBinary(realTxData); err != nil {
			return nil, fmt.Errorf("failed to decode transaction item (header: %x): %w", realTxData[:2], err)
		}
		txs = append(txs, &tx)
	}
	if err := s.ListEnd(); err != nil {
		return nil, fmt.Errorf("list traversal not finished properly: %w", err)
	}
	return &txs, nil
}

func unwrapRLPBytesRecursively(data []byte) ([]byte, error) {
	curr := data
	for {
		if len(curr) == 0 {
			return curr, nil
		}
		firstByte := curr[0]
		if firstByte <= 0x7f {
			return curr, nil
		}
		if firstByte >= 0xf8 {
			return curr, nil
		}
		if firstByte >= 0x80 && firstByte <= 0xbf {
			var inner []byte
			if err := rlp.DecodeBytes(curr, &inner); err != nil {
				return curr, nil
			}
			curr = inner
			continue
		}
		return curr, nil
	}
}

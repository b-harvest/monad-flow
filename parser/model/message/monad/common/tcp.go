package common

import (
	"encoding/binary"
	"fmt"
	"io"
	"monad-flow/util"
)

type TcpMsgHdr struct {
	Magic   uint32
	Version uint32
	Length  uint64
}

type SignedMessage struct {
	Signature []byte
	Payload   []byte
}

func ReadTcpMsgHdr(r io.Reader) (*TcpMsgHdr, error) {
	headerBytes := make([]byte, util.HeaderSize)
	if _, err := io.ReadFull(r, headerBytes); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, fmt.Errorf("L1: failed to read SSNC header: %w", err)
	}

	hdr := &TcpMsgHdr{
		Magic:   binary.LittleEndian.Uint32(headerBytes[0:4]),
		Version: binary.LittleEndian.Uint32(headerBytes[4:8]),
		Length:  binary.LittleEndian.Uint64(headerBytes[8:16]),
	}

	if hdr.Magic != util.HeaderMagic {
		return nil, fmt.Errorf("L1: invalid magic. expected %x, got %x", util.HeaderMagic, hdr.Magic)
	}
	return hdr, nil
}

func ReadSignedMsg(r io.Reader, hdr *TcpMsgHdr) (*SignedMessage, error) {
	if hdr.Length <= util.SignatureSize {
		return nil, fmt.Errorf("L2: payload length (%d) is too short for signature (%d)", hdr.Length, util.SignatureSize)
	}

	payload := make([]byte, hdr.Length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return nil, fmt.Errorf("L2: failed to read payload (expected %d bytes): %w", hdr.Length, err)
	}

	return &SignedMessage{
		Signature: payload[:util.SignatureSize],
		Payload:   payload[util.SignatureSize:],
	}, nil
}

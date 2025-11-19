package common

import (
	"fmt"
	"monad-flow/util"

	"github.com/ethereum/go-ethereum/rlp"
)

type RoundCertificateWrapper struct {
	Certificate RoundCertificate
}

type RoundCertificate interface {
	isRoundCertificate()
}

type RoundCertificateQC struct {
	QC *QuorumCertificate
}

type RoundCertificateTC struct {
	TC *TimeoutCertificate
}

func (r *RoundCertificateQC) isRoundCertificate() {}
func (r *RoundCertificateTC) isRoundCertificate() {}

func (r *RoundCertificateWrapper) DecodeRLP(s *rlp.Stream) error {
	if _, err := s.List(); err != nil {
		return err
	}

	typeID, err := s.Uint8()
	if err != nil {
		return err
	}

	var cert RoundCertificate
	switch typeID {
	case util.QC:
		cert = new(RoundCertificateQC)
	case util.TC:
		cert = new(RoundCertificateTC)
	default:
		return fmt.Errorf("unknown RoundCertificate Type ID: %d", typeID)
	}

	if err := s.Decode(cert); err != nil {
		return fmt.Errorf("failed to decode certificate payload: %w", err)
	}

	r.Certificate = cert
	return nil
}

func (r *RoundCertificateQC) DecodeRLP(s *rlp.Stream) error {
	if r.QC == nil {
		r.QC = new(QuorumCertificate)
	}
	return s.Decode(r.QC)
}

func (r *RoundCertificateTC) DecodeRLP(s *rlp.Stream) error {
	if r.TC == nil {
		r.TC = new(TimeoutCertificate)
	}
	return s.Decode(r.TC)
}

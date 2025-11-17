package decoder

import (
	"context"
	"errors"
	"io"
	"log"
	"monad-flow/model/message/monad/common"
	"monad-flow/parser"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

const readTimeout = 10 * time.Second

type MonadTcpStreamFactory struct {
	Ctx context.Context
}

func (f *MonadTcpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	s := &MonadTcpStream{
		net:       net,
		transport: transport,
		r:         tcpreader.NewReaderStream(),
		ctx:       f.Ctx,
	}

	go s.run()

	return &s.r
}

type MonadTcpStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
	ctx            context.Context
}

func (s *MonadTcpStream) run() {
	log.Printf("[TCP Reassembly] start stream handler: %s", s.net.Src())
	defer log.Printf("[TCP Reassembly] finish stream handler: %s", s.net.Src())

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("[TCP Reassembly] Context cancellation signal detected. Stopping stream handler (%s).", s.net.Src())
			return
		default:
		}

		var hdr *common.TcpMsgHdr
		var err error
		hdrChan := make(chan error, 1)

		go func() {
			hdr, err = common.ReadTcpMsgHdr(&s.r)
			hdrChan <- err
		}()

		select {
		case <-s.ctx.Done():
			return
		case <-time.After(readTimeout):
			log.Printf("[L1] SSNC header read timed out (10s): %s", s.net.Src())
			return
		case err = <-hdrChan:
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF || errors.Is(err, io.ErrClosedPipe) {
					log.Printf("[L1] Stream normally closed (EOF/Closed): %s", s.net.Src())
					return
				}
				log.Printf("[L1] Failed to read/parse SSNC header: %v", err)
				return
			}
		}

		var signedMsg *common.SignedMessage
		payloadChan := make(chan error, 1)

		go func() {
			signedMsg, err = common.ReadSignedMsg(&s.r, hdr)
			payloadChan <- err
		}()

		select {
		case <-s.ctx.Done():
			return
		case <-time.After(readTimeout):
			log.Printf("[REASSEMBLY FAILED] Payload read timed out (10s): Expected %d bytes (Stream: %s)", hdr.Length, s.net.Src())
			return
		case err = <-payloadChan:
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF || errors.Is(err, io.ErrClosedPipe) {
					log.Printf("[L2] Stream closed while reading payload (EOF/Closed): %s", s.net.Src())
				} else {
					log.Printf("[L2] Failed to read/parse payload: %v", err)
				}
				return
			}
		}
		if err := parser.HandleDecodedMessage(signedMsg.Payload); err != nil {
			log.Printf("[L3-L5] Message handler error: %v", err)
		}
	}
}

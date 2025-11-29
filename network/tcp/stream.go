package tcp

import (
	"context"
	"errors"
	"io"
	"log"
	"monad-flow/model/message/outbound_router/monad/common"
	"monad-flow/parser"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"

	"github.com/zishang520/socket.io/clients/socket/v3"
)

const readTimeout = 10 * time.Second

type MonadTcpStreamFactory struct {
	Ctx         context.Context
	Client      *socket.Socket
	ClientMutex *sync.Mutex
}

func (f *MonadTcpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	s := &MonadTcpStream{
		net:         net,
		transport:   transport,
		r:           tcpreader.NewReaderStream(),
		ctx:         f.Ctx,
		client:      f.Client,
		clientMutex: f.ClientMutex,
	}

	go s.run()

	return &s.r
}

type MonadTcpStream struct {
	net, transport gopacket.Flow
	r              tcpreader.ReaderStream
	ctx            context.Context
	client         *socket.Socket
	clientMutex    *sync.Mutex
}

func (s *MonadTcpStream) run() {
	log.Printf("[TCP Reassembly] start stream handler: %s", s.net.Src())
	defer log.Printf("[TCP Reassembly] finish stream handler: %s", s.net.Src())

	for {
		select {
		case <-s.ctx.Done():
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
			return
		case err = <-hdrChan:
			if err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF || errors.Is(err, io.ErrClosedPipe) {
					return
				}
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
			return
		case err = <-payloadChan:
			if err != nil {
				return
			}
		}
		if err := parser.HandleDecodedMessage(signedMsg.Payload, "none"); err != nil {
			log.Printf("[L3-L5] Message handler error: %v", err)
		}
	}
}

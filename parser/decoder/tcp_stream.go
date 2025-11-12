package decoder

import (
	"context"
	"errors"
	"io"
	"log"
	"monad-flow/model/message/monad/common"
	"monad-flow/parser"

	"github.com/google/gopacket"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
)

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
	log.Printf("[TCP Reassembly] 스트림 핸들러 시작: %s", s.net.Src())
	defer log.Printf("[TCP Reassembly] 스트림 핸들러 종료: %s", s.net.Src())

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("[TCP Reassembly] Context 종료 신호 감지. 스트림 핸들러(%s) 종료.", s.net.Src())
			return
		default:
		}

		hdr, err := common.ReadTcpMsgHdr(&s.r)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF || errors.Is(err, io.ErrClosedPipe) {
				log.Printf("[L1] 스트림 정상 종료 (EOF/Closed): %s", s.net.Src())
				return
			}
			log.Printf("[L1] SSNC 헤더 읽기/파싱 실패: %v", err)
			return
		}

		signedMsg, err := common.ReadSignedMsg(&s.r, hdr)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF || errors.Is(err, io.ErrClosedPipe) {
				log.Printf("[L2] 페이로드 읽기 중 스트림 종료 (EOF/Closed): %s", s.net.Src())
			} else {
				log.Printf("[L2] 페이로드 읽기/파싱 실패: %v", err)
			}
			return
		}
		if err := parser.HandleDecodedMessage(signedMsg.Payload); err != nil {
			log.Printf("[L3-L5] 메시지 핸들러 에러: %v", err)
		}
	}
}

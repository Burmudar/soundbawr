package server

// The following is HEAVILY inspired from hashicorp/memberlist
// https://github.com/hashicorp/memberlist/blob/master/net_transport.go

import (
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	UDP_PACKET_SIZE = 256
)

type Packet struct {
	Src       string
	Raw       []byte
	Timestamp time.Time
}

type Transport interface {
	Shutdown() error
	RecvCh() chan *Packet
	Send(raw []byte) (time.Time, error)
}

type NetTransport struct {
	wg       *sync.WaitGroup
	target   string
	recvCh   chan *Packet
	udp      *net.UDPConn
	shutdown int32
	Logger   *log.Logger
}

func NewNetTransport(address, target string, logger *log.Logger) (Transport, error) {
	addr, err := net.ResolveUDPAddr("udp", address)

	if err != nil {
		logger.Printf("[ERR] Failed to resolve listening address '%s': %v", addr, err)
		return nil, err
	}

	udp, err := net.ListenUDP("udp", addr)
	if err != nil {
		logger.Printf("[ERR] Failed to listen on address '%s': %v", addr, err)
		return nil, err
	}

	tp := &NetTransport{
		target: target,
		recvCh: make(chan *Packet),
		Logger: logger,
	}

	tp.wg.Add(1)
	go tp.listen(udp)

	return tp, nil
}

func (tp *NetTransport) listen(conn *net.UDPConn) {
	defer tp.wg.Done()

	for {
		buff := make([]byte, UDP_PACKET_SIZE)
		n, src, err := conn.ReadFrom(buff)

		if err != nil {
			if v := atomic.LoadInt32(&tp.shutdown); v == 1 {
				break
			}
			tp.Logger.Printf("[ERR] Failed to read from UDP connection: %v", err)
			continue
		}

		tp.recvCh <- &Packet{
			Src:       src.String(),
			Raw:       buff[:n],
			Timestamp: time.Now(),
		}

	}
}

func (tp *NetTransport) Shutdown() error {
	atomic.StoreInt32(&tp.shutdown, 1)

	return tp.udp.Close()
}

func (tp *NetTransport) RecvCh() chan *Packet {
	return tp.recvCh
}

func (tp *NetTransport) Send(raw []byte) (time.Time, error) {
	tp.udp.WriteToUDP()
	return time.Now(), nil
}

package main

import (
	"context"
	"log"
	"net"
	"time"
)

type Conn struct {
	ctx    context.Context
	cancel func()
	*net.UDPConn
	remoteAddr *net.UDPAddr
	localAddr  *net.UDPAddr
}

func NewConn(ctx context.Context, localAddr, remoteAddr string) (*Conn, error) {
	laddr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		return nil, err
	}

	raddr, err := net.ResolveUDPAddr("udp", remoteAddr)
	if err != nil {
		return nil, err
	}

	udpConn, err := net.ListenUDP("udp", laddr)
	if err != nil {
		return nil, err
	}

	iCtx, cancel := context.WithCancel(ctx)
	conn := &Conn{
		ctx:        iCtx,
		cancel:     cancel,
		UDPConn:    udpConn,
		localAddr:  laddr,
		remoteAddr: raddr,
	}

	go conn.receiveLoop()
	return conn, nil
}

// Close ...
func (conn *Conn) Close() {
	conn.cancel()
	conn.UDPConn.Close()
}

func (conn *Conn) Send(data []byte) error {
	if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	_, err := conn.WriteToUDP(data, conn.remoteAddr)
	return err
}

func (conn *Conn) receiveLoop() {
	for {
		select {
		case <-conn.ctx.Done():
			return
		default:
		}

		buf := make([]byte, 4096)
		oob := make([]byte, 4096)

		if err := conn.SetReadDeadline(time.Now().Add(5 * time.Second)); err != nil {
			panic(err)
		}
		n, oobn, flags, addr, err := conn.ReadMsgUDP(buf, oob)
		if err != nil {
			log.Fatalf("receive error: %v", err)
			continue
		}

		log.Printf("data received: %v", n)
		log.Printf("oob received: %v", oobn)
		log.Printf("flags: %v", flags)
		log.Printf("remote addr: %v", addr.String())
	}
}

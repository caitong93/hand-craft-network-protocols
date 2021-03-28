package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	conn, err := NewConn(ctx, "0.0.0.0:68", "255.255.255.255:67")
	if err != nil {
		log.Fatal(err)
	}

	macAddr, err := getMacAddr("en0")
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := Discovery(conn, macAddr); err != nil {
				log.Printf("error discovery: %v", err)
			}

			select {
			case <-ctx.Done():
			case <-time.After(60 * time.Second):
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c)
	<-c
	cancel()
	conn.Close()
}

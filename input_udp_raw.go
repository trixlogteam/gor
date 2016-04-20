package main

//TODO: Implements with raw_socket_listener

import (
	"encoding/hex"
	"log"
	"net"
	"strings"
	"time"
)

// RAWUDPInput used for intercepting traffic for given address
type RAWUDPInput struct {
	data    chan []byte
	address string
	expire  time.Duration
	quit    chan bool
	conn    net.PacketConn
}

// NewRAWUDPInput constructor for RAWUDPInput. Accepts address with port as argument.
func NewRAWUDPInput(address string, expire time.Duration) (i *RAWUDPInput) {
	i = new(RAWUDPInput)
	i.data = make(chan []byte)
	i.address = address
	i.expire = expire
	i.quit = make(chan bool)

	go listen(address, i)

	return
}

func (i *RAWUDPInput) Read(data []byte) (int, error) {
	//Debug("InputRaw-Msg: ", hex.EncodeToString(msg))
	buf := <-i.data
	Debug("InputRaw-buf: ", hex.EncodeToString(buf))
	Debug("InputRaw-WithoutHeader ", hex.EncodeToString(buf[8:]))

	//MobileID,MobileIDType=0x83=131
	if (buf[8] == 131) && (len(buf) > 27) {
		//Offset 8 byte reserverd to header
		copy(data, buf[8:])
		return len(buf), nil
	}
	return 0, nil
}

func listen(address string, i *RAWUDPInput) {
	address = strings.Replace(address, "[::]", "127.0.0.1", -1)

	Debug("Listening for traffic on: " + address)

	host, port, err := net.SplitHostPort(address)
	Debug("Host: ", host)
	Debug("Port: ", port)

	if err != nil {
		log.Fatal("input-raw: error while parsing address", err)
	}

	conn, e := net.ListenPacket("ip4:udp", host)
	i.conn = conn

	if e != nil {
		log.Fatal(e)
	}

	defer i.conn.Close()

	buf := make([]byte, 64*1024) // 64kb

	for {
		// Note: ReadFrom receive messages without IP header
		n, addr, err := i.conn.ReadFrom(buf)

		if err != nil {
			if strings.HasSuffix(err.Error(), "closed network connection") {
				return
			} else {
				log.Println("Raw listener error:", err, " In address: ", addr)
				continue
			}
		}

		if n > 0 {
			newBuf := make([]byte, n)
			copy(newBuf, buf[:n])
			i.data <- newBuf
		}
	}
}

func (i *RAWUDPInput) String() string {
	return "RAW Socket input: " + i.address
}

func (i *RAWUDPInput) Close() {
	i.conn.Close()
	close(i.quit)
}

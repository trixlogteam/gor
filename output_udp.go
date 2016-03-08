package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"
)

// UDPOutput used for sending raw udp payloads
// Currently used for internal communication between listener and replay server
// Can be used for transfering binary payloads like protocol buffers
type UDPOutput struct {
	address  string
	limit    int
	buf      chan []byte
	bufStats *GorStat
}

// NewUDPOutput constructor for UDPOutput
// Initialize 10 workers which hold keep-alive connection
func NewUDPOutput(address string) io.Writer {
	o := new(UDPOutput)

	o.address = address

	o.buf = make(chan []byte, 100)
	if Settings.outputUDPStats {
		o.bufStats = NewGorStat("output_udp")

	}

	for i := 0; i < 1; i++ {
		go o.worker()

	}

	return o

}

func (o *UDPOutput) worker() {
	Debug("Workers running...")
	conn, err := o.connect(o.address)
	for ; err != nil; conn, err = o.connect(o.address) {
		time.Sleep(2 * time.Second)

	}

	defer conn.Close()

	for {
		Debug("Sending packet - UDP....")
		conn.Write(<-o.buf)
		Debug("post send - UDP")

		if err != nil {
			log.Println("Worker failed on write, exitings and starting new worker")
			go o.worker()
			break

		}

	}

}

func (o *UDPOutput) Write(data []byte) (n int, err error) {
	// We have to copy, because sending data in multiple threads
	newBuf := make([]byte, len(data))
	copy(newBuf, data)

	o.buf <- newBuf

	if Settings.outputUDPStats {
		o.bufStats.Write(len(o.buf))
	}
	Debug("OutputUDPwrite= ", newBuf)
	return len(data), nil

}

func (o *UDPOutput) connect(address string) (conn net.Conn, err error) {
	conn, err = net.Dial("udp", address)

	if err != nil {
		log.Println("Connection error ", err, o.address)

	}

	return

}

func (o *UDPOutput) String() string {
	return fmt.Sprintf("UDP output %s, limit: %d", o.address, o.limit)
}

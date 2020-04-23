/*
Package implements a binary for read serial port nmea.

*/
package gpsnmea

import (
	"bufio"
	"log"
	"net"
	"strings"
	"time"
)

type DeviceTCP struct {
	server net.Listener
	filter []string
	ok     bool
}

const (
	timeoutDeadLine = 20 * time.Second
)

func NewDeviceTCP(socket string, filters ...string) (*DeviceTCP, error) {
	log.Println("listen server ...")

	sentencesFilter := make([]string, 0)
	sentencesFilter = append(sentencesFilter, filters...)
	server, err := net.Listen("tcp", socket)

	if err != nil {
		return nil, err
	}

	dev := &DeviceTCP{
		// conn:   conn,
		server: server,
		filter: sentencesFilter,
		ok:     true,
	}
	log.Println("create server TCP!")
	return dev, nil
}

func (dev *DeviceTCP) Close() bool {
	dev.ok = false
	if err := dev.server.Close(); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (dev *DeviceTCP) Read() chan string {

	if !dev.ok {
		log.Println("Device server gps is closed")
		return nil
	}

	tcplistener := dev.server.(*net.TCPListener)
	tcplistener.SetDeadline(time.Now().Add(timeoutDeadLine))

	conn, err := dev.server.Accept()
	if err != nil {
		return nil
	}
	ch := make(chan string)

	//buf := make([]byte, 128)

	countError := 0
	go func() {
		defer close(ch)
		bf := bufio.NewReader(conn)
		for {
			conn.SetDeadline(time.Now().Add(timeoutDeadLine))
			b, err := bf.ReadBytes('\n')
			if err != nil {
				log.Println(err)
				if countError > 3 {
					conn.Close()
					return
				}
				time.Sleep(1 * time.Second)
				countError++
				continue
			}
			data := string(b[:])
			//log.Printf("serial reading: %q\n", data)
			if isSentence(data, dev.filter) {
				ch <- strings.TrimSpace(data)
			}
		}
	}()
	log.Println("reading conn")
	return ch
}

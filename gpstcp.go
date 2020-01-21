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
	conn   net.Conn
	filter []string
	ok     bool
	server net.Listener
}

func NewDeviceTCP(socket string, filters ...string) (*DeviceTCP, error) {
	log.Println("listen server ...")

	sentencesFilter := make([]string, 0)
	sentencesFilter = append(sentencesFilter, filters...)
	server, err := net.Listen("tcp", socket)
	if err != nil {
		return nil, err
	}
	// conn, err := server.Accept()
	// if err != nil {
	// 	return nil, err
	// }
	dev := &DeviceTCP{
		// conn:   conn,
		server: server,
		filter: sentencesFilter,
		ok:     false,
	}
	log.Println("Accept connection!")
	return dev, nil
}

func (dev *DeviceTCP) Close() bool {
	dev.ok = false
	if err := dev.conn.Close(); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func (dev *DeviceTCP) Accept() error {
	conn, err := dev.server.Accept()
	if err != nil {
		return err
	}
	dev.ok = true
	dev.conn = conn
	return nil
}

func (dev *DeviceTCP) Read() chan string {

	if !dev.ok {
		log.Println("Device is closed")
		return nil
	}
	ch := make(chan string)

	//buf := make([]byte, 128)

	countError := 0
	go func() {
		defer close(ch)
		bf := bufio.NewReader(dev.conn)
		for {
			b, err := bf.ReadBytes('\n')
			if err != nil {
				log.Println(err)
				if countError > 3 {
					dev.Close()
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

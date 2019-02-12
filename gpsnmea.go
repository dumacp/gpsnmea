/*
Package implements a binary for read serial port nmea.

*/
package main

import (
	"bufio"
	"flag"
	"log"
	"time"
	"errors"
	"fmt"
	"github.com/tarm/serial"
	"github.com/dumacp/pubsub"
)

var timeout int
var baudRate int
var port string
var mqtt bool

func init() {
	flag.IntVar(&timeout, "timeout", 30, "timeout to capture frames.")
	flag.IntVar(&baudRate, "baudRate", 115200, "baud rate to capture nmea's frames.")
	flag.StringVar(&port, "port", "/dev/ttyUSB1", "device serial to read.")
	flag.BoolVar(&mqtt, "mqtt", false, "send messages to local broker.")
}

//var pub *pubsub.PubSub

func main() {

	flag.Parse()


	var msgChan chan string
	if mqtt {
		pub, err := pubsub.NewConnection("go-gpsnmea")
		if err != nil {
			log.Fatal(err)
		}
		defer pub.Disconnect()
		msgChan = make(chan string)
		go pub.Publish("EVENTS/gps", msgChan)
		go func() {
			for v := range pub.Err {
				log.Println(v)
			}
		}()
	}


	log.Println("port serial config ...")
	config := &serial.Config{
		Name:        port,
		Baud:        baudRate,
		//ReadTimeout: time.Second * 3,
	}

	for {

		log.Println("open port serial")
		s, err := serial.OpenPort(config)

		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 5)
			continue
		}

		ch := make(chan []byte)
		go read(s, ch, timeout)

		for v := range ch {
			reader := &Reader{
				buf: v,
			}

			b, err := reader.Gprmc()
			if err != nil {
				log.Println(err)
				continue
			}
			log.Printf("frame: %q\n", b)
			//only publish if frame GPRMC is not quiet
			if mqtt {
				timeStamp := float64(time.Now().UnixNano())/1000000000
				frame := string(b[:])
				if frame[0:8] != "$GPRMC,," {
					msg := fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": \"GPRMC\"}",timeStamp, frame)
					msgChan <- msg
				}
			}
		}
	}
}

type Reader struct {
	buf	[]byte
}

func (r *Reader) Read(b []byte) (int, error) {
	copy(b, r.buf[:])
	return len(b), nil
}

func (r *Reader) Gprmc() ([]byte, error) {
	bf := bufio.NewReader(r)
	for {
		b, _, err := bf.ReadLine()
		if err != nil {
			return nil, err
		}
		if len(b) <= 0 {
			return nil, errors.New("EOF")
		}
		if len(b) > 6 {
			if string(b[0:6]) == "$GPRMC" {
				return b, nil
			}
		}
	}
	return nil, nil
}

func read(s *serial.Port, ch chan []byte, timeout int) {

	defer func() {
		if err := s.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	defer close(ch)

	buf := make([]byte, 128)
	tick := time.NewTicker(time.Duration(timeout) * time.Second)
	defer tick.Stop()
	eof := make(chan bool)
	defer close(eof)

	countError := 0
	for {
		select {
		case <-tick.C:
			log.Println("port serial reading 1")
			result := make([]byte, 0)
			for {
				n, _ := s.Read(buf)
				if n == 0 {
					log.Println("port serial reading 2")
					break
				}
				log.Printf("port serial reading 3: %s\n", buf[:n])
				result = append(result, buf[:n]...)
			}
			if len(result) > 0 {
				ch <- result
			}
		default:
			go func() {
				if _, err := s.Read(buf); err != nil {
					log.Println(err)
					eof <- true
					return
				}
				countError = 0
				eof <- false
			}()
			select {
			case b := <-eof:
				countError++
				if b && (countError > 10) {
					return
				}

			case <-time.After(5 * time.Second):
				log.Println("timeout!!!")
				return
			}
		}
	}
}

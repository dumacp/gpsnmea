/*
Package implements a binary for read serial port nmea.

*/
package main

import (
	_ "io"
	"bufio"
	"flag"
	"log"
	"time"
	"strings"
	"fmt"
	"github.com/tarm/serial"
	"github.com/dumacp/pubsub"
)

var timeout int
var baudRate int
var port string
var mqtt bool

var sentences = []string{"$GPRMC", "$GNGNS", "$GPVTG"}

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
		s, err := gpsnmea.NewDevice(port, baudRate)

		if err != nil {
			log.Println(err)
			time.Sleep(time.Second * 5)
			continue
		}
		ch := make(chan []byte,0)
		t1 := time.NewTicker(time.Duration(timeout) * time.Second)
		defer t1.Stop()
		go s.read(ch)

		for v := range ch {

			frame := string(v)
			if !isSentence(frame) {
				continue
			}

			log.Printf("frame: %s\n", v)
			//only publish if frame GPRMC is not quiet
			select {
			case <-t1.C:
				if mqtt {
					timeStamp := float64(time.Now().UnixNano())/1000000000
					if frame[0:8] != "$GPRMC,," {
						msg := fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": \"GPRMC\"}",timeStamp, frame)
						msgChan <- msg
					}
				}
			default:
			}
		}
	}
}

/**
type Reader struct {
	buff	[]byte
	eof	bool
}

func (r *Reader) Read(b []byte) (int, error) {
	if r.eof {
		return 0, io.EOF
	}
	for i, v := range r.buff {
		b[i] = v
	}
	r.eof = true
	return len(r.buff), nil
}
/**/

func isSentence(s1 string) (bool) {
	if len(s1) > 8 {
		for _, v := range sentences {
			if strings.HasPrefix(s1, v) {
				//if s1[1:8] != "GPRMC,," {
					return true
				//}
			}
		}
	}
	return false
}

func read(s *serial.Port, ch chan []byte) {

	defer func() {
		if err := s.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	defer close(ch)

	//buf := make([]byte, 128)

	countError := 0
	for {
		bf := bufio.NewReader(s)
		b, err := bf.ReadBytes('\n')
		if err != nil {
			log.Println(err)
			if countError > 3 {
				break
			}
			time.Sleep(1 * time.Second)
			countError++
			continue
		}
		data := b[:]
		//log.Printf("serial reading: %q\n", data)
		ch <- data
	}
}

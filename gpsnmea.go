/*
Package implements a binary for regular expressions.

The syntax of the regular expressions accepted is:

    regexp:
        concatenation { '|' concatenation }
    concatenation:
        { closure }
    closure:
        term [ '*' | '+' | '?' ]
    term:
        '^'
        '$'
        '.'
        character
        '[' [ '^' ] character-ranges ']'
        '(' regexp ')'
*/
package main

import (
	"flag"
	"github.com/tarm/serial"
	"log"
	"time"
)

var timeout int
var baudRate int
var port string

func init() {
	flag.IntVar(&timeout, "timeout", 30, "timeout to capture frames.")
	flag.IntVar(&baudRate, "baudRate", 115200, "baud rate to capture nmea's frames.")
	flag.StringVar(&port, "port", "/dev/ttyUSB1", "device serial to read.")
}

func main() {

	flag.Parse()
	log.Println("port serial config ...")
	config := &serial.Config{
		Name:        port,
		Baud:        baudRate,
		ReadTimeout: time.Second * 3,
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
			log.Printf("%q\n", v)
		}
	}
}

func read(s *serial.Port, ch chan []byte, timeout int) {

	defer func() {
		if err := s.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	defer close(ch)

	buf := make([]byte, 128)
	tick := time.Tick(time.Duration(timeout) * time.Second)
	eof := make(chan bool)
	defer close(eof)

	countError := 0
	for {
		select {
		case <-tick:
			log.Println("port serial reading")
			result := make([]byte, 0)
			for {
				n, _ := s.Read(buf)
				if n == 0 {
					break
				}
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

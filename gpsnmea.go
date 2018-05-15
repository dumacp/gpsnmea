package main

import (
	"log"
	"flag"
	"github.com/tarm/serial"
	"time"
)


var timeout int
func init() {
	flag.IntVar(&timeout, "timeout", 30, "timeout to capture frames (default: 30)")
}



func main() {
	
	flag.Parse()
	log.Println("configurando serial")
        config := &serial.Config{
		Name: "/dev/ttyUSB1", 
		Baud: 115200,
		ReadTimeout: time.Second * 3,
	}

	for {

		log.Println("abriendo serial")
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

	countError := 0
	for {
		select {
		case <-tick:
			log.Println("leyendo serial")
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
				if _ , err := s.Read(buf); err != nil {
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
					log.Println("10 lecturas")
					return
				}

			case <-time.After(5 * time.Second):
				log.Println("timeout!!!")
				return
			}
		}
	}
}


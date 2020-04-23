/*
Package implements a binary for read serial port nmea.

*/
package gpsnmea

import (
	"bufio"
	"log"
	"strings"
	"time"

	"github.com/tarm/serial"
)

type Device struct {
	config *serial.Config
	port   *serial.Port
	filter []string
	ok     bool
}

func NewDevice(portName string, baudRate int, filters ...string) (*Device, error) {
	log.Println("port serial config ...")
	config := &serial.Config{
		Name: portName,
		Baud: baudRate,
		//ReadTimeout: time.Second * 3,
	}
	sentencesFilter := make([]string, 0)
	sentencesFilter = append(sentencesFilter, filters...)
	dev := &Device{
		config: config,
		filter: sentencesFilter,
	}
	log.Println("port serial Open!")
	return dev, nil
}

func (dev *Device) Open() error {
	s, err := serial.OpenPort(dev.config)
	if err != nil {
		return err
	}
	dev.port = s
	dev.ok = true
	return nil
}

func (dev *Device) Close() bool {
	dev.ok = false
	if dev.port == nil {
		return false
	}
	if err := dev.port.Close(); err != nil {
		log.Println(err)
		return false
	}
	return true
}

func isSentence(s1 string, filter []string) bool {
	if len(s1) > 8 {
		for _, v := range filter {
			if strings.HasPrefix(s1, v) {
				//if s1[1:8] != "GPRMC,," {
				return true
				//}
			}
		}
	}
	return false
}

func (dev *Device) Read() chan string {

	if !dev.ok {
		log.Println("Device is closed")
		return nil
	}
	ch := make(chan string)

	//buf := make([]byte, 128)

	countError := 0
	go func() {
		defer close(ch)
		bf := bufio.NewReader(dev.port)
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
	log.Println("reading port")
	return ch
}

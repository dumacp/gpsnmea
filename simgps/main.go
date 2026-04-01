package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"io"
	"log"
	"os"
	"time"

	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

var timeout time.Duration
var enableMqtt bool
var ip string
var port int

func init() {
	flag.DurationVar(&timeout, "timeout", 1*time.Second, "timeout for the request")
	flag.BoolVar(&enableMqtt, "mqtt", false, "enable mqtt send to broker")
	flag.StringVar(&ip, "ip", "127.0.0.1", "ip address for mqtt broker")
	flag.IntVar(&port, "port", 1883, "port for mqtt broker")
}

func main() {
	// points := [][]float64{{6.3, -75.9}, {6.98, -75.87}, {11.016587, -74.859140}}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso de %s:  <binary> [OPTION...] [FILE]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Genera tramas GPRMC a partir de un archivo (FILE) JSON (itinenario desde plataforma) con puntos GPS\n\n")
		fmt.Fprintf(os.Stderr, "Si FILE no se especifica, se leerá desde la entrada estándar, y los datos de entradas deben ser puntos, no un arreglo de puntos como en el caso del archivo (FILE). Cuando se usa la entrada estandar (STDIN) el timeout será controlado por el ingreso de datos desde la STDIN.\n\n")
		fmt.Fprintf(os.Stderr, "Opciones:\n")

		flag.PrintDefaults()
	}

	flag.Parse()

	// fmt.Printf("len : %d\n", len(flag.Args()))

	var c mqtt.Client

	if enableMqtt {
		log.Println("MQTT enabled")
		url := fmt.Sprintf("tcp://%s:%d", ip, port)
		opt := mqtt.NewClientOptions().AddBroker(url)
		opt.SetClientID("simulator-gps")

		c = mqtt.NewClient(opt)
		t := c.Connect()
		if t.Wait() && t.Error() != nil {
			log.Fatal(t.Error())
		}
	}

	var err error
	r := os.Stdin
	stdin := true
	if len(os.Args) > 0 && len(flag.Args()) > 0 {
		r, err = os.Open(os.Args[len(os.Args)-1])
		if err != nil {
			log.Fatal(err)
		}
		stdin = false
	}
	if stdin {
		bf := bufio.NewReader(r)
		for {
			point, err := parseJSONWWithReader(bf)
			if err != nil {
				log.Fatal(err)
			}
			v := generatePointGPRMC(point)
			fmt.Println(v)
			if enableMqtt && c != nil && c.IsConnected() {
				t := float64(time.Now().UnixNano()) / 1000_000_000
				trama1 := fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": \"GPRMC\"}", t, v)
				token1 := c.Publish("EVENTS/GPS", 0, false, []byte(trama1))
				if token1.Wait() && token1.Error() != nil {
					log.Println(token1.Error())
				}
				trama2 := v
				token2 := c.Publish("GPS", 0, false, []byte(trama2))
				if token2.Wait() && token2.Error() != nil {
					log.Println(token2.Error())
				}
			}
		}
	} else {
		data, err := io.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		points, err := parseJSON(data)
		if err != nil {
			log.Fatal(err)
		}

		gprmc := chGenerateGPRMC(points)

		for v := range gprmc {
			fmt.Println(v)
			if enableMqtt && c != nil && c.IsConnected() {
				t := float64(time.Now().UnixNano()) / 1000_000_000
				trama1 := fmt.Sprintf("{\"timeStamp\": %f, \"value\": %q, \"type\": \"GPRMC\"}", t, v)
				token1 := c.Publish("EVENTS/GPS", 0, false, []byte(trama1))
				if token1.Wait() && token1.Error() != nil {
					log.Println(token1.Error())
				}
				trama2 := v
				token2 := c.Publish("GPS", 0, false, []byte(trama2))
				if token2.Wait() && token2.Error() != nil {
					log.Println(token2.Error())
				}
			}
			time.Sleep(timeout)
		}
	}

}

func openStdinOrFile() io.Reader {
	var err error
	r := os.Stdin
	if len(os.Args) > 0 && len(flag.Args()) > 0 {
		r, err = os.Open(os.Args[len(os.Args)-1])

		if err != nil {
			log.Fatal(err)
		}
	}
	return r
}

func parseJSON(data []byte) ([][]float64, error) {
	var points []*point

	err := json.Unmarshal(data, &points)
	if err != nil {
		return nil, err
	}

	var result [][]float64
	for _, p := range points {
		result = append(result, []float64{p.Lat, p.Long})
	}

	return result, nil
}

func parseJSONWWithReader(bf *bufio.Reader) ([]float64, error) {
	var p *point

	data, err := bf.ReadBytes('}')
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, err
	}

	return []float64{p.Lat, p.Long}, nil
}

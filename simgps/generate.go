package main

import (
	"fmt"
	"time"
)

// convertDegreesToGPRMC convierte grados decimales a grados y minutos como string
func convertDegreesToGPRMC(deg float64, isLat bool) string {
	degrees := int(deg)
	minutes := func() float64 {
		min := (deg - float64(degrees)) * 60
		if min < 0 {
			min = -min
		}
		return min
	}()
	if isLat {
		return fmt.Sprintf("%02d%07.4f", abs(degrees), minutes)
	}
	return fmt.Sprintf("%03d%07.4f", abs(degrees), minutes)
}

// abs devuelve el valor absoluto de un entero
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// calculateChecksum calcula el checksum de una trama NMEA
func calculateChecksum(trama string) string {
	var checksum byte
	for i := 1; i < len(trama); i++ {
		checksum ^= trama[i]
	}
	return fmt.Sprintf("%02X", checksum)
}

// generateGPRMC genera un slice de strings con las tramas GPRMC para los puntos dados
func generateGPRMC(points [][]float64) []string {
	var gprmc []string
	// Generar la fecha y hora actual
	currentTime := time.Now()

	// Datos ficticios para velocidad, rumbo y variación magnética
	velocidad := "4.0"      // nudos
	rumbo := "261.1"        // grados
	varMagnetica := "5.7,W" // variación magnética

	for _, point := range points {
		lat := point[0]
		long := point[1]

		latGPRMC := convertDegreesToGPRMC(lat, true)
		longGPRMC := convertDegreesToGPRMC(long, false)
		latDir := "N"
		if lat < 0 {
			latDir = "S"
		}
		longDir := "E"
		if long < 0 {
			longDir = "W"
		}

		// Formatear la fecha y la hora
		timeStr := currentTime.Format("020106")       // ddmmyy
		timeHHMMSS := currentTime.Format("150405.00") // hhmmss.ss

		// Construir la trama GPRMC sin el checksum
		tramaSinChecksum := fmt.Sprintf("$GPRMC,%s,A,%s,%s,%s,%s,%s,%s,%s,%s,A", timeHHMMSS, latGPRMC, latDir, longGPRMC, longDir, velocidad, rumbo, timeStr, varMagnetica)

		// Calcular el checksum
		checksum := calculateChecksum(tramaSinChecksum)

		// Añadir el checksum a la trama
		gprmcStr := fmt.Sprintf("%s*%s", tramaSinChecksum, checksum)
		gprmc = append(gprmc, gprmcStr)

		// Incrementar un segundo para la próxima trama
		currentTime = currentTime.Add(time.Second)
	}

	return gprmc
}

func chGenerateGPRMC(points [][]float64) chan string {
	// var gprmc []string
	// Generar la fecha y hora actual

	// Datos ficticios para velocidad, rumbo y variación magnética
	velocidad := "4.0"      // nudos
	rumbo := "261.1"        // grados
	varMagnetica := "5.7,W" // variación magnética

	ch := make(chan string)

	go func() {
		defer close(ch)
		for _, point := range points {
			lat := point[0]
			long := point[1]

			latGPRMC := convertDegreesToGPRMC(lat, true)
			longGPRMC := convertDegreesToGPRMC(long, false)
			latDir := "N"
			if lat < 0 {
				latDir = "S"
			}
			longDir := "E"
			if long < 0 {
				longDir = "W"
			}

			currentTime := time.Now()
			// Formatear la fecha y la hora
			timeStr := currentTime.Format("020106")       // ddmmyy
			timeHHMMSS := currentTime.Format("150405.00") // hhmmss.ss

			// Construir la trama GPRMC sin el checksum
			tramaSinChecksum := fmt.Sprintf("$GPRMC,%s,A,%s,%s,%s,%s,%s,%s,%s,%s,A", timeHHMMSS, latGPRMC, latDir, longGPRMC, longDir, velocidad, rumbo, timeStr, varMagnetica)

			// Calcular el checksum
			checksum := calculateChecksum(tramaSinChecksum)

			// Añadir el checksum a la trama
			gprmcStr := fmt.Sprintf("%s*%s", tramaSinChecksum, checksum)

			select {
			case ch <- gprmcStr:
			}

		}
	}()

	return ch
}

func generatePointGPRMC(point []float64) string {
	// var gprmc []string
	// Generar la fecha y hora actual

	// Datos ficticios para velocidad, rumbo y variación magnética
	velocidad := "4.0"      // nudos
	rumbo := "261.1"        // grados
	varMagnetica := "5.7,W" // variación magnética

	lat := point[0]
	long := point[1]

	latGPRMC := convertDegreesToGPRMC(lat, true)
	longGPRMC := convertDegreesToGPRMC(long, false)
	latDir := "N"
	if lat < 0 {
		latDir = "S"
	}
	longDir := "E"
	if long < 0 {
		longDir = "W"
	}

	currentTime := time.Now()
	// Formatear la fecha y la hora
	timeStr := currentTime.Format("020106")       // ddmmyy
	timeHHMMSS := currentTime.Format("150405.00") // hhmmss.ss

	// Construir la trama GPRMC sin el checksum
	tramaSinChecksum := fmt.Sprintf("$GPRMC,%s,A,%s,%s,%s,%s,%s,%s,%s,%s,A", timeHHMMSS, latGPRMC, latDir, longGPRMC, longDir, velocidad, rumbo, timeStr, varMagnetica)

	// Calcular el checksum
	checksum := calculateChecksum(tramaSinChecksum)

	// Añadir el checksum a la trama
	gprmcStr := fmt.Sprintf("%s*%s", tramaSinChecksum, checksum)

	return gprmcStr
}

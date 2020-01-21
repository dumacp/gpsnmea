package gpsnmea

import (
	"fmt"
	"strconv"
	"strings"

	//"log"
	"math"
)

type Gprmc struct {
	TimeStamp       string
	Validity        bool
	Lat             float64
	LatCord         string
	Long            float64
	LongCord        string
	Speed           float64
	TrueCourse      float64
	DateStamp       string
	MagneticVar     float64
	MagneticVarCord string
	Checksum        int64
}

type Gpgga struct {
	TimeStamp     string
	Lat           float64
	LatCord       string
	Long          float64
	LongCord      string
	FixQuality    int
	NumberSat     int
	HDop          float64
	Altitude      float64
	AltCord       string
	Geoidal       float64
	GeoidalUnit   string
	Dgpsupdate    float64
	DrefStationId float64
	Checksum      int64
}

func ParseRMC(s string) *Gprmc {
	fields := strings.Split(s, ",")
	//log.Printf("fields: %v,\nlen: %d", fields, len(fields))

	if len(fields) < 12 {
		return nil
	}
	timeStamp := fields[1]
	validity := false
	if fields[2] == "A" {
		validity = true
	}
	lat, _ := strconv.ParseFloat(fields[3], 64)
	long, _ := strconv.ParseFloat(fields[5], 64)
	speed, _ := strconv.ParseFloat(fields[7], 64)
	trueCourse, _ := strconv.ParseFloat(fields[8], 64)
	dateStamp := fields[9]
	magneticVar, _ := strconv.ParseFloat(fields[10], 64)

	fields2 := strings.Split(fields[12], "*")
	if len(fields2) < 2 {
		return nil
	}
	magneticVarCord := fields2[0]
	checksum, _ := strconv.ParseInt(fields2[1], 16, 32)

	return &Gprmc{
		timeStamp,
		validity,
		lat,
		fields[4],
		long,
		fields[6],
		speed,
		trueCourse,
		dateStamp,
		magneticVar,
		magneticVarCord,
		checksum,
	}
}

func ParseGGA(s string) *Gpgga {
	fields := strings.Split(s, ",")
	if len(fields) < 14 {
		return nil
	}
	timeStamp := fields[1]
	lat, _ := strconv.ParseFloat(fields[2], 64)
	long, _ := strconv.ParseFloat(fields[4], 64)
	fixQ, _ := strconv.Atoi(fields[6])
	numberSat, _ := strconv.Atoi(fields[7])
	dop, _ := strconv.ParseFloat(fields[8], 64)
	alt, _ := strconv.ParseFloat(fields[9], 64)
	geoid, _ := strconv.ParseFloat(fields[11], 64)
	dgps, _ := strconv.ParseFloat(fields[13], 64)

	fields2 := strings.Split(fields[14], "*")
	drefStation, _ := strconv.ParseFloat(fields2[0], 64)
	checksum, _ := strconv.ParseInt(fields2[1], 16, 32)

	return &Gpgga{
		timeStamp,
		lat,
		fields[3],
		long,
		fields[5],
		fixQ,
		numberSat,
		dop,
		alt,
		fields[10],
		geoid,
		fields[12],
		dgps,
		drefStation,
		checksum,
	}
}

func LatLongToDecimalDegree(num float64, cord string) float64 {
	dec, fra := math.Modf(num / 100)
	if strings.Contains(cord, "W") || strings.Contains(cord, "S") {
		return (-1) * (dec + fra*100/60)
	}
	return dec + fra*100/60

}

func DecimalDegreeToLat(lat float64) string {
	latDirection := "N"
	if lat <= 0 {
		latDirection = "S"
		lat = -lat
	}
	latitude := uint8(lat)
	latitudeMinutes := uint8((lat - float64(latitude)) * 60)
	latitudeSeconds := (lat - float64(latitude) - float64(latitudeMinutes)/60) * 3600

	return fmt.Sprintf("%02d%02d.%v,%v", latitude, latitudeMinutes, int(latitudeSeconds*100000/60), latDirection)
}

func DecimalDegreeToLon(lon float64) string {
	lonDirection := "E"
	if lon <= 0 {
		lonDirection = "W"
		lon = -lon
	}
	longitude := uint8(lon)
	longitudeMinutes := uint8((lon - float64(longitude)) * 60)
	longitudeSeconds := (lon - float64(longitude) - float64(longitudeMinutes)/60) * 3600

	return fmt.Sprintf("%03d%02d.%v,%v", longitude, longitudeMinutes, int(longitudeSeconds*1000000/60), lonDirection)
}

//:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
//:::                                                                         :::
//:::  This routine calculates the distance between two points (given the     :::
//:::  latitude/longitude of those points). It is being used to calculate     :::
//:::  the distance between two locations using GeoDataSource (TM) prodducts  :::
//:::                                                                         :::
//:::  Definitions:                                                           :::
//:::    South latitudes are negative, east longitudes are positive           :::
//:::                                                                         :::
//:::  Passed to function:                                                    :::
//:::    lat1, lon1 = Latitude and Longitude of point 1 (in decimal degrees)  :::
//:::    lat2, lon2 = Latitude and Longitude of point 2 (in decimal degrees)  :::
//:::    unit = the unit you desire for results                               :::
//:::           where: 'M' is statute miles (default)                         :::
//:::                  'K' is kilometers                                      :::
//:::                  'N' is nautical miles                                  :::
//:::                                                                         :::
//:::  Worldwide cities and other features databases with latitude longitude  :::
//:::  are available at https://www.geodatasource.com                         :::
//:::                                                                         :::
//:::  For enquiries, please contact sales@geodatasource.com                  :::
//:::                                                                         :::
//:::  Official Web site: https://www.geodatasource.com                       :::
//:::                                                                         :::
//:::               GeoDataSource.com (C) All Rights Reserved 2018            :::
//:::                                                                         :::
//:::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::
func Distance(lat1 float64, lng1 float64, lat2 float64, lng2 float64, unit ...string) float64 {
	const PI float64 = 3.141592653589793

	radlat1 := float64(PI * lat1 / 180)
	radlat2 := float64(PI * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(PI * theta / 180)

	dist := math.Sin(radlat1)*math.Sin(radlat2) + math.Cos(radlat1)*math.Cos(radlat2)*math.Cos(radtheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515

	if len(unit) > 0 {
		if unit[0] == "K" {
			dist = dist * 1.609344
		} else if unit[0] == "N" {
			dist = dist * 0.8684
		}
	}

	return dist
}

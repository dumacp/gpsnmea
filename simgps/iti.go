package main

type point struct {
	CheckPointId string  `json:"checkPointId"`
	Type         string  `json:"type"`
	Name         string  `json:"name"`
	Radios       int     `json:"radios"`
	MaxSpeed     string  `json:"maxSpeed"`
	Long         float64 `json:"long"`
	Lat          float64 `json:"lat"`
}

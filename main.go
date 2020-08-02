package main

import (
	"fmt"
	"net/http"
	"time"
)

const (
	PortName        = "/dev/ttyS0"
	MINIMAMREADSIZE = 4
	NCCS            = 19
	kBOTHER         = 0x1000
)

var (
	jst *time.Location
	res string
)

func init() {
	var err error
	jst, err = time.LoadLocation("Asia/Tokyo")
	if err != nil {
		panic(err)
	}
}

type Res struct {
	Co2       int64     `json:"co_2"`
	Pressure  float64   `json:"pressure"`
	Humidity  float64   `json:"humidity"`
	Temputure float64   `json:"temputure"`
	Date      time.Time `json:"date"`
}

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, res)
}

func main() {
	w, err := NewWorker()
	if err != nil {
		panic(err)
	}
	go w.Run()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}

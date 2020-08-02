package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/nozo-moto/mhz19-bme280-go/bme280"
	"github.com/nozo-moto/mhz19-bme280-go/mh_z19"
	"github.com/nozo-moto/mhz19-bme280-go/serial"
)

type Worker struct {
	Co2 mh_z19.MhZ19
	THP bme280.Bme280
}

func NewWorker() (*Worker, error) {
	var w Worker
	err := w.init()
	if err != nil {
		return &w, fmt.Errorf("failed to init")
	}
	return &w, nil
}

func (w *Worker) close() error {
	if err := w.Co2.Close(); err != nil {
		return fmt.Errorf("failed to close: %w", err)
	}
	if err := w.THP.Close(); err != nil {
		return fmt.Errorf("failed to close: %w", err)
	}

	return nil
}

func (w *Worker) init() error {
	uart, err := serial.NewUart("/dev/serial0")
	if err != nil {
		return fmt.Errorf("i2c error %w", err)
	}
	co2 := mh_z19.New(uart)

	device, err := serial.NewI2C("/dev/i2c-1")
	if err != nil {
		return fmt.Errorf("i2c error %w", err)
	}
	thp, err := bme280.New(device, bme280.Forced)
	if err != nil {
		return fmt.Errorf("sensor initalize error %w", err)
	}
	w.Co2 = co2
	w.THP = thp

	return nil
}

func (w *Worker) getSensorData() (err error) {
	if err = w.init(); err != nil {
		return
	}
	var (
		co2              int64
		temp, hum, press uint32
	)
	co2, err = w.Co2.Read()
	if err != nil {
		if !errors.Is(err, mh_z19.ErrChecksum) {
			return
		}
	}
	if co2 == 0 {
		return errors.New("Co2 == 0")
	}

	temp, press, hum, err = w.THP.Read()
	if err != nil {
		return
	}

	r := &Res{
		Co2:       co2,
		Temputure: float64(temp) / 100,
		Humidity:  float64(hum>>12) / 1024,
		Pressure:  float64(press / 25600),
		Date:      time.Now().In(jst),
	}
	b, err := json.Marshal(r)
	if err != nil {
		return
	}
	res = string(b)
	w.close()
	return
}

func (w *Worker) Run() {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := w.getSensorData(); err != nil {
				log.Printf("Error %+v", err)
			} else {
				log.Println("Co2 is ", res)
			}
		}
	}
}

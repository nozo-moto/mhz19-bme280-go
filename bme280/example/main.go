package main

import (
	"fmt"

	"github.com/nozo-moto/mhz19-bme280-go/bme280"
	"github.com/nozo-moto/mhz19-bme280-go/serial"
)

func main() {
	portName := "/dev/i2c-1"
	device, err := serial.NewI2C(portName)
	if err != nil {
		panic(fmt.Sprintf("i2c error %v", err))
	}
	sensor, err := bme280.New(device, bme280.Forced)
	if err != nil {
		panic(fmt.Sprintf("sensor initalize error %v", err))
	}
	temp, press, hum, err := sensor.Read()
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	fmt.Printf("temp %d, press %d, hum %f\n ", temp, press/25600, float64(hum>>12)/1024.0)
}

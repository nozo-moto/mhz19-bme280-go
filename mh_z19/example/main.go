package main

import (
	"errors"
	"fmt"

	"github.com/nozo-moto/mhz19-bme280-go/mh_z19"
	"github.com/nozo-moto/mhz19-bme280-go/serial"
)

func main() {
	portName := "/dev/serial0"
	uart, err := serial.NewUart(portName)
	if err != nil {
		panic(fmt.Sprintf("i2c error %v", err))
	}

	sensor := mh_z19.New(uart)
	co2, err := sensor.Read()
	if err != nil {
		if errors.Is(err, mh_z19.ErrChecksum) {
			fmt.Println("co2 ", co2)
		}
		panic(err)
	}

	fmt.Println("co2 ", co2)
}

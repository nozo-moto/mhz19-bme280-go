package mh_z19

import (
	"errors"

	"github.com/nozo-moto/mhz19-bme280-go/serial"
)

var (
	ErrChecksum = errors.New("checksum is not good")
)

type MhZ19 interface {
	Read() (int64, error)
	Close() error
}

type mhZ19impl struct {
	uart serial.Conn
}

func New(c serial.Conn) MhZ19 {
	return &mhZ19impl{
		uart: c,
	}
}

func (m *mhZ19impl) Read() (int64, error) {
	var (
		checksum int
		co2      int64
	)
	if err := m.uart.Write([]byte{0xff, 0x01, 0x86, 0x00, 0x00, 0x00, 0x00, 0x00, 0x79}); err != nil {
		return 0, err
	}

	buf := make([]byte, 1)
	for i := 0; i < 9; i++ {
		err := m.uart.Read(buf)
		if err != nil {
			return 0, err
		}
		if i != 0 && i != 8 {
			checksum += int(buf[0])
		}
		switch i {
		case 2: // high level concentration
			co2 += int64(buf[0]) * 256
		case 3: // low level concentration
			co2 += int64(buf[0])
		case 8: //calcurate checksum
			checksumBuf := int(buf[0])
			if 256-checksum != checksumBuf {
				return co2, ErrChecksum
			}
		}
	}

	return co2, nil
}

func (m *mhZ19impl) Close() error {
	return m.uart.Close()
}

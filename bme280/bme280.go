package bme280

import (
	"encoding/binary"
	"fmt"

	"github.com/nozo-moto/mhz19-bme280-go/serial"
)

// reference by bme280 data sheet's implement
// https://ae-bst.resource.bosch.com/media/_tech/media/datasheets/BST-BME280-DS002.pdf

type Bme280 interface {
	Read() (temp, press, hum uint32, err error)
	Close() error
}

// Mode is bme280's 3 modes
type Mode byte

const (
	ctrlMeasAddr = 0xF4
	pressMsbAddr = 0xF7 // read from this address
	configAddr   = 0xF5
	ctrlHumAddr  = 0xF4

	Sleep  Mode = 0x00
	Forced Mode = 0x01
	Normal Mode = 0x03

	// TODO rename to calib00Addr
	calib00 = 0x88
	calib26 = 0xE1
)

// New return bme280 impl
func New(s serial.Conn, mode Mode) (Bme280, error) {
	b := &bme280impl{
		i2c:  s,
		mode: mode,
	}
	if err := b.setup(); err != nil {
		return nil, fmt.Errorf("faile to setup: %w", err)
	}
	//if err = b.i2c.Write([]byte{byte(b.mode)}); err != nil {
	//	err = fmt.Errorf("faile to write mode: %w", err)
	//	return
	//}

	if err := b.readCalib(); err != nil {
		return nil, fmt.Errorf("read calib error: %w", err)
	}
	return b, nil
}

type bme280impl struct {
	i2c serial.Conn

	// mode has 3 types sleep, force, normal
	mode Mode

	// tFine using fine temperature for calibration of sensor
	tFine uint32
	digT1 uint16
	digT2 int16
	digT3 int16

	digP1 uint16
	digP2 int16
	digP3 int16
	digP4 int16
	digP5 int16
	digP6 int16
	digP7 int16
	digP8 int16
	digP9 int16

	digH1 int8
	digH2 int16
	digH3 int8
	digH4 int16
	digH5 int16
	digH6 int8
}

func (b *bme280impl) Close() error {
	return b.i2c.Close()
}

func (b *bme280impl) readCalib() error {
	// get 0x88 -> 0xA1
	buf := make([]byte, 26)
	err := b.i2c.ReadReg(calib00, buf)
	if err != nil {
		return fmt.Errorf("read calib00 error: %w", err)
	}
	b.digT1 = binary.LittleEndian.Uint16(buf)
	b.digT2 = int16(binary.LittleEndian.Uint16(buf[2:]))
	b.digT3 = int16(binary.LittleEndian.Uint16(buf[4:]))
	b.digP1 = binary.LittleEndian.Uint16(buf[6:])
	b.digP2 = int16(binary.LittleEndian.Uint16(buf[8:]))
	b.digP3 = int16(binary.LittleEndian.Uint16(buf[10:]))
	b.digP4 = int16(binary.LittleEndian.Uint16(buf[12:]))
	b.digP5 = int16(binary.LittleEndian.Uint16(buf[14:]))
	b.digP6 = int16(binary.LittleEndian.Uint16(buf[16:]))
	b.digP7 = int16(binary.LittleEndian.Uint16(buf[18:]))
	b.digP8 = int16(binary.LittleEndian.Uint16(buf[20:]))
	b.digP9 = int16(binary.LittleEndian.Uint16(buf[22:]))
	b.digH1 = int8(buf[25])

	// get 0xE1 -> 0xE7
	buf = make([]byte, 7)
	err = b.i2c.ReadReg(calib26, buf)
	if err != nil {
		return fmt.Errorf("read calib26 error: %w", err)
	}

	b.digH2 = int16(buf[1])<<8 | int16(buf[0])
	b.digH3 = int8(buf[2])

	b.digH4 = int16(buf[3])<<4 + int16(buf[4]&0x0f)
	b.digH5 = int16(buf[4]&0xF0)<<4 | int16(buf[5])
	b.digH6 = int8(buf[6])

	return nil
}

func (b *bme280impl) setup() error {
	if err := b.i2c.ReadReg(configAddr, []byte{0x80}); err != nil {
		return fmt.Errorf("failed to read config Addr: %w", err)
	}
	if err := b.i2c.ReadReg(ctrlHumAddr, []byte{0x05}); err != nil {
		return fmt.Errorf("failed to read ctrlHum Addr: %w", err)
	}
	if err := b.i2c.ReadReg(ctrlMeasAddr, []byte{0xA9}); err != nil {
		return fmt.Errorf("failed to read ctrlMeasAddr: %w", err)
	}

	return nil
}

// Read read sensor data then return data
func (b *bme280impl) Read() (temp, press, hum uint32, err error) {
	var t, p, h int32
	t, p, h, err = b.readSensorData()
	temp = b.compensateT(t)
	press = b.compensateP(p)
	hum = b.compensateH(h)

	return
}

func (b *bme280impl) readSensorData() (temp, press, hum int32, err error) {
	buf := make([]byte, 8)
	if err = b.i2c.ReadReg(pressMsbAddr, buf); err != nil {
		err = fmt.Errorf("failed to read data from sensor: %w", err)
		return
	}
	press = int32(buf[0])<<12 | int32(buf[1])<<4 | int32(buf[2])>>4
	temp = int32(buf[3])<<12 | int32(buf[4])<<4 | int32(buf[5])>>4
	hum = int32(buf[6])<<8 | int32(buf[7])
	return
}

func (b *bme280impl) compensateT(adcT int32) uint32 {
	var (
		var1, var2 uint32
	)

	var1 = (uint32(adcT>>3) - uint32(b.digT1)<<1) * uint32(b.digT2) >> 11
	var2 = (uint32(adcT>>4) - uint32(b.digT1)) * (uint32(adcT>>4) - uint32(b.digT1)) >> 12 * uint32(b.digT3) >> 14
	b.tFine = var1 + var2
	return (b.tFine*5 + 128) >> 8
}

func (b *bme280impl) compensateP(adcP int32) uint32 {
	var (
		var1, var2, p int64
	)
	var1 = int64(b.tFine) - 128000
	var2 = var1 * var1 * int64(b.digP6)
	var2 = var2 + (var1 * int64(b.digP5) << 17)
	var2 = var2 + (int64(b.digP4) << 35)
	var1 = (var1*var1*int64(b.digP3))>>8 + (var1*int64(b.digP2))<<12
	var1 = ((int64(1)<<47 + var1) * int64(b.digP1)) >> 33
	if var1 == 0 {
		return 0
	}
	p = int64(1048576 - adcP)
	p = ((p<<31 - var2) * 3125) / var1
	var1 = (int64(b.digP9) * (p >> 13) * (p >> 13)) >> 25
	var2 = (int64(b.digP8) * p) >> 19
	p = (p+var1+var2)>>8 + int64(b.digP7)<<4
	return uint32(p)
}

func (b *bme280impl) compensateH(adcH int32) uint32 {
	var (
		vX1U32r int32
	)
	vX1U32r = int32(b.tFine - 76800)

	vX1U32r = (((adcH<<14 - int32(b.digH4)<<20 - int32(b.digH5)*vX1U32r) + int32(16384)) >> 15) *
		(((((((vX1U32r*int32(b.digH6))>>10)*(((vX1U32r*int32(b.digH3))>>11)+int32(32768)))>>10)+
			int32(2097152))*int32(b.digH2) + 8192) >> 14)

	vX1U32r = vX1U32r - ((((vX1U32r>>15)*(vX1U32r>>15))>>7)*int32(b.digH1))>>4
	if vX1U32r < 0 {
		return 0
	} else if vX1U32r > 419430400 {
		return 419430400
	}
	return uint32(vX1U32r)
}

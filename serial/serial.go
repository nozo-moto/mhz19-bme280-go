package serial

import (
	"errors"
	"fmt"
	"os"
	"syscall"
	"unsafe"
)

type serial struct {
	file *os.File
}

type Conn interface {
	Write(buf []byte) (err error)
	WriteReg(reg byte, buf []byte) (err error)
	Read(buf []byte) (err error)
	ReadReg(reg byte, buf []byte) (err error)
	Close() (err error)
}

func NewUart(portName string) (Conn, error) {
	file, err := os.OpenFile(
		portName,
		syscall.O_RDWR|syscall.O_NOCTTY|syscall.O_NONBLOCK,
		0600,
	)
	if err != nil {
		return nil, fmt.Errorf("faile to open file %w", err)
	}
	err = syscall.SetNonblock(int(file.Fd()), false)
	if err != nil {
		return nil, fmt.Errorf("faile to set nonblock %w", err)
	}
	termios := newTermios()
	file, err = ioctl(file, 0x402C542B, uintptr(unsafe.Pointer(termios)))
	if err != nil {
		return nil, fmt.Errorf("faile to ioctl %w", err)
	}

	return &serial{
		file: file,
	}, err
}

func NewI2C(portName string) (Conn, error) {
	file, err := os.OpenFile(
		portName,
		os.O_RDWR,
		os.ModeDevice,
	)
	if err != nil {
		return nil, fmt.Errorf("faile to open file %w", err)
	}

	file, err = ioctl(file, 0x0703, 0x76)
	if err != nil {
		err = fmt.Errorf("faile to ioctl %w", err)
	}
	return &serial{file: file}, err
}

func (d *serial) Write(buf []byte) (err error) {
	num, err := d.file.Write(buf)
	if err != nil {
		return fmt.Errorf("failed to Write data: %w", err)
	}
	if len(buf) != num {
		return errors.New("write length and response is diffrent")
	}
	return
}

func (d *serial) WriteReg(reg byte, buf []byte) (err error) {
	return d.Write(append([]byte{reg}, buf...))
}

func (d *serial) Read(buf []byte) (err error) {
	_, err = d.file.Read(buf)
	if err != nil {
		err = fmt.Errorf("failed to read data: %w", err)
	}
	return
}

func (d *serial) ReadReg(reg byte, buf []byte) (err error) {
	err = d.Write([]byte{reg})
	if err != nil {
		return fmt.Errorf("faile to write reg: %w", err)
	}
	err = d.Read(buf)
	if err != nil {
		return fmt.Errorf("faile to read reg: %w", err)
	}
	return
}

func (d *serial) Close() (err error) {
	return d.file.Close()
}

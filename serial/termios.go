package serial

import (
	"errors"
	"fmt"
	"os"
	"syscall"
)

const (
	PortName        = "/dev/serial0"
	MINIMAMREADSIZE = 4
	NCCS            = 32
	kBOTHER         = 0x1000
)

type tcflag_t uint
type speed_t uint
type cc_t byte

type termios struct {
	c_iflag  tcflag_t   // 入力
	c_oflag  tcflag_t   // 出力
	c_cflag  tcflag_t   // 制御
	c_lflag  tcflag_t   // local
	c_cc     [NCCS]cc_t // 特殊制御文字
	c_ispeed speed_t    // 入力スピード
	c_ospeed speed_t    // 出力スピード
}

func newTermios() *termios {
	c := [NCCS]cc_t{}
	c[syscall.VTIME] = cc_t(0)
	c[syscall.VMIN] = cc_t(MINIMAMREADSIZE)
	return &termios{
		c_cflag:  syscall.CLOCAL | syscall.CREAD | kBOTHER | syscall.CS8,
		c_cc:     c,
		c_ispeed: speed_t(9600),
		c_ospeed: speed_t(9600),
	}
}

func ioctl(file *os.File, a1, a2 uintptr) (*os.File, error) {
	r, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		uintptr(file.Fd()),
		uintptr(a1),
		uintptr(a2),
	)
	if errno != 0 {
		return file, fmt.Errorf("faile to syscall.Syscall: %w", errno)
	}
	if r != 0 {
		return file, errors.New("unknown error from SYS_IOCTL")
	}

	return file, nil
}

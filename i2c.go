// Package i2c provides low level control over the linux i2c bus.
//
// Before usage you should load the i2c-dev kenel module
//
//      sudo modprobe i2c-dev
//
// Each i2c bus can address 127 independent i2c devices, and most
// linux systems contain several buses.
package i2c

import (
	"fmt"
	"os"
	"syscall"
)

const (
	i2c_SLAVE = 0x0703
)

// I2C represents a connection to an i2c device.
type I2C struct {
	rc *os.File
}

// New opens a connection to an i2c device.
func New(addr uint8, bus int) (*I2C, error) {
	f, err := os.OpenFile(fmt.Sprintf("/dev/i2c-%d", bus), os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	if err := ioctl(f.Fd(), i2c_SLAVE, uintptr(addr)); err != nil {
		return nil, err
	}
	return &I2C{f}, nil
}

// Write sends buf to the remote i2c device. The interpretation of
// the message is implementation dependant.
func (i2c *I2C) Write(buf []byte) (int, error) {
	n, err := i2c.rc.Write(buf)
	return n, err
}

func (i2c *I2C) WriteByte(b byte) (int, error) {
	var buf []byte
	buf[0] = b
	n, err := i2c.rc.Write(buf)
	return n, err
}

func (i2c *I2C) Read(p []byte) (int, error) {
	n, err := i2c.rc.Write(p)
	return n, err
}

func (i2c *I2C) Close() error {
	err := i2c.Close()
	return err
}

func ioctl(fd, cmd, arg uintptr) (err error) {
	_, _, e1 := syscall.Syscall6(syscall.SYS_IOCTL, fd, cmd, arg, 0, 0, 0)
	if e1 != 0 {
		err = e1
	}
	return
}

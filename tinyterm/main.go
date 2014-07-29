package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"sync"

	"github.com/davecheney/i2c"
)

var (
	flags  = flag.NewFlagSet("", flag.ContinueOnError)
	stderr = flags.Bool("e", true, "redirect stderr")
)

func init() {
	flags.Parse(os.Args[1:])
	if len(flags.Args()) < 1 {
		log.Fatal("command missing")
	}
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type lcdWriter struct {
	sync.Mutex
	pos byte       // screen position, 4 x 16 chars
	buf [0x60]byte // hd44780 ram, arranged by lunacy
	lcd *i2c.Lcd
}

// maps from linear memory locations to crackful hd44780 memory locations
var lcdtab = [...]byte{
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f,
	0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f,
}

func (w *lcdWriter) Write(buf []byte) (int, error) {
	w.Lock()
	defer w.Unlock()
	var n int
	for ; n < len(buf); n++ {
		switch c := buf[n]; c {
		case '\r':
			continue
		case '\t':
			w.pos += 4
			w.pos -= (w.pos % 0x4)
		case '\n':
			w.pos += 0x10
			w.pos -= (w.pos % 0x10)
		default:
			index := lcdtab[w.pos]
			w.buf[index] = c
			w.lcd.Command(i2c.CMD_DDRAM_Set + index)
			w.lcd.Write(w.buf[index : index+1])
			w.pos++
		}
		w.scrollup()
	}
	return n, nil
}

func (w *lcdWriter) scrollup() {
	redraw := false
	for ; w.pos >= 0x40; w.pos -= 0x10 {
		copy(w.buf[0x00:0x10], w.buf[0x40:0x50])
		copy(w.buf[0x40:0x50], w.buf[0x10:0x20])
		copy(w.buf[0x10:0x20], w.buf[0x50:0x60])
		fill(w.buf[0x50:0x60], 0x20)
		redraw = true
	}
	if redraw {
		w.redraw()
	}
}

func (w *lcdWriter) redraw() {
	w.lcd.Command(i2c.CMD_DDRAM_Set + 0x00) // home
	w.lcd.Write(w.buf[0x00:0x20])
	w.lcd.Command(i2c.CMD_DDRAM_Set + 0x40)
	w.lcd.Write(w.buf[0x40:0x60])
}

func fill(b []byte, c byte) {
	for i := range b {
		b[i] = c
	}
}

func main() {
	dev, err := i2c.New(0x27, 1) // vga port
	check(err)

	lcd, err := i2c.NewLcd(dev, 2, 1, 0, 4, 5, 6, 7, 3)
	check(err)
	lcd.BacklightOn()
	lcd.Clear()

	w := lcdWriter{lcd: lcd}
	fill(w.buf[:], 0x20)
	cmd := exec.Command(flags.Args()[0], flags.Args()[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = &w
	if *stderr {
		cmd.Stderr = &w
	}

	cmd.Run()
}

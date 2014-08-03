package i2c

// i2c LCD adapter
// Adapted from http://think-bowl.com/raspberry-pi/installing-the-think-bowl-i2c-libraries-for-python/

import "time"

const (
	// Commands
	CMD_Clear_Display        = 0x01
	CMD_Return_Home          = 0x02
	CMD_Entry_Mode           = 0x04
	CMD_Display_Control      = 0x08
	CMD_Cursor_Display_Shift = 0x10
	CMD_Function_Set         = 0x20
	CMD_DDRAM_Set            = 0x80

	// Options
	OPT_Increment = 0x02 // CMD_Entry_Mode
	// OPT_Display_Shift  = 0x01 // CMD_Entry_Mode
	OPT_Enable_Display = 0x04 // CMD_Display_Control
	OPT_Enable_Cursor  = 0x02 // CMD_Display_Control
	OPT_Enable_Blink   = 0x01 // CMD_Display_Control
	OPT_Display_Shift  = 0x08 // CMD_Cursor_Display_Shift
	OPT_Shift_Right    = 0x04 // CMD_Cursor_Display_Shift 0 = Left
	OPT_2_Lines        = 0x08 // CMD_Function_Set 0 = 1 line
	OPT_5x10_Dots      = 0x04 // CMD_Function_Set 0 = 5x7 dots
)

type Lcd struct {
	i2c                                   *I2C
	en, rw, rs, d4, d5, d6, d7, backlight byte
	backlight_state                       bool
}

func NewLcd(i2c *I2C, en, rw, rs, d4, d5, d6, d7, backlight byte) (*Lcd, error) {
	lcd := Lcd{
		i2c:       i2c,
		en:        en,
		rw:        rw,
		rs:        rs,
		d4:        d4,
		d5:        d5,
		d6:        d6,
		d7:        d7,
		backlight: backlight,
	}
	// Activate LCD
	var data byte
	data = pinInterpret(lcd.d4, data, true)
	data = pinInterpret(lcd.d5, data, true)
	lcd.enable(data)
	time.Sleep(200 * time.Millisecond)
	lcd.enable(data)
	time.Sleep(100 * time.Millisecond)
	lcd.enable(data)
	time.Sleep(100 * time.Millisecond)

	// Initialize 4-bit mode
	data = pinInterpret(lcd.d4, data, false)
	lcd.enable(data)
	time.Sleep(10 * time.Millisecond)

	lcd.command(CMD_Function_Set | OPT_2_Lines)
	lcd.command(CMD_Display_Control | OPT_Enable_Display | OPT_Enable_Cursor)
	lcd.command(CMD_Clear_Display)
	lcd.command(CMD_Entry_Mode | OPT_Increment | OPT_Display_Shift)

	return &lcd, nil
}

func (lcd *Lcd) enable(data byte) {
	// Determine if black light is on and insure it does not turn off or on
	if lcd.backlight_state {
		data = pinInterpret(lcd.backlight, data, true)
	} else {
		data = pinInterpret(lcd.backlight, data, false)
	}
	lcd.i2c.Write(data)
	lcd.i2c.Write(pinInterpret(lcd.en, data, true))
	lcd.i2c.Write(data)
}

func (lcd *Lcd) command(data byte) {
	lcd.write(data, true)
}

func (lcd *Lcd) write(data byte, command bool) {
	var i2c_data byte

	// Add data for high nibble
	hi_nibble := data >> 4
	i2c_data = pinInterpret(lcd.d4, i2c_data, (hi_nibble&0x01 == 0x01))
	i2c_data = pinInterpret(lcd.d5, i2c_data, ((hi_nibble>>1)&0x01 == 0x01))
	i2c_data = pinInterpret(lcd.d6, i2c_data, ((hi_nibble>>2)&0x01 == 0x01))
	i2c_data = pinInterpret(lcd.d7, i2c_data, ((hi_nibble>>3)&0x01 == 0x01))

	// # Set the register selector to 1 if this is data
	if !command {
		i2c_data = pinInterpret(lcd.rs, i2c_data, true)
	}

	//  Toggle Enable
	lcd.enable(i2c_data)

	i2c_data = 0x00

	// Add data for high nibble
	low_nibble := data & 0x0F
	i2c_data = pinInterpret(lcd.d4, i2c_data, (low_nibble&0x01 == 0x01))
	i2c_data = pinInterpret(lcd.d5, i2c_data, ((low_nibble>>1)&0x01 == 0x01))
	i2c_data = pinInterpret(lcd.d6, i2c_data, ((low_nibble>>2)&0x01 == 0x01))
	i2c_data = pinInterpret(lcd.d7, i2c_data, ((low_nibble>>3)&0x01 == 0x01))

	// Set the register selector to 1 if this is data
	if !command {
		i2c_data = pinInterpret(lcd.rs, i2c_data, true)
	}

	lcd.enable(i2c_data)
}

func (lcd *Lcd) BacklightOn() {
	lcd.i2c.Write(pinInterpret(lcd.backlight, 0x00, true))
	lcd.backlight_state = true
}

func (lcd *Lcd) BacklightOff() {
	lcd.i2c.Write(pinInterpret(lcd.backlight, 0x00, false))
	lcd.backlight_state = false
}

func (lcd *Lcd) Clear() {
	lcd.command(CMD_Clear_Display)
}

func (lcd *Lcd) Home() {
	lcd.command(CMD_Return_Home)
}

func (lcd *Lcd) SetPosition(line, pos byte) {
	var address byte
	switch line {
	case 1:
		address = pos
	case 2:
		address = 0x40 + pos
	case 3:
		address = 0x10 + pos
	case 4:
		address = 0x50 + pos
	}
	lcd.command(CMD_DDRAM_Set + address)
}

func (lcd *Lcd) Command(cmd byte) {
	lcd.command(cmd)
}

func (lcd *Lcd) writeChar(char byte) {
	lcd.write(char, false)
}

func (lcd *Lcd) Write(buf []byte) (int, error) {
	for _, c := range buf {
		lcd.writeChar(c)
	}
	return len(buf), nil
}

func pinInterpret(pin, data byte, value bool) byte {
	if value {
		// Construct mask using pin
		var mask byte = 0x01 << (pin)
		data = data | mask
	} else {
		// Construct mask using pin
		var mask byte = 0x01<<(pin) ^ 0xFF
		data = data & mask
	}
	return data
}

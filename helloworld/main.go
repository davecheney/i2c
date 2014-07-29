package main

import "github.com/davecheney/i2c"
import "log"
import "fmt"
import "time"

func check(err error) {
	if err != nil { log.Fatal(err) }
}

func main() {
	i, err := i2c.New(0x27, 1)
	check(err)
	lcd, err := i2c.NewLcd(i, 2, 1, 0, 4, 5, 6, 7, 3)
	check(err)
	lcd.BacklightOn()
	lcd.Clear()
	for {
		lcd.Home()
		t := time.Now()
		lcd.SetPosition(1, 0)
		fmt.Fprint(lcd, t.Format("Monday Jan 2"))
		lcd.SetPosition(2, 0)
		fmt.Fprint(lcd, t.Format("15:04:05 2006"))
		lcd.SetPosition(4, 0)
		fmt.Fprint(lcd, "i2c, VGA, and Go")
		time.Sleep(333 * time.Millisecond)
	}
}	

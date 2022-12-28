package keys

import (
	"fmt"
	"log"
)

func Open() *TTY {
	tty, err := open("/dev/tty")
	if err != nil {
		log.Fatal(err)
	}
	return tty
}

func (tty *TTY) Close() error {
	return tty.close()
}

func (tty *TTY) ReadKey() string {
	r, err := tty.readRune()
	if err != nil {
		return ""
	}
	char := fmt.Sprintf("%c", r)
	if char == "d" || char == "D" {
		return "d"
	}
	if char == "s" || char == "S" {
		return "s"
	}
	return ""
}

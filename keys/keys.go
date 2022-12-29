package keys

import (
	"fmt"
	"log"
)

var key string

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
	key = key + char
	if tty.buffered() {
		return ""
	}
	return checkReadKey()
}

func checkReadKey() string {
	readKey := key
	key = ""
	switch readKey {
	case "a", "A":
		return "a"
	case "s", "S":
		return "s"
	case "d", "D":
		return "d"
	}
	return ""
}

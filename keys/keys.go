package keys

import "fmt"

func Open() (*TTY, error) {
	return open("/dev/tty")
}

func (tty *TTY) Close() error {
	return tty.close()
}

func (tty *TTY) ReadKey() string {
	r, err := tty.readRune()
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%c", r)
}

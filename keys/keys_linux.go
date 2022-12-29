//go:build !windows
// +build !windows

package keys

import (
	"bufio"
	"os"
	"os/signal"

	"golang.org/x/sys/unix"
)

type TTY struct {
	in      *os.File
	bin     *bufio.Reader
	out     *os.File
	termios unix.Termios
	ss      chan os.Signal
}

const (
	ioctlReadTermios  = unix.TCGETS
	ioctlWriteTermios = unix.TCSETS
)

func open(path string) (*TTY, error) {
	tty := new(TTY)

	in, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	tty.in = in
	tty.bin = bufio.NewReader(in)

	out, err := os.OpenFile(path, unix.O_WRONLY, 0)
	if err != nil {
		return nil, err
	}
	tty.out = out

	termios, err := unix.IoctlGetTermios(int(tty.in.Fd()), ioctlReadTermios)
	if err != nil {
		return nil, err
	}
	tty.termios = *termios

	termios.Iflag &^= unix.ISTRIP | unix.INLCR | unix.ICRNL | unix.IGNCR | unix.IXOFF
	termios.Lflag &^= unix.ECHO | unix.ICANON /*| unix.ISIG*/
	termios.Cc[unix.VMIN] = 1
	termios.Cc[unix.VTIME] = 0
	if err := unix.IoctlSetTermios(int(tty.in.Fd()), ioctlWriteTermios, termios); err != nil {
		return nil, err
	}

	tty.ss = make(chan os.Signal, 1)

	return tty, nil
}

func (tty *TTY) close() error {
	signal.Stop(tty.ss)
	close(tty.ss)
	return unix.IoctlSetTermios(int(tty.in.Fd()), ioctlWriteTermios, &tty.termios)
}

func (tty *TTY) readRune() (rune, error) {
	r, _, err := tty.bin.ReadRune()
	return r, err
}

func (tty *TTY) buffered() bool {
	return tty.bin.Buffered() > 0
}

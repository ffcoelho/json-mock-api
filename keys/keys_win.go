//go:build windows
// +build windows

package keys

import (
	"context"
	"os"
	"syscall"
	"unsafe"
)

const (
	enableEchoInput = 0x4

	keyEvent              = 0x1
	mouseEvent            = 0x2
	windowBufferSizeEvent = 0x4
)

var kernel32 = syscall.NewLazyDLL("kernel32.dll")

var (
	procReadConsoleInput = kernel32.NewProc("ReadConsoleInputW")
	procGetConsoleMode   = kernel32.NewProc("GetConsoleMode")
	procSetConsoleMode   = kernel32.NewProc("SetConsoleMode")
)

type wchar uint16
type dword uint32
type word uint16

type inputRecord struct {
	eventType word
	_         [2]byte
	event     [16]byte
}

type keyEventRecord struct {
	keyDown         int32
	repeatCount     word
	virtualKeyCode  word
	virtualScanCode word
	unicodeChar     wchar
	controlKeyState dword
}

type TTY struct {
	in                *os.File
	out               *os.File
	st                uint32
	rs                []rune
	sigwinchCtx       context.Context
	sigwinchCtxCancel context.CancelFunc
	readNextKeyUp     bool
}

func readConsoleInput(fd uintptr, record *inputRecord) (err error) {
	var w uint32
	r1, _, err := procReadConsoleInput.Call(fd, uintptr(unsafe.Pointer(record)), 1, uintptr(unsafe.Pointer(&w)))
	if r1 == 0 {
		return err
	}
	return nil
}

func open(path string) (*TTY, error) {
	tty := new(TTY)
	tty.in = os.Stdin
	tty.out = os.Stdout
	h := tty.in.Fd()
	var st uint32
	r1, _, err := procGetConsoleMode.Call(h, uintptr(unsafe.Pointer(&st)))
	if r1 == 0 {
		return nil, err
	}
	tty.st = st

	st &^= enableEchoInput

	// ignore error
	procSetConsoleMode.Call(h, uintptr(st))

	tty.sigwinchCtx, tty.sigwinchCtxCancel = context.WithCancel(context.Background())

	return tty, nil
}

func (tty *TTY) close() error {
	procSetConsoleMode.Call(tty.in.Fd(), uintptr(tty.st))
	tty.sigwinchCtxCancel()
	return nil
}

func (tty *TTY) readRune() (rune, error) {
	if len(tty.rs) > 0 {
		r := tty.rs[0]
		tty.rs = tty.rs[1:]
		return r, nil
	}
	var ir inputRecord
	err := readConsoleInput(tty.in.Fd(), &ir)
	if err != nil {
		return 0, err
	}

	switch ir.eventType {
	case keyEvent:
		kr := (*keyEventRecord)(unsafe.Pointer(&ir.event))
		if kr.keyDown == 0 {
			if kr.unicodeChar != 0 && tty.readNextKeyUp {
				tty.readNextKeyUp = false
				if 0x2000 <= kr.unicodeChar && kr.unicodeChar < 0x3000 {
					return rune(kr.unicodeChar), nil
				}
			}
		} else {
			if kr.unicodeChar > 0 {
				return rune(kr.unicodeChar), nil
			}
			return 0, nil
		}
	}
	return 0, nil
}

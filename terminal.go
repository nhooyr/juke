package main

import (
	"log"
	"os"
	"syscall"
	"unsafe"
)

// xterm terminal escape sequences defined
const (
	CSI         = "\033["
	NORMAL      = CSI + "0m"
	CURSORPOS   = CSI + "6n"
	CURSORVIS   = CSI + "34h\033[?25h"
	CURSORINVIS = CSI + "?25l"
	CURSORADDR  = CSI + "%d;%dH"
	BLUE        = CSI + "34m"
	GREEN       = CSI + "32m"
	RED         = CSI + "31m"
	MAGENTA     = CSI + "35m"
)

func readTermios() (t syscall.Termios) {
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), ioctlReadTermios, uintptr(unsafe.Pointer(&t)), 0, 0, 0); err != 0 {
		log.Fatal(err)
	}
	return
}

func writeTermios(t syscall.Termios) {
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), ioctlWriteTermios, uintptr(unsafe.Pointer(&t)), 0, 0, 0); err != 0 {
		panic(err)
	}
}

func getDimensions() (dimensions [4]uint16) {
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(os.Stdin.Fd()), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0); err != 0 {
		panic(err)
	}
	return
}

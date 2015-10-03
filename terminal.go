package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"
)

// terminal escape sequences defined
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

func getCursorPos() position {
	// get cursor position
	os.Stdin.WriteString(CURSORPOS)
	r := bufio.NewReader(os.Stdin)
	p, err := r.ReadString('R')
	if err != nil {
		log.Fatal(err)
	}
	i := strings.Index(p, ";")
	row, err := strconv.ParseUint(p[2:i], 10, 16)
	if err != nil {
		log.Fatal(err)
	}
	col, err := strconv.ParseUint(p[i+1:len(p)-1], 10, 16)
	if err != nil {
		log.Fatal(err)
	}
	return position{y: uint16(row), x: uint16(col)}
}

func readTermios() (t syscall.Termios) {
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), ioctlReadTermios, uintptr(unsafe.Pointer(&t)), 0, 0, 0); err != 0 {
		log.Fatalln("not a terminal, got:", err)
	}
	return
}

func writeTermios(t syscall.Termios) {
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), ioctlWriteTermios, uintptr(unsafe.Pointer(&t)), 0, 0, 0); err != 0 {
		log.Fatal(err)
	}
}

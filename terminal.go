package main

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

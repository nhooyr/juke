package main

import "os"

// terminal escape sequences defined
var normal = []byte{27, 91, 48, 109}
var cursorVisible = []byte{27, 91, 51, 52, 104, 27, 91, 63, 50, 53, 104}
var cursorInvisible = []byte{27, 91, 63, 50, 53, 108}

// printColor handles the other bits
var blue = []byte("34")
var green = []byte("32")
var red = []byte("31")
var magenta = []byte("35")

func printColor(player uint) {
	var c []byte
	esc := []byte{27, 91}
	switch player {
	case 1:
		c = blue
	case 2:
		c = green
	case 3:
		c = red
	case 4:
		c = magenta
	}
	esc = append(esc, c...)
	esc = append(esc, 109)
	os.Stdout.Write(esc)
}

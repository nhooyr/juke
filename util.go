package main

import (
	"os"
	"strconv"
)

func printColor(i uint) {
	var c string
	switch i {
	case 0:
		c = strconv.FormatUint(34, 10)
	case 1:
		c = strconv.FormatUint(32, 10)
	case 2:
		c = strconv.FormatUint(31, 10)
	case 3:
		c = strconv.FormatUint(35, 10)
	}
	esc := []byte{27, 91}
	esc = append(esc, []byte(c)...)
	esc = append(esc, 109)
	os.Stdout.Write(esc)

}

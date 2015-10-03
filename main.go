package main

import (
	"log"
	"os"
	"syscall"
)

func main() {
	g := new(game)
	log.SetPrefix(NORMAL + "juke: ")
	log.SetFlags(0)
	g.parseFlags()
	// hide cursor
	os.Stdin.WriteString(CURSORINVIS)
	// save current termios
	g.oldTios = g.readTermios()
	g.captureSignals()
	// set raw mode
	raw := g.oldTios
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	g.writeTermios(raw)
	g.setOrigin()
	g.setDimensions()
	// start game
	g.start()
}

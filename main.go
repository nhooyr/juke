package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	g := new(game)
	log.SetPrefix("juke: ")
	log.SetFlags(0)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Stderr.WriteString(`
Controls:
  Pn is up, left, down, right
        P1 is ↑←↓→ (arrow keys)
        P2 is wasd
        P3 is yghj
        P4 is pl;'
  t to pause (and unpause)
  r to restart
  q to quit
`)
	}
	tmph := flag.Uint("h", 0, "height of playground (default height of tty)")
	tmpw := flag.Uint("w", 0, "width of playground (default width of tty)")
	tmpi := flag.Uint("i", 3, "initital size of snake")
	tmpp := flag.Uint("p", 1, "number of players")
	tmps := flag.Int64("s", 20, "unit's per second for snake")
	tmpf := flag.Int64("f", 1, "how many blocks each food adds")
	flag.Parse()
	if g.players > 4 {
		log.Fatal("cannot be more than 4 players")
	}
	g.h = uint16(*tmph)
	g.w = uint16(*tmpw)
	g.init = uint16(*tmpi)
	if g.init == 0 {
		log.Fatal("initial size of snake cannot be 0")
	}
	g.players = uint16(*tmpp)
	g.foodVal = uint16(*tmpf)
	g.speed = time.Duration(*tmps)

	// hide cursor
	os.Stdin.WriteString(CURSORINVIS)
	/// save current termios
	old := readTermios()
	cleanup := func() {
		// restore text to normal
		os.Stdout.WriteString(NORMAL)
		// make cursor visible
		os.Stdin.WriteString(CURSORVIS)
		// set tty to normal
		writeTermios(old)
	}
	// capture signals
	g.sigs = make(chan os.Signal)
	signal.Notify(g.sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-g.sigs
		os.Stdout.WriteString("\n")
		cleanup()
		os.Exit(0)
	}()
	raw := old
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	writeTermios(raw)
	g.origin = getCursorPos()
	g.setDimensions()
	if g.w < 4 || g.h < 4 {
		cleanup()
		log.Fatal("width or height cannot be less than 4")
	}
	maxInit := g.w / 3
	if g.init > maxInit {
		cleanup()
		log.Fatalln("init too big, max init size for this width is", maxInit)
	}
	log.SetPrefix(NORMAL + "juke: ")
	// start game
	g.start()
}

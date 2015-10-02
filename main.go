package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"
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
	tmps := flag.Int64("s", 20, "unit's per second for snake")
	tmpf := flag.Int64("f", 1, "how many blocks each food adds")
	flag.UintVar(&g.players, "p", 1, "number of players")
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
	g.foodVal = uint16(*tmpf)
	g.speed = time.Duration(*tmps)

	// hide cursor
	os.Stdin.WriteString(CURSORINVIS)

	/// save current termios
	var old syscall.Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), ioctlReadTermios, uintptr(unsafe.Pointer(&old)), 0, 0, 0); err != 0 {
		log.Fatalln("not a terminal, got:", err)
	}
	cleanup := func() {
		// restore text to normal
		os.Stdout.WriteString(NORMAL)
		// make cursor visible
		os.Stdin.WriteString(CURSORVIS)
		// set tty to normal
		if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), ioctlWriteTermios, uintptr(unsafe.Pointer(&old)), 0, 0, 0); err != 0 {
			log.Fatal(err)
		}
	}
	// capture signals
	g.sigs = make(chan os.Signal)
	signal.Notify(g.sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-g.sigs
		cleanup()
		os.Stdout.WriteString("\n")
		os.Exit(0)
	}()
	// set raw mode
	raw := old
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), ioctlWriteTermios, uintptr(unsafe.Pointer(&raw)), 0, 0, 0); err != 0 {
		log.Fatal(err)
	}

	g.setDimensions()
	if g.w < 6 || g.h < 4 {
		cleanup()
		log.Fatal("width or height cannot be less than 4")
	}
	maxInit := g.w / 3
	if g.init > maxInit {
		cleanup()
		log.Fatalln("init too big, max init size for this width is", maxInit)
	}
	log.SetPrefix(NORMAL + "juke: ")
	// initialize game
	g.initialize()
	t := time.NewTimer(time.Second / g.speed)
	for ;;t.Reset(time.Second / g.speed) {
		select {
		case <-g.restart:
			g.nextGame()
			continue
		case <-g.pause:
			select {
			case <-g.pause:
				// unpause
			case <-g.restart:
				g.nextGame()
			}
		case <-t.C:
			// next frame time
		}
		g.checkFood()
		g.moveSnakes()
		g.checkCollisions()
		g.updateSnakes()
		g.moveTo(position{g.h - 1, g.w - 1})
	}
}

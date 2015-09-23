package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"unsafe"
)

func main() {
	g := new(game)
	log.SetPrefix("goSnake: ")
	log.SetFlags(0)
	tmph := flag.Uint("h", 0, "height of playground (default height of tty)")
	tmpw := flag.Uint("w", 0, "width of playground, (default width of tty)")
	tmpi := flag.Uint("i", 3, "initital size of snake")
	tmps := flag.Int64("s", 20, "unit's per second for snake")
	flag.UintVar(&g.players, "p", 1, "number of players; controls: P1: arrows, P2: wasd, P3: yghj, P4: pl;'")
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
	g.speed = time.Duration(*tmps)

	// hide cursor
	os.Stdin.Write([]byte{27, 91, 63, 50, 53, 108})

	/// save current termios
	var old syscall.Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCGETA, uintptr(unsafe.Pointer(&old)), 0, 0, 0); err != 0 {
		log.Fatalln("not a terminal, got:", err)
	}
	cleanup := func() {
		// make cursor visible
		os.Stdin.Write([]byte{27, 91, 51, 52, 104, 27, 91, 63, 50, 53, 104})
		// set tty to normal
		if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCSETA, uintptr(unsafe.Pointer(&old)), 0, 0, 0); err != 0 {
			log.Fatal(err)
		}
	}
	// capture signals
	g.sigs = make(chan os.Signal)
	signal.Notify(g.sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-g.sigs
		cleanup()
		os.Exit(0)
	}()
	// set raw mode
	raw := old
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCSETA, uintptr(unsafe.Pointer(&raw)), 0, 0, 0); err != 0 {
		log.Fatal(err)
	}

	g.setDimensions()

	var maxInit uint16
	if g.players == 1 {
		maxInit = g.w/3 - 1
	} else {
		maxInit = g.w/3 - 1
	}
	if g.init > maxInit {
		log.Println("init too big, max init size for this h/w is", maxInit)
		cleanup()
		os.Exit(0)
	}
	// start game
	g.initialize()
	for {
		g.printSnakes()
		g.moveTo(position{g.h - 1, g.w - 1})
		time.Sleep(time.Second / g.speed)
		g.moveSnakes()
	}
}

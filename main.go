package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

type game struct {
	h         uint16
	w         uint16
	rowOffSet uint16
	sigs      chan os.Signal
	s         snake
	food      position
	init      uint16
	origin    position
	speed     time.Duration
}

func (g *game) initialize() {
	g.s.g = g
	g.s.initialize()
	g.addFood()
}

func (g *game) getValidFoodPos() (vp []position) {
	vp = []position{}
	for i := uint16(1); i < g.h-1; i++ {
		for j := uint16(1); j < g.w-1; j++ {
			if g.s.isNotOn(position{i, j}) {
				vp = append(vp, position{y: uint16(i), x: uint16(j)})
			}
		}
	}
	return
}

func (g *game) addFood() {
	vp := g.getValidFoodPos()
	rand.Seed(time.Now().UnixNano())
	g.food = vp[rand.Intn(len(vp))]
	g.moveTo(g.food)
	fmt.Print("+")
}

func main() {
	g := new(game)
	log.SetPrefix("goSnake: ")
	log.SetFlags(0)
	tmph := flag.Uint("h", 0, "height of playground")
	tmpw := flag.Uint("w", 0, "width of playground")
	tmpi := flag.Uint("i", 1, "initital size of snake")
	tmps := flag.Int64("s", 10, "unit's per second for snake")
	flag.Parse()

	g.h = uint16(*tmph)
	g.w = uint16(*tmpw)
	g.init = uint16(*tmpi)
	g.speed = time.Duration(*tmps)

	if g.init == 0 {
		log.Fatal("initial size of snake cannot be 0")
	}

	// hide cursor
	os.Stdin.Write([]byte{27, 91, 63, 50, 53, 108})

	/// save current termios
	var old syscall.Termios
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCGETA, uintptr(unsafe.Pointer(&old)), 0, 0, 0); err != 0 {
		log.Fatalln("not a terminal, got:", err)
	}
	// capture signals
	g.sigs = make(chan os.Signal)
	signal.Notify(g.sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-g.sigs

		// make cursor visible
		os.Stdin.Write([]byte{27, 91, 51, 52, 104, 27, 91, 63, 50, 53, 104})

		// set tty to normal
		if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, os.Stdin.Fd(), syscall.TIOCSETA, uintptr(unsafe.Pointer(&old)), 0, 0, 0); err != 0 {
			log.Fatal(err)
		}

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
	if g.h < g.w {
		maxInit = g.h/2 - 1
	} else {
		maxInit = g.w/2 - 1
	}
	if g.init > maxInit {
		log.Println("init too big, max init size for this h/w is", maxInit)
		g.sigs <- syscall.SIGTERM
		g.sigs <- syscall.SIGTERM
	}

	// start game
	g.printGround()
	g.initialize()
	go g.s.processInput()
	for {
		g.s.print()
		g.moveTo(position{g.h - 1, g.w - 1})
		time.Sleep(time.Second / g.speed)
		g.s.move()
	}
}

func (g *game) setDimensions() {
	// get cursor position
	os.Stdin.Write([]byte{27, 91, 54, 110})
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
	g.origin = position{y: uint16(row), x: uint16(col)}
	// get dimensions and check if offset needed
	var dimensions [4]uint16
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(os.Stdin.Fd()), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0); err != 0 {
		log.Fatal(err)
	}
	if g.h == 0 {
		g.h = dimensions[0]
	}
	if g.w == 0 {
		g.w = dimensions[1]
	}
	if i := uint16(row) + g.h; i > dimensions[0] {
		g.rowOffSet = i - dimensions[0] - 1
	}
}

// print current ground
func (g *game) printGround() {
	g.moveTo(position{g.rowOffSet, 0})
	for i := uint16(0); i < g.h; i++ {
		for j := uint16(0); j < g.w; j++ {
			switch {
			case (i == g.h-1 || i == 0) && (j == 0 || j == g.w-1):
				fmt.Print("┼")
			case i == 0 || i == g.h-1:
				fmt.Print("─")
			case j == 0 || j == g.w-1:
				fmt.Print("│")
			default:
				g.moveTo(position{i + g.rowOffSet, g.w - 1})
			}
		}
		if i < g.h-1 {
			fmt.Print("\n")
		}
	}
}

func (g *game) moveTo(p position) {
	esc := []byte{27, 91}
	esc = append(esc, []byte(strconv.FormatUint(uint64(p.y+g.origin.y-g.rowOffSet), 10))...)
	esc = append(esc, 59)
	esc = append(esc, []byte(strconv.FormatUint(uint64(p.x+g.origin.x), 10))...)
	esc = append(esc, 72)
	os.Stdin.Write(esc)
}

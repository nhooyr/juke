package main

import (
	"bufio"
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
	ground    [][]string
	h         uint16
	w         uint16
	rowOffSet uint16
	sigs      chan os.Signal
	s         snake
	input     chan uint16
	food      block
	init      int
}

func (g *game) initialize() {
	g.input = make(chan uint16, 2)
	g.init = 5
	g.s = make([]block, 1)
	g.s[0].dir = left
	g.s[0].pdir = left
	g.s[0].pos.x, g.s[0].pos.y = g.w-uint16(g.init)-2, g.h/2
	for i := g.init; i > 0; i-- {
		b := new(block)
		b.pos = g.s[len(g.s)-1].pos
		switch b.dir = g.s[len(g.s)-1].dir; b.dir {
		case up:
			b.pos.y += 1
		case right:
			b.pos.x -= 1
		case down:
			b.pos.y -= 1
		case left:
			b.pos.x += 1
		}
		g.s = append(g.s, *b)
		g.s[len(g.s)-1].pdir = left
	}
	g.ground = make([][]string, g.h)
	for i := 0; i < len(g.ground); i++ {
		g.ground[i] = make([]string, g.w)
		for j := 0; j < len(g.ground[i]); j++ {
			if (i == len(g.ground)-1 || i == 0) && (j == 0 || j == len(g.ground[i])-1) {
				g.ground[i][j] = "┼"
			} else if i == len(g.ground)-1 || i == 0 {
				g.ground[i][j] = "─"
			} else if j == 0 || j == len(g.ground[i])-1 {
				g.ground[i][j] = "│"
			} else {
				g.ground[i][j] = " "
			}
		}
	}
	g.addFood()
}

func (g *game) addFood() {
	for {
		n := rand.Int()%(len(g.ground)-2) + 1
		m := rand.Int()%(len(g.ground[0])-2) + 1
		if g.ground[n][m] == " " {
			g.ground[n][m] = "+"
			g.food.pos.y = uint16(n)
			g.food.pos.x = uint16(m)
			return
		}
	}
}

//TODO BREAK INTO SEPARATE METHODS AND LOOK AT tput cm
func main() {
	g := new(game)
	log.SetPrefix("goSnake: ")

	os.Stdin.Write([]byte{27, 55})

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

	// get cursor position
	os.Stdin.Write([]byte{27, 91, 54, 110})
	r := bufio.NewReader(os.Stdin)
	p, err := r.ReadString('R')
	if err != nil {
		log.Fatal(err)
	}
	i := strings.Index(p, ";")
	rowt, err := strconv.ParseUint(p[2:i], 10, 16)
	if err != nil {
		log.Fatal(err)
	}
	row := uint16(rowt)

	// get dimensions and check if offset needed
	var dimensions [4]uint16
	if _, _, err := syscall.Syscall6(syscall.SYS_IOCTL, uintptr(os.Stdin.Fd()), uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&dimensions)), 0, 0, 0); err != 0 {
		log.Fatal(err)
	}
	tmp, err := strconv.ParseUint(os.Args[1], 10, 16)
	if err != nil {
		log.Println(err)
	}
	g.h = uint16(tmp)
	if i := row + g.h; i > dimensions[0] {
		g.rowOffSet = i - dimensions[0] - 1
	}
	tmp, err = strconv.ParseUint(os.Args[2], 10, 16)
	if err != nil {
		log.Println(err)
	}
	g.w = uint16(tmp)

	// start game
	g.initialize()
	go g.processInput()
	for {
		g.updateGround()
		g.printGround()
		time.Sleep(500 * time.Millisecond)
		g.restoreCursor()
	}
}

// TODO make more responsive
func (g *game) updateGround() {
	select {
	case in := <-g.input:
		g.s[0].dir = in
		g.s[0].pdir = in
	default:
		// no input
	}
	for i, _ := range g.s {
		g.ground[g.s[i].pos.y][g.s[i].pos.x] = " "
		if g.s[i].dir == up {
			g.s[i].pos.y -= 1
		} else if g.s[i].dir == right {
			g.s[i].pos.x += 1
		} else if g.s[i].dir == down {
			g.s[i].pos.y += 1
		} else if g.s[i].dir == left {
			g.s[i].pos.x -= 1
		}
		g.ground[g.s[i].pos.y][g.s[i].pos.x] = "="
		if i != 0 {
			g.s[i].pdir = g.s[i].dir
			g.s[i].dir = g.s[i-1].pdir
		}
	}
	if g.food.pos == g.s[0].pos {
		g.addFood()
		b := new(block)
		b.pos = g.s[len(g.s)-1].pos
		switch b.dir = g.s[len(g.s)-1].dir; b.dir {
		case up:
			b.pos.y += 1
		case right:
			b.pos.x -= 1
		case down:
			b.pos.y -= 1
		case left:
			b.pos.x += 1
		}
		g.s = append(g.s, *b)
		g.ground[b.pos.y][b.pos.x] = "="
	}
}

func (g *game) processInput() {
	b := make([]byte, 3)
	var prevIn byte
	for {
		_, err := os.Stdin.Read(b)
		if err != nil {
			log.Print(err)
			g.sigs <- syscall.SIGTERM
		}
		if b[0] == 27 && b[1] == 91 {
			if b[2] == prevIn {
				continue
			}
			if b[2] == 65 {
				g.input <- up
				prevIn = 65
			} else if b[2] == 67 {
				g.input <- right
				prevIn = 67
			} else if b[2] == 66 {
				g.input <- down
				prevIn = 66
			} else if b[2] == 68 {
				g.input <- left
				prevIn = 68
			}
		}
	}
}

// restore cursor position
func (g *game) restoreCursor() {
	os.Stdin.Write([]byte{27, 56})
	if g.rowOffSet != 0 {
		for i := uint16(0); i < g.rowOffSet; i++ {
			os.Stdin.Write([]byte{27, 77})
		}
	}
}

// print current ground
func (g *game) printGround() {
	for i := 0; i < len(g.ground); i++ {
		for j := 0; j < len(g.ground[i]); j++ {
			fmt.Print(g.ground[i][j])
		}
		if i < len(g.ground)-1 {
			fmt.Print("\n")
		}
	}
}

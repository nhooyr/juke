package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
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
	s         []snake
	food      position
	init      uint16
	origin    position
	speed     time.Duration
	players   uint
}

func (g *game) initialize() {
	g.printGround()
	g.s = make([]snake, g.players)
	for i := uint(0); i < g.players; i++ {
		g.s[i].g = g
		g.s[i].initialize(i+1)
	}
	g.addFood()
}

func (g *game) getValidFoodPos() (vp []position) {
	vp = []position{}
	for i := uint16(1); i < g.h-1; i++ {
		for j := uint16(1); j < g.w-1; j++ {
			for s := uint(0); s < g.players; s++ {
				if g.s[s].on(position{i, j}) {
					break
				}
				vp = append(vp, position{y: i, x: j})
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

// process the input
func (s *snake) processInput() {
	b := make([]byte, 3)
	var prevDir uint16
	for {
		_, err := os.Stdin.Read(b)
		if err != nil {
			log.Print(err)
			s.g.sigs <- syscall.SIGTERM
		}
		if b[0] == 27 && b[1] == 91 {
			dir := uint16(1 << (b[2] - 65))
			if dir == prevDir {
				continue
			}
			switch dir {
			case up, down, right, left:
				s.input <- dir
				prevDir = dir
			}
		}
	}
}

func (g *game) printSnakes() {
	for i := uint(0); i < g.players; i++ {
		g.s[i].print()
	}
}

func (g *game) moveSnakes() {
	for i := uint(0); i < g.players; i++ {
		g.s[i].move()
	}
}

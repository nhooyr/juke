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
	players   uint
	origin    position
	speed     time.Duration
}

func (g *game) initialize() {
	g.printGround()
	g.s = make([]snake, g.players)
	for i := uint(0); i < g.players; i++ {
		g.s[i].g = g
		g.s[i].initialize(i + 1)
	}
	g.addFood()
	go g.processInput()
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
func (g *game) processInput() {
	b := make([]byte, 1)
	var prevDir = make([]uint16, g.players)
	read := func() {
		_, err := os.Stdin.Read(b)
		if err != nil {
			panic(err)
		}
	}
	defer func() {
		log.Println(recover())
		g.sigs <- syscall.SIGTERM
	}()
	for {
		read()
		if b[0] == 27 && g.s[0].dead == false {
			read()
			if b[0] == 91 {
				read()
				// special trick to make things easier, 65 is up, 66 is down, 67 is right and 68 is left so if you subtract 65 and shift the bits in 1 by it you get the exact direction!
				dir := uint16(1 << (b[0] - 65))
				if dir == prevDir[0] {
					continue
				}
				switch dir {
				case up, down, right, left:
					g.s[0].input <- dir
					prevDir[0] = dir
				}
			}
		}
		if g.players > 1 && g.s[1].dead == false {
			switch b[0] {
			case 'w':
				if prevDir[1] == up {
					continue
				}
				g.s[1].input <- up
			case 'd':
				if prevDir[1] == right {
					continue
				}
				g.s[1].input <- right
			case 's':
				if prevDir[1] == down {
					continue
				}
				g.s[1].input <- down
			case 'a':
				if prevDir[1] == left {
					continue
				}
				g.s[1].input <- left
			}
		}
		if g.players > 2 && g.s[2].dead == false {
			switch b[0] {
			case 'y':
				if prevDir[2] == up {
					continue
				}
				g.s[2].input <- up
			case 'j':
				if prevDir[2] == right {
					continue
				}
				g.s[2].input <- right
			case 'h':
				if prevDir[2] == down {
					continue
				}
				g.s[2].input <- down
			case 'g':
				if prevDir[2] == left {
					continue
				}
				g.s[2].input <- left
			}
		}
		if g.players > 3 && g.s[3].dead == false {
			switch b[0] {
			case 'p':
				if prevDir[3] == up {
					continue
				}
				g.s[3].input <- up
			case '\'':
				if prevDir[3] == right {
					continue
				}
				g.s[3].input <- right
			case ';':
				if prevDir[3] == down {
					continue
				}
				g.s[3].input <- down
			case 'l':
				if prevDir[3] == left {
					continue
				}
				g.s[3].input <- left
			}
		}
	}
}

func (g *game) printSnakes() {
	for i := uint(0); i < g.players; i++ {
		if g.s[i].dead == false {
			g.s[i].print()
		}
	}
}

func (g *game) moveSnakes() {
	for i := uint(0); i < g.players; i++ {
		if g.s[i].dead == false {
			g.s[i].move()
		}
	}
	for i, _ := range g.s {
		if g.s[i].dead == true {
			continue
		}
		for j, _ := range g.s {
			if j != i && (g.s[j].on(g.s[i].bs[0].pos) || g.s[i].onExceptFirst(g.s[i].bs[0].pos)){
				g.s[i].die()
			}
		}
	}
}

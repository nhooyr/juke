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
	foodVal   uint16
	food      []position
	init      uint16
	players   uint
	origin    position
	speed     time.Duration
	restart   chan struct{}
}

func (g *game) start(i uint) {
	g.s[i].g = g
	g.s[i].player = i
	g.s[i].initialize()
	g.addFood(int(i))
}
func (g *game) initialize() {
	g.restart = make(chan struct{})
	g.s = make([]snake, g.players)
	g.food = make([]position, g.players)
	g.printGround()
	for i := uint(0); i < g.players; i++ {
		g.start(i)
	}
	g.printSnakes()
	go g.processInput()
}

func (g *game) getValidFoodPos() (vp []position) {
	vp = []position{}
	for y := uint16(1); y < g.h-1; y++ {
	xLoop:
		for x := uint16(1); x < g.w-1; x++ {
			for s := uint(0); s < g.players; s++ {
				for f := 0; uint(f) < g.players; f++ {
					if g.food[f] == (position{y, x}) || g.s[s].on(position{y, x}, 0, len(g.s[s].bs), 1) {
						continue xLoop
					}
				}
			}
			vp = append(vp, position{y, x})
		}
	}
	return
}

func (g *game) addFood(i int) {
	vp := g.getValidFoodPos()
	if len(vp) == 0 {
		return
	}
	rand.Seed(time.Now().UnixNano())
	g.food[i] = vp[rand.Intn(len(vp))]
	g.moveTo(g.food[i])
	os.Stdout.WriteString(NORMAL)
	os.Stdout.WriteString("A")
}

func (g *game) setDimensions() {
	// get cursor position
	os.Stdin.WriteString(CURSORPOS)
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

func (g *game) clear(i uint) {
	g.s[i].printOverAll(" ")
	g.s[i].dead = false
	g.moveTo(g.food[i])
	os.Stdout.WriteString(" ")
}

// print current ground
func (g *game) printGround() {
	for y := uint16(0); y < g.h; y++ {
		for x := uint16(0); x < g.w; x++ {
			switch {
			case (y == g.h-1 || y == 0) && (x == 0 || x == g.w-1):
				os.Stdout.WriteString("┼")
			case y == 0 || y == g.h-1:
				os.Stdout.WriteString("─")
			case x == 0 || x == g.w-1:
				os.Stdout.WriteString("│")
			default:
				g.moveTo(position{y + g.rowOffSet, g.w - 1})
			}
		}
		if y < g.h-1 {
			os.Stdout.WriteString("\n")
		}
	}
}

func (g *game) moveTo(p position) {
	os.Stdin.WriteString(fmt.Sprintf(CURSORADDR, p.y+g.origin.y-g.rowOffSet, p.x+g.origin.x))
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
		if r := recover(); r != nil {
			log.Println(r)
			g.sigs <- syscall.SIGTERM
		}
	}()
	changeDir := func(s uint, dir uint16) {
		if g.players <= s || g.s[s].dead == true || prevDir[s] == dir {
			return
		}
		g.s[s].Lock()
		g.s[s].bs[0].d = dir
		g.s[s].Unlock()
	}
	for {
		read()
		if b[0] == 27 && g.s[0].dead == false {
			read()
			if b[0] == 91 {
				read()
				dir := uint16(b[0] - 64)
				switch dir {
				case up, down, right, left:
					changeDir(0, dir)
				}
			}
		} else {
			switch b[0] {
			case 'w':
				changeDir(1, up)
			case 'd':
				changeDir(1, right)
			case 's':
				changeDir(1, down)
			case 'a':
				changeDir(1, left)
			case 'y':
				changeDir(2, up)
			case 'j':
				changeDir(2, right)
			case 'h':
				changeDir(2, down)
			case 'g':
				changeDir(2, left)
			case 'p':
				changeDir(3, up)
			case '\'':
				changeDir(3, right)
			case ';':
				changeDir(3, down)
			case 'l':
				changeDir(3, left)
			case 'r':
				g.restart <- struct{}{}
				<-g.restart
			case 'q':
				g.sigs <- syscall.SIGTERM
				return
			}
		}
	}
}

func (g *game) updateSnakes() {
	for i := uint(0); i < g.players; i++ {
		if g.s[i].dead == false {
			g.s[i].update()
		}
	}
}

func (g *game) printSnakes() {
	for i := uint(0); i < g.players; i++ {
		if g.s[i].dead == false {
			g.s[i].printOverAll("=")
		}
	}
}

func (g *game) moveSnakes() {
	for i := uint(0); i < g.players; i++ {
		if g.s[i].dead == false {
			g.s[i].move()
		}
	}
}

func (g *game) checkFood() {
	for i := uint(0); i < g.players; i++ {
		for j := 0; uint(j) < g.players; j++ {
			if g.s[i].bs[0].p == g.food[j] {
				g.s[i].Lock()
				g.s[i].appendBlocks(g.foodVal)
				g.s[i].Unlock()
				g.addFood(j)
			}
		}
	}
}

func (g *game) checkCollisions() {
	var min, end, inc int
	setRand := func(min, end, inc *int) {
		rand.Seed(time.Now().UnixNano())
		if rand.Intn(2) == 0 {
			*inc = 1
			*min = 0
			*end = int(g.players)
		} else {
			*inc = -1
			*min = int(g.players) - 1
			*end = -1
		}

	}
	setRand(&min, &end, &inc)
	for i := min; i != end; i += inc {
		if g.s[i].dead == true {
			continue
		}
		var inc, end, min int
		setRand(&min, &end, &inc)
		for j := min; j != end; j += inc {
			if j != i {
				// first check if any of j is on the first block of i, then if len of i's bs is just one, make sure their first elements are opposite dir and then check if i is on any of j's oldBs or if i is on any of j's new Bs (this is needed for when one is len of just 1 and the other is greater, eg 2)
				if g.s[j].on(g.s[i].bs[0].p, 0, len(g.s[j].bs), 1) || (len(g.s[i].bs) == 1 && oppositeDir(g.s[i].bs[0].d, g.s[j].bs[0].d) && (g.s[i].on(g.s[j].oldBs[0].p, 0, len(g.s[i].bs), 1) || g.s[j].on(g.s[i].oldBs[0].p, 0, len(g.s[j].bs), 1))) {
					g.s[i].die()
				}
			}
		}
	}
}

func oppositeDir(d1, d2 uint16) bool {
	if d1%2 == 1 {
		if d1+1 == d2 {
			return true
		}
	} else {
		if d1-1 == d2 {
			return true
		}
	}
	return false
}

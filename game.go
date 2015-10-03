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
)

type game struct {
	h         uint16
	w         uint16
	hf        float64
	wf        float64
	rowOffSet uint16
	sigs      chan os.Signal
	s         []snake
	foodVal   uint16
	f         *food
	init      uint16
	players   uint16
	origin    position
	speed     time.Duration
	restart   chan struct{}
	pause     chan struct{}
	oldTios   syscall.Termios
}

func (g *game) captureSignals() {
	g.sigs = make(chan os.Signal)
	signal.Notify(g.sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-g.sigs
		os.Stdout.WriteString("\n")
		g.cleanup()
		os.Exit(0)
	}()
}

func (g *game) parseFlags() {
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

Rules:
    1. You can go through the walls, you will appear on the other side.
    2. You can go through yourself.
    3. If you touch anyone else with your head you die. E.g

       `+BLUE+`=
     `+ RED+`==`+BLUE+`=
       =

    `+NORMAL+`Then

       `+BLUE+`=
     `+RED+`xx`+BLUE+`=
       =

    `+NORMAL+`4. If both heads hit each other, both snakes die. E.g

    `+RED+`==`+BLUE+`==

    `+NORMAL+`Then

    `+RED+`xx`+BLUE+`xx

    `+NORMAL+`5. If they both don't hit each other but are going to land
    on the same square. E.g

    `+RED+`=`+NORMAL+`_`+BLUE+`=

    `+NORMAL+`(_ is square they are both going to land on)

    Then randomly one is chosen to get the square and the other dies.

    `+RED+`x`+BLUE+`=

    `+NORMAL+`6. Food increases the length of the snake but not instantly,
    it begins to grow by the value of the food, this is best understood by
    actually playing the game.
`)
	}
	tmph := flag.Uint("h", 0, "height of playground (default height of tty)")
	tmpw := flag.Uint("w", 0, "width of playground (default width of tty)")
	tmpi := flag.Uint("i", 3, "initital size of snake")
	tmpp := flag.Uint("p", 2, "number of players")
	tmps := flag.Uint("s", 30, "unit's per second for snake")
	tmpf := flag.Uint("f", 5, "how many blocks each food adds")
	flag.Parse()
	g.h = uint16(*tmph)
	g.w = uint16(*tmpw)
	g.init = uint16(*tmpi)
	if g.init == 0 {
		log.Fatal("initial size of snake cannot be 0")
	}
	g.players = uint16(*tmpp)
	if g.players > 4 {
		log.Fatal("cannot have more than 4 players")
	}
	g.foodVal = uint16(*tmpf)
	if g.foodVal == 0 {
		log.Fatal("food value cannot be 0")
	}
	g.speed = time.Duration(*tmps)
	if g.speed == 0 {
		log.Fatal("speed cannot be 0 units per second")
	}

}

func (g *game) next() {
	for i := uint16(0); i < g.players; i++ {
		g.s[i].clear()
		g.f.clearFood(i)
		g.s[i].initialize()
	}
	for i := uint16(0); i < g.players; i++ {
		g.f.fillVP(i + 1)
		if len(g.f.vp) != 0 {
			g.f.addFood(i)
			g.f.printFood(i)
		} else {
			g.f.p[i] = position{}
			break
		}
	}
	g.printSnakes()
}

func (g *game) loop() {
	for t := time.NewTimer(time.Second / g.speed); ; t.Reset(time.Second / g.speed) {
		select {
		case <-g.restart:
			g.next()
			continue
		case <-g.pause:
			select {
			case <-g.pause:
			case <-g.restart:
				g.next()
			}
			continue
		case <-t.C:
			// next frame time
		}
		g.moveSnakes()
		g.checkCollisions()
		g.updateSnakes()
		g.checkFood()
		g.moveTo(position{g.h - 1, g.w - 1})
	}
}

func (g *game) initialize() {
	g.restart = make(chan struct{})
	g.pause = make(chan struct{})
	g.s = make([]snake, g.players)
	g.f = new(food)
	g.f.p = make([]position, g.players)
	g.f.g = g
	for i := uint16(0); i < g.players; i++ {
		g.s[i].g = g
		g.s[i].player = i
		g.s[i].initialize()
	}
	for i := uint16(0); i < g.players; i++ {
		g.f.fillVP(0)
		if len(g.f.vp) != 0 {
			g.f.addFood(i)
		} else {
			break
		}
	}
}

func (g *game) start() {
	g.initialize()
	g.printGround()
	g.printSnakes()
	g.printAllFood()
	go g.processInput()
	g.loop()
}

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
	changeDir := func(s uint16, dir uint16) {
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
			case 'w', 'W':
				changeDir(1, up)
			case 'd', 'D':
				changeDir(1, right)
			case 's', 'S':
				changeDir(1, down)
			case 'a', 'A':
				changeDir(1, left)
			case 'y', 'Y':
				changeDir(2, up)
			case 'j', 'J':
				changeDir(2, right)
			case 'h', 'H':
				changeDir(2, down)
			case 'g', 'G':
				changeDir(2, left)
			case 'p', 'P':
				changeDir(3, up)
			case '\'', '|':
				changeDir(3, right)
			case ';', ':':
				changeDir(3, down)
			case 'l', 'L':
				changeDir(3, left)
			case 't', 'T':
				g.pause <- struct{}{}
			case 'r', 'R':
				g.restart <- struct{}{}
			case 'q', 'Q':
				g.sigs <- syscall.SIGTERM
			}
		}
	}
}

func (g *game) updateSnakes() {
	for i := uint16(0); i < g.players; i++ {
		g.s[i].update()
	}
}

func (g *game) printSnakes() {
	for i := uint16(0); i < g.players; i++ {
		g.s[i].printOverAll("=")
	}
}

func (g *game) printAllFood() {
	for i := uint16(0); i < g.players; i++ {
		g.f.printFood(i)
	}
}

func (g *game) moveSnakes() {
	for i := uint16(0); i < g.players; i++ {
		g.s[i].move()
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
				// first check if any of j is on the first block of i,
				// then if len of i's bs is one, make sure their first
				// elements are opposite dir and then check if i is on any
				// of j's oldBs or if j is on any of i's oldBs
				if g.s[j].on(g.s[i].bs[0].p, 0, len(g.s[j].bs), 1) ||
					(len(g.s[i].bs) == 1 && oppositeDir(g.s[i].bs[0].d, g.s[j].bs[0].d) &&
						(g.s[i].on(g.s[j].oldBs[0].p, 0, len(g.s[i].bs), 1) ||
							g.s[j].on(g.s[i].oldBs[0].p, 0, len(g.s[j].bs), 1))) {
					g.s[i].die()
					g.checkCollisions()
					return
				}
			}
		}
	}
}

func (g *game) isUsed(p position) bool {
	for i := uint16(0); i < g.players; i++ {
		if g.s[i].on(p, 0, len(g.s[i].bs), 1) || g.f.p[i] == p {
			return true
		}
	}
	return false
}

func (g *game) checkFood() {
	for i := uint16(0); i < g.players; i++ {
		for j := uint16(0); j < g.players; j++ {
			if g.s[i].bs[0].p == g.f.p[j] {
				g.s[i].queued += g.foodVal
				g.f.fillVP(j + 1)
				if len(g.f.vp) != 0 {
					g.f.addFood(j)
					g.f.printFood(j)
				} else {
					g.f.p[j] = position{}
				}
			}
		}
	}
	for i := uint16(0); i < g.players; i++ {
		if g.f.p[i] == (position{}) {
			g.f.fillVP(0)
			if len(g.f.vp) != 0 {
				g.f.addFood(i)
				g.f.printFood(i)
			}
		}
	}
}

// get dimensions and check if rowOffSet needed
func (g *game) setDimensions() {
	dimensions := getDimensions()
	if g.h == 0 {
		g.h = dimensions[0]
	}
	if g.w == 0 {
		g.w = dimensions[1]
	}
	if g.w < 4 || g.h < 4 {
		panic("width or height cannot be less than 4")
	}
	g.wf = float64(g.w - 1)
	if g.init > uint16(g.wf/3) {
		panic(fmt.Sprintf("max init size of snake for this width is %d", uint16(g.wf/3)))
	}
	g.hf = float64(g.h - 1)
	if i := uint16(g.origin.y) + g.h; i > dimensions[0] {
		g.rowOffSet = i - dimensions[0] - 1
	}
}

func (g *game) setTTY() {
	os.Stdin.WriteString(CURSORINVIS)
	raw := g.oldTios
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	writeTermios(raw)
}

func (g *game) cleanup() {
	// make text normal and cursor visible
	os.Stdin.WriteString(NORMAL + CURSORVIS)
	// set tty to normal
	writeTermios(g.oldTios)
}

func (g *game) setOrigin() {
	os.Stdin.WriteString(CURSORPOS)
	r := bufio.NewReader(os.Stdin)
	p, err := r.ReadString('R')
	if err != nil {
		panic(err)
	}
	i := strings.Index(p, ";")
	row, err := strconv.ParseUint(p[2:i], 10, 16)
	if err != nil {
		panic(err)
	}
	col, err := strconv.ParseUint(p[i+1:len(p)-1], 10, 16)
	if err != nil {
		panic(err)
	}
	g.origin = position{uint16(row), uint16(col)}
}

func (g *game) moveTo(p position) {
	os.Stdin.WriteString(fmt.Sprintf(CURSORADDR, p.y+g.origin.y-g.rowOffSet, p.x+g.origin.x))
}

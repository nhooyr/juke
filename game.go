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
	"sync"
	"syscall"
	"time"
)

type game struct {
	h          uint16
	w          uint16
	hf         float64
	wf         float64
	rowOffSet  uint16
	sigs       chan os.Signal
	s          []*snake
	foodVal    uint16
	f          *food
	init       uint16
	players    uint16
	origin     position
	speed      time.Duration
	restart    chan struct{}
	pauseLoop  chan struct{}
	pauseInput chan struct{}
	pausedLoop bool
	oldTios    syscall.Termios
	sync.Mutex
	speedM sync.Mutex
}

func (g *game) captureSignals() {
	g.sigs = make(chan os.Signal)
	g.pauseLoop = make(chan struct{})
	signal.Notify(g.sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
	go func() {
		for {
			s := <-g.sigs
			g.Lock()
			if !g.pausedLoop {
				g.pauseLoop <- struct{}{}
			}
			g.cleanup()
			if s.String() != syscall.SIGTSTP.String() {
				os.Stdout.WriteString("\n")
				os.Exit(0)
			}
			g.pauseInput <- struct{}{}
			signal.Reset(syscall.SIGTSTP)
			syscall.Kill(os.Getpid(), syscall.SIGTSTP)
			signal.Notify(g.sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGTSTP)
			g.setTTY()
			g.printGround()
			g.printAllSnakes()
			g.printAllFood()
			g.moveTo(position{g.h - 1, g.w - 1})
			g.Unlock()
			g.pauseInput <- struct{}{}
		}
	}()
}

func printRules() {
	ind := "     "
	indB := ind + BLUE
	indM := ind + MAGENTA
	t := "Rules of juke"
	r1 := ind + "1. Snakes can go through the walls, they will appear on the other side."
	r2 := ind + "2. Snakes can go through themselves without death."
	r3 := ind + "3. If a snake's head's next movement means going through another snake it dies."
	d1 := fmt.Sprintf("\n%s===      ===      ===", indB)
	d1 += fmt.Sprintf("\n%s           =       x", indM)
	d1 += fmt.Sprintf("\n%s   =       =       x", indM)
	d1 += fmt.Sprintf("\n%s   =       =       x", indM)
	d1 += fmt.Sprintf("\n%s   =%s\n", indM, NORMAL)
	r4 := ind + "4. If two snake heads are going to collide into each other, they both die."
	d2 := fmt.Sprintf("\n%s =", indB)
	d2 += fmt.Sprintf("\n%s =        =        x", indB)
	d2 += fmt.Sprintf("\n%s =        =        x", indB)
	d2 += fmt.Sprintf("\n%s          =        x", indB)
	d2 += fmt.Sprintf("\n%s          =        x", indM)
	d2 += fmt.Sprintf("\n%s =        =        x", indM)
	d2 += fmt.Sprintf("\n%s =        =        x", indM)
	d2 += fmt.Sprintf("\n%s =%s\n", indM, NORMAL)
	r5 := ind + "5. If two snake heads are going to land onto the exact same square,\n" + ind + "one is randomly chosen to die and the other takes the square."
	d3 := fmt.Sprintf("\n%s =", indB)
	d3 += fmt.Sprintf("\n%s =        =", indB)
	d3 += fmt.Sprintf("\n%s =        =        =", indB)
	d3 += fmt.Sprintf("\n%s          =        =", indB)
	d3 += fmt.Sprintf("\n%s                   =", indB)
	d3 += fmt.Sprintf("\n%s          =        x", indM)
	d3 += fmt.Sprintf("\n%s =        =        x", indM)
	d3 += fmt.Sprintf("\n%s =        =        x", indM)
	d3 += fmt.Sprintf("\n%s =%s\n", indM, NORMAL)
	r6 := ind + "6. Food (the A) increases the length of the snake but not instantly,\n" + ind + "the way it grows is best understood by actually playing the game and eating some food."
	fmt.Printf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n", t, r1, r2, r3, d1, r4, d2, r5, d3, r6)
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
`)
	}
	tmpr := flag.Bool("rules", false, "print out rules and exit")
	tmph := flag.Uint("h", 0, "height of playground (default height of tty)")
	tmpw := flag.Uint("w", 0, "width of playground (default width of tty)")
	tmpi := flag.Uint("i", 3, "initital size of snake")
	tmpp := flag.Uint("p", 2, "number of players")
	tmps := flag.Uint("s", 30, "unit's per second for snake")
	tmpf := flag.Uint("f", 5, "how many blocks each food adds")

	flag.Parse()
	if *tmpr {
		printRules()
		os.Exit(0)
	}
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
	g.printAllSnakes()
	g.moveTo(position{g.h - 1, g.w - 1})
}

func (g *game) getSpeed() time.Duration {
	g.speedM.Lock()
	defer g.speedM.Unlock()
	return g.speed
}

func (g *game) loop() {
	for t := time.NewTimer(time.Second / g.getSpeed()); ; t.Reset(time.Second / g.getSpeed()) {
		select {
		case <-g.restart:
			g.next()
			continue
		case <-g.pauseLoop:
			g.Lock()
			g.pausedLoop = true
			g.Unlock()
			select {
			case <-g.pauseLoop:
			case <-g.restart:
				g.next()
			}
			g.Lock()
			g.pausedLoop = false
			g.Unlock()
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
	g.restart = make(chan struct{}, 1)
	g.pauseInput = make(chan struct{}, 1)
	g.s = make([]*snake, g.players)
	g.f = new(food)
	g.f.p = make([]position, g.players)
	g.f.g = g
	for i := uint16(0); i < g.players; i++ {
		g.s[i] = new(snake)
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
	g.printAllSnakes()
	g.printAllFood()
	go g.processInput()
	g.moveTo(position{g.h - 1, g.w - 1})
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
	var b byte
	r := bufio.NewReader(os.Stdin)
	read := func() {
		for {
			select {
			case <-g.pauseInput:
				<-g.pauseInput
			default:
				var err error
				if b, err = r.ReadByte(); err == nil {
					return
				}
			}
		}
	}
	changeDir := func(i uint16, d uint16) {
		if g.players <= i || g.s[i].getDead() == true {
			return
		}
		g.s[i].setDir(d)
	}
	for {
		read()
		switch b {
		case 27:
			for {
				read()
				switch b {
				// handle shift arrow keys
				case 91, 49, 59, 50:
					continue
				case 65, 67, 66, 68:
					changeDir(0, uint16(b-65))
				}
				break
			}
		case 'w', 'W':
			changeDir(1, up)
		case 's', 'S':
			changeDir(1, down)
		case 'a', 'A':
			changeDir(1, left)
		case 'd', 'D':
			changeDir(1, right)
		case 'y', 'Y':
			changeDir(2, up)
		case 'h', 'H':
			changeDir(2, down)
		case 'g', 'G':
			changeDir(2, left)
		case 'j', 'J':
			changeDir(2, right)
		case 'p', 'P':
			changeDir(3, up)
		case ';', ':':
			changeDir(3, down)
		case 'l', 'L':
			changeDir(3, left)
		case '\'', '|':
			changeDir(3, right)
		case '=':
			g.speedM.Lock()
			g.speed = time.Duration(float64(g.speed+1) * 1.1)
			g.speedM.Unlock()
		case '-':
			g.speedM.Lock()
			if float64(g.speed)/1.3 < 1 {
				g.speed = 1
			} else {
				g.speed = time.Duration(float64(g.speed) / 1.3)
			}
			g.speedM.Unlock()
		case 't', 'T':
			g.pauseLoop <- struct{}{}
		case 'r', 'R':
			select {
			case g.restart <- struct{}{}:
			default:
				// already restarting
			}
		case 'q', 'Q', 4:
			g.sigs <- syscall.SIGTERM
		}
	}
}

func (g *game) updateSnakes() {
	for i := uint16(0); i < g.players; i++ {
		g.s[i].update()
	}
}

func (g *game) printAllSnakes() {
	for i := uint16(0); i < g.players; i++ {
		if g.s[i].getDead() != true {
			g.s[i].printOverAll("=")
		} else {
			g.s[i].printOverAll("x")
		}
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
					(len(g.s[i].bs) == 1 && oppositeDir(g.s[i].getDir(), g.s[j].getDir()) &&
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
	d := getDimensions()
	if g.h == 0 {
		g.h = d[0]
	}
	if g.w == 0 {
		g.w = d[1]
	}
	if g.w < 5 {
		panic("width cannot be less than 5")
	} else if g.h < 4 {
		panic("height cannot be less than 4")
	}
	g.wf = float64(g.w - 2)
	if g.init > uint16(g.wf/3) {
		panic(fmt.Sprintf("max init size of snake for this width is %d", uint16(g.wf/3)))
	}
	g.hf = float64(g.h - 1)
	if i := uint16(g.origin.y) + g.h; i > d[0] {
		g.rowOffSet = i - d[0] - 1
	}
}

func (g *game) setTTY() {
	os.Stdin.WriteString(CURSORINVIS)
	raw := g.oldTios
	raw.Lflag &^= syscall.ECHO | syscall.ICANON
	writeTermios(raw)
	g.setOrigin()
	g.setDimensions()
}

func (g *game) cleanup() {
	// make text normal and cursor visible
	os.Stdin.WriteString(NORMAL + CURSORVIS)
	// set tty to normal
	writeTermios(g.oldTios)
}

func (g *game) setOrigin() {
	os.Stdin.WriteString(CURSORPOS)
	//TODO FIX THIS with regex
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

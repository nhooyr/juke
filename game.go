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
}

func (g *game) initialize() {
	g.printGround()
	g.s = make([]snake, g.players)
	for i := uint(0); i < g.players; i++ {
		g.s[i].g = g
		g.s[i].initialize(i + 1)
	}
	g.food = make([]position, g.players)
	for i, _ := range g.food {
		g.addFood(i)
	}
	go g.processInput()
}

func (g *game) getValidFoodPos() (vp []position) {
	vp = []position{}
	for i := uint16(1); i < g.h-1; i++ {
	xLoop:
		for j := uint16(1); j < g.w-1; j++ {
			for s := uint(0); s < g.players; s++ {
				for f := 0; uint(f) < g.players; f++ {
					if g.food[f] == (position{i, j}) || g.s[s].on(position{i, j}, 0) {
						continue xLoop
					}
				}
			}
			vp = append(vp, position{i, j})
		}
	}
	return
}

func (g *game) addFood(i int) {
	vp := g.getValidFoodPos()
	rand.Seed(time.Now().UnixNano())
	g.food[i] = vp[rand.Intn(len(vp))]
	g.moveTo(g.food[i])
	fmt.Print("A")
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
	changeDir := func(s uint, dir uint16) {
		if g.players <= s || g.s[s].dead == true || prevDir[s] == dir {
			return
		}
		g.s[s].Lock()
		g.s[s].bs[0].dir = dir
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
			case 10:
				return
			}
			// TODO proper restart
		}
	}
}

func (g *game) printSnakes() {
	for i := uint(0); i < g.players; i++ {
		if g.s[i].dead == false {
			printColor(i)
			g.s[i].print()
		}
	}
	os.Stdout.Write([]byte{27, 91, 48, 109})
}

func (g *game) moveSnakes() {
	for i, _ := range g.s {
		for j, _ := range g.food {
			if g.s[i].bs[0].pos == g.food[j] {
				g.s[i].appendBlocks(g.foodVal)
				g.addFood(j)
			}
		}
		if g.s[i].dead == false {
			g.s[i].move()
		}
	}
	for i, _ := range g.s {
		if g.s[i].dead == true {
			continue
		}
		for j, _ := range g.s {
			if j != i {
				if g.s[j].on(g.s[i].bs[0].pos, 0) {
					printColor(uint(i))
					g.s[i].die()
				}
				// TODO weird with one size snake
			}
		}
	}
	os.Stdout.Write([]byte{27, 91, 48, 109})
}

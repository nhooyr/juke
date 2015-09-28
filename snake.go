package main

import (
	"os"
	"sync"
)

const (
	up = iota + 1
	down
	right
	left
)

type position struct {
	y uint16
	x uint16
}

type block struct {
	d uint16
	p position
}

type snake struct {
	bs    []block
	oldBs []block
	g     *game
	dead  bool
	sync.Mutex
	player uint
}

func (s *snake) printColor() {
	switch s.player {
	case 0:
		os.Stdout.WriteString(BLUE)
	case 1:
		os.Stdout.WriteString(GREEN)
	case 2:
		os.Stdout.WriteString(RED)
	case 3:
		os.Stdout.WriteString(MAGENTA)
	}
}

func (s *snake) printOverAll(p string) {
	if p != " " {
		s.printColor()
	}
	for i, _ := range s.bs {
		s.g.moveTo(s.bs[i].p)
		os.Stdout.WriteString(p)
	}
}

func (s *snake) update() {
	if !s.on(s.oldBs[len(s.oldBs)-1].p, len(s.bs)-1, -1, -1) {
		var used bool
		for i := uint(0); i < s.g.players; i++ {
			if uint(i) == s.player {
				continue
			}
			if s.g.s[i].on(s.oldBs[len(s.oldBs)-1].p, 0, len(s.g.s[i].bs), 1) {
				used = true
				break
			}
		}
		if used == false {
			s.g.moveTo(s.oldBs[len(s.oldBs)-1].p)
			os.Stdout.WriteString(" ")
		}
	}
	if s.on(s.bs[0].p, 1, len(s.bs), 1) {
		return
	}
	s.printColor()
	s.g.moveTo(s.bs[0].p)
	os.Stdout.WriteString("=")
}
func (s *snake) die() {
	s.Lock()
	s.bs = s.oldBs
	s.Unlock()
	s.dead = true
	s.printOverAll("x")
}

func (s *snake) on(p position, min, end, inc int) bool {
	for i := min; i != end; i += inc {
		if s.bs[i].p == p {
			return true
		}
	}
	return false
}

func (s *snake) move() {
	s.Lock()
	s.oldBs = make([]block, len(s.bs))
	copy(s.oldBs, s.bs)
	for i := len(s.bs) - 1; i >= 0; i-- {
		switch s.bs[i].d {
		case up:
			s.bs[i].p.y -= 1
		case right:
			s.bs[i].p.x += 1
		case down:
			s.bs[i].p.y += 1
		case left:
			s.bs[i].p.x -= 1
		}
		if i != 0 {
			s.bs[i].d = s.bs[i-1].d
		}
		s.g.wallHax(&s.bs[i].p)
	}
	s.Unlock()
}

func (g *game) wallHax(p *position) {
	switch {
	case p.y == g.h-1:
		p.y = 1
	case p.y == 0:
		p.y = g.h - 2
	case p.x == g.w-1:
		p.x = 1
	case p.x == 0:
		p.x = g.w - 2
	}
}

func (s *snake) appendBlocks(i uint16) {
	for j := uint16(0); j < i; j++ {
		b := s.bs[len(s.bs)-1]
		b.d = s.bs[len(s.bs)-1].d
		switch b.d {
		case up:
			b.p.y += 1
		case right:
			b.p.x -= 1
		case down:
			b.p.y -= 1
		case left:
			b.p.x += 1
		}
		s.g.wallHax(&b.p)
		s.bs = append(s.bs, b)
	}
}

func (s *snake) initialize() {
	s.bs = make([]block, 1)
	s.bs[0].d = right
	switch s.player {
	case 0:
		s.bs[0].p.x, s.bs[0].p.y = s.g.w/3*2, s.g.h/3
	case 1:
		s.bs[0].p.x, s.bs[0].p.y = s.g.w/3, s.g.h/3
	case 2:
		s.bs[0].p.x, s.bs[0].p.y = s.g.w/3, s.g.h/3*2
	case 3:
		s.bs[0].p.x, s.bs[0].p.y = s.g.w/3*2, s.g.h/3*2
	}
	s.appendBlocks(s.g.init - 1)
}

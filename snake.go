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

func (b *block) moveBack() {
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

}

func (b *block) moveForward() {
	switch b.d {
	case up:
		b.p.y -= 1
	case right:
		b.p.x += 1
	case down:
		b.p.y += 1
	case left:
		b.p.x -= 1
	}
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
		if s.on(s.bs[i].p, i-1, -1, -1) {
			continue
		}
		s.g.moveTo(s.bs[i].p)
		os.Stdout.WriteString(p)
	}
}

func (s *snake) update() {
	var used bool
	for i := uint(0); i < s.g.players; i++ {
		if s.g.s[i].on(s.oldBs[len(s.oldBs)-1].p, 0, len(s.g.s[i].bs), 1) /*|| s.g.food[i] == s.oldBs[len(s.oldBs)-1].p */ {
			used = true
			break
		}
	}
	if used == false {
		s.g.moveTo(s.oldBs[len(s.oldBs)-1].p)
		os.Stdout.WriteString(" ")
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
	if len(s.bs) != len(s.oldBs) {
		for i := len(s.bs); i > len(s.oldBs)-1; i-- {
			s.oldBs = append(s.oldBs, block{})
		}
	}
	copy(s.oldBs, s.bs)
	for i := len(s.bs) - 1; i >= 0; i-- {
		s.bs[i].moveForward()
		if i != 0 {
			s.bs[i].d = s.bs[i-1].d
		}
		s.g.wallHax(&s.bs[i].p)
	}
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
		b.moveBack()
		s.g.wallHax(&b.p)
		s.bs = append(s.bs, b)
	}
}

func (s *snake) initialize() {
	s.bs = make([]block, s.g.init)
	s.oldBs = make([]block, s.g.init)
	s.bs[0].d = right
	switch s.player {
	case 0:
		s.bs[0].p.x, s.bs[0].p.y = s.g.w/3*2+s.g.w/3/2, s.g.h/3
	case 1:
		s.bs[0].p.x, s.bs[0].p.y = s.g.w/3, s.g.h/3
	case 2:
		s.bs[0].p.x, s.bs[0].p.y = s.g.w/3, s.g.h/3*2
	case 3:
		s.bs[0].p.x, s.bs[0].p.y = s.g.w/3*2+s.g.w/3/2, s.g.h/3*2
	}
	for i := uint16(1); i < s.g.init; i++ {
		s.bs[i] = s.bs[i-1]
		s.bs[i].moveBack()
		s.g.wallHax(&s.bs[i].p)
	}
}

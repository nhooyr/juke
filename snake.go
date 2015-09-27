package main

import (
	"fmt"
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

func (s *snake) print() {
	printColor(s.player)
	for i, _ := range s.bs {
		s.g.moveTo(s.bs[i].p)
		fmt.Print("=")
	}
}
func (s *snake) die() {
	printColor(s.player)
	s.dead = true
	for i, _ := range s.oldBs {
		s.g.moveTo(s.oldBs[i].p)
		fmt.Print("x")
	}
}

func (s *snake) on(p position) bool {
	for i := 0; i < len(s.bs); i++ {
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
	s.g.moveTo(s.bs[len(s.bs)-1].p)
	fmt.Print(" ")
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

func (s *snake) initialize(player uint) {
	s.bs = make([]block, 1)
	s.player = player
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

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
	dir uint16
	pos position
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
		s.g.moveTo(s.bs[i].pos)
		fmt.Print("=")
	}
}
func (s *snake) die() {
	printColor(s.player)
	s.dead = true
	for i, _ := range s.oldBs {
		s.g.moveTo(s.oldBs[i].pos)
		fmt.Print("x")
	}
}

func (s *snake) on(p position) bool {
	for i := 0; i < len(s.bs); i++ {
		if s.bs[i].pos == p {
			return true
		}
	}
	return false
}

func (s *snake) move() {
	s.Lock()
	s.oldBs = make([]block, len(s.bs))
	copy(s.oldBs, s.bs)
	s.g.moveTo(s.bs[len(s.bs)-1].pos)
	fmt.Print(" ")
	for i := len(s.bs) - 1; i >= 0; i-- {
		switch s.bs[i].dir {
		case up:
			s.bs[i].pos.y -= 1
		case right:
			s.bs[i].pos.x += 1
		case down:
			s.bs[i].pos.y += 1
		case left:
			s.bs[i].pos.x -= 1
		}
		if i != 0 {
			s.bs[i].dir = s.bs[i-1].dir
		}
		s.g.wallHax(&s.bs[i].pos)
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
		b.dir = s.bs[len(s.bs)-1].dir
		switch b.dir {
		case up:
			b.pos.y += 1
		case right:
			b.pos.x -= 1
		case down:
			b.pos.y -= 1
		case left:
			b.pos.x += 1
		}
		s.g.wallHax(&b.pos)
		s.bs = append(s.bs, b)
	}
}

func (s *snake) initialize(player uint) {
	s.bs = make([]block, 1)
	s.player = player
	s.bs[0].dir = right
	switch s.player {
	case 1:
		s.bs[0].pos.x, s.bs[0].pos.y = s.g.w/3*2, s.g.h/3
	case 2:
		s.bs[0].pos.x, s.bs[0].pos.y = s.g.w/3, s.g.h/3
	case 3:
		s.bs[0].pos.x, s.bs[0].pos.y = s.g.w/3, s.g.h/3*2
	case 4:
		s.bs[0].pos.x, s.bs[0].pos.y = s.g.w/3*2, s.g.h/3*2
	}
	s.appendBlocks(s.g.init - 1)
}

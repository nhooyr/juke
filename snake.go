package main

import "fmt"

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
	input chan uint16
	dead  bool
}

func (s *snake) print() {
	for i, _ := range s.bs {
		s.g.moveTo(s.bs[i].pos)
		fmt.Print("=")
	}
}
func (s *snake) die() {
	s.dead = true
	s.bs = s.oldBs
	for i, _ := range s.bs {
		s.g.moveTo(s.bs[i].pos)
		fmt.Print("x")
	}
}

func (s *snake) on(p position, start int) bool {
	for i := start; i < len(s.bs)-1; i++ {
		if s.bs[i].pos == p {
			return true
		}
	}
	return false
}

func (s *snake) move() {
	s.oldBs = make([]block, len(s.bs))
	copy(s.oldBs, s.bs)
	s.g.moveTo(s.bs[len(s.bs)-1].pos)
	fmt.Print(" ")
	select {
	case dir := <-s.input:
		s.bs[0].dir = dir
	default:
		//
	}
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
		switch {
		case s.bs[i].pos.y == s.g.h-1:
			s.bs[i].pos.y = 1
		case s.bs[i].pos.y == 0:
			s.bs[i].pos.y = s.g.h - 2
		case s.bs[i].pos.x == s.g.w-1:
			s.bs[i].pos.x = 1
		case s.bs[i].pos.x == 0:
			s.bs[i].pos.x = s.g.w - 2
		}
	}
}

func (s *snake) appendBlock(i uint16) {
	for j := uint16(0); j < i; j++ {
		b := s.bs[len(s.bs)-1]
		switch b.dir = s.bs[len(s.bs)-1].dir; b.dir {
		case up:
			b.pos.y += 1
		case right:
			b.pos.x -= 1
		case down:
			b.pos.y -= 1
		case left:
			b.pos.x += 1
		}
		s.bs = append(s.bs, b)
	}
}

func (s *snake) initialize(player uint) {
	s.input = make(chan uint16)
	s.bs = make([]block, 1)
	s.bs[0].dir = right
	switch player {
	case 1:
		s.bs[0].pos.x, s.bs[0].pos.y = s.g.w/3, s.g.h/3
	case 2:
		s.bs[0].pos.x, s.bs[0].pos.y = s.g.w/3*2, s.g.h/3
	case 3:
		s.bs[0].pos.x, s.bs[0].pos.y = s.g.w/3, s.g.h/3*2
	case 4:
		s.bs[0].pos.x, s.bs[0].pos.y = s.g.w/3*2, s.g.h/3*2
	}
	s.appendBlock(s.g.init-1)
}

package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"syscall"
	"time"
)

const (
	up = 1 << iota
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
	bs      []block
	g       *game
	input   chan uint16
	lastDir uint16
}

func (s *snake) print() {
	for i, _ := range s.bs {
		s.g.moveTo(s.bs[i].pos)
		fmt.Print("=")
	}
}

func (s *snake) isNotOn(p position) bool {
	for i, _ := range s.bs {
		if s.bs[i].pos == p {
			return false
		}
	}
	return true
}

func (s *snake) move() {
	s.g.moveTo(s.bs[len(s.bs)-1].pos)
	fmt.Print(" ")
	select {
	case dir := <-s.input:
		s.bs[0].dir = dir
	default:
		//
	}
	if s.bs[0].pos == s.g.food {
		s.appendBlock()
		s.g.addFood()
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

// process the input
func (s *snake) processInput() {
	b := make([]byte, 3)
	var prevDir uint16
	for {
		_, err := os.Stdin.Read(b)
		if err != nil {
			log.Print(err)
			s.g.sigs <- syscall.SIGTERM
		}
		if b[0] == 27 && b[1] == 91 {
			dir := uint16(1 << (b[2] - 65))
			if dir == prevDir {
				continue
			}
			switch dir {
			case up, down, right, left:
				s.input <- dir
				prevDir = dir
			}
		}
	}
}

func (s *snake) appendBlock() {
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

func (s *snake) initialize() {
	s.input = make(chan uint16)
	s.bs = make([]block, 1)
	rand.Seed(time.Now().UnixNano())
	s.bs[0].dir = 1 << uint16(rand.Int63n(3))
	s.bs[0].pos.x, s.bs[0].pos.y = s.g.w/2, s.g.h/2
	for i := s.g.init - 1; i > 0; i-- {
		s.appendBlock()
	}
}

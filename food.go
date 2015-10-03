package main

import (
	"math/rand"
	"os"
	"time"
)

type food struct {
	p  []position
	vp []position // valid positions
	x  uint16
	y  uint16
	s  uint16
	i  uint16
	g  *game
}

func (f *food) loopY() {
	for f.y = uint16(1); f.y < f.g.h-1; f.y++ {
		f.loopX()
	}
}

func (f *food) loopX() {
	for f.x = uint16(1); f.x < f.g.w-1; f.x++ {
		if f.loopS() {
			continue
		}
		f.vp = append(f.vp, position{f.y, f.x})
	}
}

func (f *food) loopS() bool {
	for f.s = uint16(0); f.s < f.g.players; f.s++ {
		if f.loopF() {
			return true
		}
	}
	return false
}

func (f *food) loopF() bool {
	for j := uint16(0); j < f.g.players; j++ {
		if (f.p[j] == (position{f.y, f.x}) || f.g.s[f.s].on(position{f.y, f.x}, 0, len(f.g.s[f.s].bs), 1)) {
			return true
		}
	}
	return false
}

func (f *food) fillVP(i uint16) {
	f.vp = make([]position, 0)
	f.i = i
	f.loopY()
}

func (f *food) addFood(i uint16) {
	f.fillVP(i)
	if len(f.vp) == 0 {
		return
	}
	rand.Seed(time.Now().UnixNano())
	f.p[i] = f.vp[rand.Intn(len(f.vp))]
}

func (f *food) printFood(i uint16) {
	if len(f.vp) == 0 {
		return
	}
	f.g.moveTo(f.p[i])
	os.Stdout.WriteString(NORMAL + "A")
}

func (f *food) clearFood(i uint16) {
	f.g.moveTo(f.p[i])
	os.Stdout.WriteString(" ")
}

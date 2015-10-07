package main

import (
	"math/rand"
	"os"
	"time"
)

type food struct {
	p  []position
	vp []position // valid positions
	i  uint16     // operating food
	s  uint16
	g  *game
}

// loop y/x then send those values to loopS
func (f *food) loopYX(cb func(x, y uint16) bool) {
	for y := uint16(1); y < f.g.h-1; y++ {
		for x := uint16(1); x < f.g.w-1; x++ {
			if cb(y, x) {
				continue
			}
			f.vp = append(f.vp, position{y, x})
		}
	}
}

// loop snake then food with x/y
func (f *food) loopSF(y, x uint16) bool {
	for f.s = uint16(0); f.s < f.g.players; f.s++ {
		for j := uint16(0); j < f.g.players; j++ {
			if ((f.i == 0 || f.i-1 != j) && f.p[j] == (position{y, x})) || f.g.s[f.s].on(position{y, x}, 0, len(f.g.s[f.s].bs), 1) {
				return true
			}
		}
	}
	return false
}

func (f *food) fillVP(i uint16) {
	f.i = i
	f.vp = make([]position, 0)
	f.loopYX(f.loopSF)
}

func (f *food) addFood(i uint16) {
	rand.Seed(time.Now().UnixNano())
	f.p[i] = f.vp[rand.Intn(len(f.vp))]
}

func (f *food) printFood(i uint16) {
	if f.p[i] == (position{}) {
		return
	}
	f.g.moveTo(f.p[i])
	os.Stdout.WriteString(NORMAL + "A")
}

func (f *food) clearFood(i uint16) {
	if f.p[i] == (position{}) {
		return
	}
	f.g.moveTo(f.p[i])
	os.Stdout.WriteString(" ")
}

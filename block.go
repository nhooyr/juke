package main

const (
	up = iota
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
		b.p.y++
	case right:
		b.p.x--
	case down:
		b.p.y++
	case left:
		b.p.x--
	}

}

func (b *block) moveForward() {
	switch b.d {
	case up:
		b.p.y--
	case right:
		b.p.x++
	case down:
		b.p.y++
	case left:
		b.p.x--
	}
}

func oppositeDir(d1, d2 uint16) bool {
	if d1%2 == 0 {
		if d1+1 == d2 {
			return true
		}
	} else {
		if d1-1 == d2 {
			return true
		}
	}
	return false
}

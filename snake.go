package main

const (
	up = 1 << iota
	right
	down
	left
)


type position struct {
	x uint16
	y uint16
}


type block struct {
	dir uint16
	pdir uint16
	pos position
}

type snake []block

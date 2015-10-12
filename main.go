package main

import "log"

func main() {
	log.SetPrefix(NORMAL + "juke: ")
	log.SetFlags(0)
	g := new(game)
	g.parseFlags()
	g.oldTios = readTermios()
	g.captureSignals()
	defer func() {
		log.Print(recover())
		g.cleanup()
	}()
	g.setTTY()
	g.start()
}

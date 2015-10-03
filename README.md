# juke
A terminal snake game in which the primary objective is to juke (basically dodge and attack other snakes) other players as well as get fod to make yourself more long.

Install with

	go get github.com/aubble/juke

As long as $GOPATH/bin is in your path just run it with

	juke

## RULES

You can go through the walls, you can go through yourself, but if you touch anyone else with your head you die. E.g if I hit my head with someone else's snake I die but they live, if both heads hit each other both people die. If they both don't hit but are going to land on the same square e.g

    =_=

(= represent the heads, and _ the square they are both going to land on)

Then randomly one is chosen to get the square and the other dies.

Check out all the settings with

	juke -help

**ITS MUCH MUCH MORE FUN WITH FRIENDS, ESPECIALLY IF YOU PLAY AROUND WITH ALL THE SETTINGS!**

<img src="https://raw.githubusercontent.com/nhooyr/juke/master/screenshot.png" border="0">

###TODO
net multiplayer

more powerups

documentation

proper interface

package main

import (
	//	"github.com/teq0/godown/webserver"
	"github.com/teq0/godown/db"
	"github.com/teq0/godown/download"
)

func main() {
	//	webserver.StartServer()

	db.Init()

	someURL := "http://s0.cyberciti.org/images/misc/static/2012/11/ifdata-welcome-0.png"
	dl := download.Download{}

	dl.Start(someURL, "/Users/craig/test1.png")
}

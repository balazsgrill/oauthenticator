package app

import (
	"os"
	"os/signal"
)

type Application interface {
	InitFlags()
	ParseFlags()
	Init()
	Start()
	Stop()
}

func Main(app Application) {
	app.InitFlags()
	app.ParseFlags()
	app.Init()

	go app.Start()

	// Setting up signal capturing
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	// Waiting for SIGINT (kill -2)
	<-stop
	app.Stop()
}

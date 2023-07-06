package app

type Application interface {
	InitFlags()
	ParseFlags()
	Init()
	Start()
}

func Main(app Application) {
	app.InitFlags()
	app.ParseFlags()
	app.Init()
	app.Start()
}

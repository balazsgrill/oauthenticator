package main

import (
	"github.com/balazsgrill/oauthenticator/app"
	"github.com/balazsgrill/oauthenticator/server"
)

func main() {
	app.Main((&server.MainApp{}))
}

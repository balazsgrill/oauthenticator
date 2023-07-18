package main

import (
	"github.com/balazsgrill/oauthenticator/app"
	cliapp "github.com/balazsgrill/oauthenticator/cliapp"
)

func main() {
	app.Main((&cliapp.MainApp{}))
}

package main

import "github.com/balazsgrill/oauthenticator/server"

func main() {
	main := &server.Main{}
	main.InitFlags()
	main.ParseFlags()
	main.Init()
	main.Start()
}

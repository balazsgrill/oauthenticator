package oauthenticator

import "golang.org/x/oauth2"

type TokenPersistence interface {
	oauth2.TokenSource

	SetToken(*oauth2.Token)
}

type Config interface {
	Identifier() string
	Label() string
	Config() *oauth2.Config
	Endpoint() oauth2.Endpoint
}

type Provider[C Config] interface {
	Configs() ([]C, error)
	Token(C) TokenPersistence
}

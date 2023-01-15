package oauthenticator

import (
	"github.com/knakk/rdf"
	"golang.org/x/oauth2"
)

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
	Config(item rdf.Term) (C, error)
	Token(C) TokenPersistence
	Options(C) []oauth2.AuthCodeOption
}

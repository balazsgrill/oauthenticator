package oauthenticator

import (
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
	Token() TokenPersistence
	Options() []oauth2.AuthCodeOption
}

type Provider interface {
	Configs() ([]Config, error)
	ConfigsOfType(ctype string) ([]Config, error)
	Config(identifier string) (Config, error)
}

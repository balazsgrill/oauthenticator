package client

import (
	"context"
	"net/http"

	"github.com/balazsgrill/oauthenticator"
	"golang.org/x/oauth2"
)

type Oauth2Client struct {
	client          *http.Client
	oauth           *oauth2.Config
	TokenPersitence oauthenticator.TokenPersistence
}

func New(oauth *oauth2.Config, tokenPersistence oauthenticator.TokenPersistence) *Oauth2Client {
	return &Oauth2Client{
		oauth:           oauth,
		TokenPersitence: tokenPersistence,
	}
}

func (ms *Oauth2Client) GetClient() *http.Client {
	if ms.client == nil {
		ms.client = oauth2.NewClient(context.Background(), ms.TokenPersitence)
	}
	return ms.client
}

func (ms *Oauth2Client) Get(request string) (resp *http.Response, err error) {
	client := ms.GetClient()
	return client.Get(request)
}

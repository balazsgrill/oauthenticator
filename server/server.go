package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/balazsgrill/oauthenticator"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type Server[C oauthenticator.Config] struct {
	provider      oauthenticator.Provider[C]
	authprocesses map[string]C
	params        func(C) []oauth2.AuthCodeOption
	favicon       FaviconService
}

func InitializeServer[C oauthenticator.Config](serveMux *http.ServeMux, provider oauthenticator.Provider[C], favicon FaviconService) {
	server := Server[C]{
		provider:      provider,
		authprocesses: make(map[string]C),
		params:        provider.Options,
		favicon:       favicon,
	}
	// handle route using handler function
	serveMux.HandleFunc("/verify", server.VerifyRequest)
	serveMux.HandleFunc("/auth", server.Authenticate)
	//http.HandleFunc("/proxy/", server.ApiReverseProxy)
	serveMux.HandleFunc("/", server.Index)
}

func (s *Server[C]) getConfigByID(id string) C {
	cs, err := s.provider.Configs()
	if err != nil {
		log.Print(err)
		var n C
		return n
	}
	for _, c := range cs {
		if id == c.Identifier() {
			return c
		}
	}
	var n C
	return n
}

func (s *Server[C]) Authenticate(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	id := query.Get("id")
	if id == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("ID is not provided")
	}
	c := s.getConfigByID(id)
	config := c.Config()
	state := uuid.NewString()
	s.authprocesses[state] = c
	http.Redirect(w, r, config.AuthCodeURL(state, s.params(c)...), http.StatusTemporaryRedirect)
}

func (s *Server[C]) VerifyRequest(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	var err error = nil

	if r.URL.Query().Has("error") {
		err = fmt.Errorf("%s: %s", r.URL.Query().Get("error"), r.URL.Query().Get("error_description"))
	}
	if state == "" {
		w.WriteHeader(http.StatusBadRequest)
		err = errors.New("state is not provided")
	}
	c, ok := s.authprocesses[state]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		err = errors.New("invalid state")
	}
	config := c.Config()

	// TODO Sparql client leaks HTTP Client settings
	http.DefaultClient = &http.Client{}

	if err == nil {
		token, err := config.Exchange(context.Background(), code, s.params(c)...)
		if err == nil {
			tokenpersistence := s.provider.Token(c)
			tokenpersistence.SetToken(token)
		}
	}
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	msg := "Auth successful"
	if err != nil {
		msg = err.Error()
	}

	fmt.Fprint(w, "<a href=\"/\">Return</a><br>")
	fmt.Fprintf(w, "<pre>%s</pre>", msg)
}

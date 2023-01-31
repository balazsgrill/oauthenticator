package server

import (
	"context"
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

func InitializeServer[C oauthenticator.Config](provider oauthenticator.Provider[C], favicon FaviconService, port int) {
	server := Server[C]{
		provider:      provider,
		authprocesses: make(map[string]C),
		params:        provider.Options,
		favicon:       favicon,
	}
	// handle route using handler function
	http.HandleFunc("/verify", server.VerifyRequest)
	http.HandleFunc("/auth", server.Authenticate)
	//http.HandleFunc("/proxy/", server.ApiReverseProxy)
	http.HandleFunc("/", server.Index)
	url := fmt.Sprintf("localhost:%d", port)
	log.Printf("Listening on %s\n", url)
	http.ListenAndServe(url, nil)
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

	msg := "Auth successful"
	if state == "" {
		w.WriteHeader(http.StatusBadRequest)
		msg = "state is not provided"
	}
	c, ok := s.authprocesses[state]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		msg = "invalid state"
	}
	config := c.Config()

	// TODO Sparql client leaks HTTP Client settings
	http.DefaultClient = &http.Client{}

	token, err := config.Exchange(context.Background(), code, s.params(c)...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		msg = err.Error()
	}
	tokenpersistence := s.provider.Token(c)

	tokenpersistence.SetToken(token)
	fmt.Fprint(w, "<a href=\"/\">Return</a><br>")
	fmt.Fprintf(w, "<pre>%s</pre>", msg)
}

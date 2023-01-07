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
}

func InitializeServer[C oauthenticator.Config](provider oauthenticator.Provider[C], port int) {
	server := Server[C]{
		provider:      provider,
		authprocesses: make(map[string]C),
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

func (s *Server[C]) Index(w http.ResponseWriter, r *http.Request) {
	cs, err := s.provider.Configs()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	fmt.Fprint(w, "<html><body>")
	for _, c := range cs {
		fmt.Fprintf(w, "<a href=\"/auth?id=%s\">%s</a><br>", c.Identifier(), c.Label())
	}
	fmt.Fprint(w, "</body></html>")
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
	http.Redirect(w, r, config.AuthCodeURL(state, oauth2.SetAuthURLParam("type", "web_server")), http.StatusTemporaryRedirect)
}

func (s *Server[C]) VerifyRequest(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	if state == "" {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("state is not provided")
	}
	c, ok := s.authprocesses[state]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Println("invalid state")
	}
	config := c.Config()

	token, err := config.Exchange(context.Background(), code, oauth2.SetAuthURLParam("type", "web_server"))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err)
	}
	tokenpersistence := s.provider.Token(c)

	tokenpersistence.SetToken(token)
	fmt.Fprintf(w, "Auth successful")
}
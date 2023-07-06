package server

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/balazsgrill/oauthenticator"
	"github.com/balazsgrill/oauthenticator/sparqlpersistence"
	"github.com/knakk/sparql"
)

type MainApp struct {
	Repourlstr string
	Port       int
	Faviconsrv string
	Repo       *sparql.Repo

	Provider oauthenticator.Provider[*sparqlpersistence.OAuthConfig]
}

func (m *MainApp) InitFlags() {
	flag.StringVar(&m.Repourlstr, "r", "", "URL of SPARQL repository. http://user:pass@host:port/path")
	flag.IntVar(&m.Port, "port", 8083, "Listening port (default 8083)")
	flag.StringVar(&m.Faviconsrv, "favicon", "", "Favicon service (currently only faviconkit is supported) e.g. https://something-subdomain.faviconkit.com")
}

func (m *MainApp) ParseFlags() {
	flag.Parse()

	if m.Repourlstr == "" {
		log.Fatal("Repository URL is not defined")
	}
}

func (m *MainApp) Init() {
	repourl, err := url.Parse(m.Repourlstr)
	if err != nil {
		log.Fatal(err)
	}

	var options []func(*sparql.Repo) error
	options = append(options, sparql.Timeout(time.Millisecond*1500))

	if repourl.User != nil {
		pw, ok := repourl.User.Password()
		if ok {
			options = append(options, sparql.BasicAuth(repourl.User.Username(), pw))
		}
		repourl.User = nil
	}

	m.Repo, err = sparql.NewRepo(repourl.String(), options...)

	if err != nil {
		log.Fatal(err)
	}

	faviconservice := InitFaviconService(m.Faviconsrv)
	if m.Faviconsrv != "" && faviconservice == nil {
		fmt.Printf("Favicon service not recognized: '%s'", m.Faviconsrv)
	}

	m.Provider = sparqlpersistence.NewSparql(m.Repo)
	InitializeServer(http.DefaultServeMux, m.Provider, faviconservice)
}

func (m *MainApp) Start() {
	url := fmt.Sprintf("localhost:%d", m.Port)
	log.Printf("Listening on %s\n", url)
	http.ListenAndServe(url, nil)
}

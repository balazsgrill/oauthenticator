package server

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/balazsgrill/oauthenticator"
	filepersistence "github.com/balazsgrill/oauthenticator/persistence/file"
	sparqlpersistence "github.com/balazsgrill/oauthenticator/persistence/sparql"
	"github.com/knakk/sparql"
)

type MainApp struct {
	Repourlstr   string
	Configdirstr string
	Port         int
	Faviconsrv   string
	Repo         *sparql.Repo

	Provider oauthenticator.Provider
	mux      *http.ServeMux
	server   *http.Server
}

func (m *MainApp) InitFlags() {
	flag.StringVar(&m.Repourlstr, "r", "", "URL of SPARQL repository. http://user:pass@host:port/path")
	flag.StringVar(&m.Configdirstr, "d", "", "Path of configuration directory. Either this or a sparql repo must be set")
	flag.IntVar(&m.Port, "port", 8083, "Listening port (default 8083)")
	flag.StringVar(&m.Faviconsrv, "favicon", "", "Favicon service (currently only faviconkit is supported) e.g. https://something-subdomain.faviconkit.com")
}

func (m *MainApp) ParseFlags() {
	flag.Parse()

	if m.Repourlstr == "" && m.Configdirstr == "" {
		log.Fatal("Either a SPARQL repository or a configuration directory must be specified!")
	}
	if m.Repourlstr != "" && m.Configdirstr != "" {
		log.Fatal("Both a SPARQL repository and a configuration directory specified. Remove one of them.")
	}
}

func (m *MainApp) initSparqlRepo() {
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

	m.Provider = sparqlpersistence.NewSparql(m.Repo)
}

func (m *MainApp) initFileRepo() {
	m.Provider = filepersistence.NewDirectory(m.Configdirstr, fmt.Sprintf("http://localhost:%d/verify", m.Port))
}

func (m *MainApp) Init() {
	if m.Repourlstr != "" {
		m.initSparqlRepo()
	}
	if m.Configdirstr != "" {
		m.initFileRepo()
	}

	faviconservice := InitFaviconService(m.Faviconsrv)
	if m.Faviconsrv != "" && faviconservice == nil {
		fmt.Printf("Favicon service not recognized: '%s'", m.Faviconsrv)
	}

	m.mux = http.NewServeMux()
	InitializeServer(m.mux, m.Provider, faviconservice)
}

func (m *MainApp) HttpServeMux() *http.ServeMux {
	return m.mux
}

func (m *MainApp) Stop() {
	m.server.Shutdown(context.Background())
}

func (m *MainApp) Start() {
	url := fmt.Sprintf("localhost:%d", m.Port)
	log.Printf("Listening on %s\n", url)
	m.server = &http.Server{Addr: url, Handler: m.HttpServeMux()}
	m.server.ListenAndServe()
}

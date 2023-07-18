package cliapp

import (
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"time"

	"github.com/balazsgrill/oauthenticator"
	"github.com/balazsgrill/oauthenticator/client"
	"github.com/balazsgrill/oauthenticator/sparqlpersistence"
	"github.com/knakk/rdf"
	"github.com/knakk/sparql"
)

type MainApp struct {
	Repourlstr string
	GetUrl     string
	ConfigIRI  string
	ConfgTerm  rdf.IRI
	Repo       *sparql.Repo

	Provider oauthenticator.Provider[*sparqlpersistence.OAuthConfig]
}

func (m *MainApp) InitFlags() {
	flag.StringVar(&m.Repourlstr, "r", "", "URL of SPARQL repository. http://user:pass@host:port/path")
	flag.StringVar(&m.GetUrl, "g", "", "URL to get")
	flag.StringVar(&m.ConfigIRI, "c", "", "OAUTH Config IRI to use")
}

func (m *MainApp) ParseFlags() {
	flag.Parse()

	if m.Repourlstr == "" {
		log.Fatal("Repository URL is not defined")
	}

	var err error
	m.ConfgTerm, err = rdf.NewIRI(m.ConfigIRI)
	if err != nil {
		log.Fatal(err)
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

	m.Provider = sparqlpersistence.NewSparql(m.Repo)
}

func (m *MainApp) Start() {
	c, err := m.Provider.Config(m.ConfgTerm)
	if err != nil {
		log.Fatal(err)
	}

	client := client.New(c.Config(), m.Provider.Token(c))
	resp, err := client.Get(m.GetUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	log.Print(string(data))
}

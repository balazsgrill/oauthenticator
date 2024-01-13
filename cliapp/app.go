package cliapp

import (
	"flag"
	"io/ioutil"
	"log"
	"net/url"
	"time"

	"github.com/balazsgrill/oauthenticator"
	"github.com/balazsgrill/oauthenticator/client"
	sparqlpersistence "github.com/balazsgrill/oauthenticator/persistence/sparql"
	"github.com/knakk/sparql"
)

type MainApp struct {
	Repourlstr string
	GetUrl     string
	ConfigIRI  string
	ConfgTerm  string
	Repo       *sparql.Repo

	Provider oauthenticator.Provider
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

	m.ConfgTerm = m.ConfigIRI
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

func (m *MainApp) Stop() {

}

func (m *MainApp) Start() {
	c, err := m.Provider.Config(m.ConfgTerm)
	if err != nil {
		log.Fatal(err)
	}

	client := client.New(c.Config(), c.Token())
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

package main

import (
	"flag"
	"log"
	"net/url"
	"time"

	"github.com/balazsgrill/oauthenticator/server"
	"github.com/balazsgrill/oauthenticator/sparqlpersistence"
	"github.com/knakk/sparql"
)

func main() {
	var repourlstr string
	flag.StringVar(&repourlstr, "r", "", "URL of SPARQL repository. http://user:pass@host:port/path")
	flag.Parse()
	if repourlstr == "" {
		log.Fatal("Repository URL is not defined")
	}

	repourl, err := url.Parse(repourlstr)
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

	repo, err := sparql.NewRepo(repourl.String(), options...)

	if err != nil {
		log.Fatal(err)
	}

	provider := sparqlpersistence.NewSparql(repo)
	server.InitializeServer(provider, 8083)
}

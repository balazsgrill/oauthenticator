package main

import (
	"flag"
	"fmt"
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

	var port int
	flag.IntVar(&port, "port", 8083, "Listening port (default 8083)")

	var faviconsrv string
	flag.StringVar(&faviconsrv, "favicon", "", "Favicon service (currently only faviconkit is supported) e.g. https://something-subdomain.faviconkit.com")

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

	faviconservice := server.InitFaviconService(faviconsrv)
	if faviconsrv != "" && faviconservice == nil {
		fmt.Printf("Favicon service not recognized: '%s'", faviconsrv)
	}

	provider := sparqlpersistence.NewSparql(repo)
	server.InitializeServer(provider, faviconservice, port)
}

package sparqlpersistence

import (
	"bytes"
	"encoding/json"
	"log"

	"github.com/balazsgrill/oauthenticator"
	"github.com/knakk/rdf"
	"github.com/knakk/sparql"
	"golang.org/x/oauth2"
)

const queries = `
# tag: clients
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX oauth: <https://oauth.net/2#>
PREFIX dc: <http://purl.org/dc/elements/1.1/>
SELECT ?clientid ?clientsecret ?redirecturl ?client ?authurl ?tokenurl ?identifier ?label
WHERE {
  GRAPH ?anygraph {
	?client rdf:type oauth:Client .
	?client oauth:clientID ?clientid .
	?client oauth:clientSecret ?clientsecret .
	?client oauth:redirectURL ?redirecturl .
	?client oauth:endpoint ?endpoint .
	?client dc:identifier ?identifier .
	?client rdfs:label ?label .
	?endpoint oauth:authurl ?authurl .
	?endpoint oauth:tokenurl ?tokenurl .
  }
}

# tag: client
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX oauth: <https://oauth.net/2#>
PREFIX dc: <http://purl.org/dc/elements/1.1/>
SELECT ?clientid ?clientsecret ?redirecturl ?authurl ?tokenurl ?identifier ?label
WHERE {
  GRAPH ?anygraph {
	{{.Client}} rdf:type oauth:Client .
	{{.Client}} oauth:clientID ?clientid .
	{{.Client}} oauth:clientSecret ?clientsecret .
	{{.Client}} oauth:redirectURL ?redirecturl .
	{{.Client}} oauth:endpoint ?endpoint .
	{{.Client}} dc:identifier ?identifier .
	{{.Client}} rdfs:label ?label .
	?endpoint oauth:authurl ?authurl .
	?endpoint oauth:tokenurl ?tokenurl .
  }
}

# tag: token
PREFIX oauth: <https://oauth.net/2#>
SELECT ?token
WHERE {
	GRAPH {{.Graph}} {
		{{.Client}} oauth:token ?token
	}
}

# tag: updatetoken
PREFIX oauth: <https://oauth.net/2#>
WITH {{.Graph}}
DELETE {
	{{.Client}} oauth:token ?oldtoken
}
INSERT {
	{{.Client}} oauth:token {{.Token}}
}
WHERE {
	OPTIONAL { {{.Client}} oauth:token ?oldtoken }
}

# tag: options
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX oauth: <https://oauth.net/2#>
SELECT ?option ?value
WHERE {
	GRAPH ?anygraph {
		{{.Client}} oauth:param ?param .
		?param rdfs:label ?option .
		?param rdf:value ?value .
	}
}

# tag: clientsOfType
PREFIX rdf: <http://www.w3.org/1999/02/22-rdf-syntax-ns#>
PREFIX rdfs: <http://www.w3.org/2000/01/rdf-schema#>
PREFIX oauth: <https://oauth.net/2#>
PREFIX dc: <http://purl.org/dc/elements/1.1/>
SELECT ?item ?token
WHERE {
  GRAPH ?anygraph {
	?item rdf:type <{{.ClientType}}> .
  }
  GRAPH ?tokengraph {
	?item oauth:token ?token .
  }
}
`

type Queries struct {
	bank sparql.Bank
}

type sparqlProvider struct {
	repo    *sparql.Repo
	queries *Queries
}

type tokenInRepo struct {
	provider *sparqlProvider
	client   rdf.Term
}

type OAuthConfig struct {
	client       rdf.Term
	identifier   string
	label        string
	clientID     string
	clientSecret string
	redirectURL  string
	authurl      string
	tokenurl     string
}

func (c *OAuthConfig) Term() rdf.Term {
	return c.client
}

func (c *OAuthConfig) Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		Endpoint:     c.Endpoint(),
		RedirectURL:  c.redirectURL,
		Scopes:       []string{},
	}
}

func (c *OAuthConfig) Identifier() string {
	return c.identifier
}

func (c *OAuthConfig) Label() string {
	return c.label
}

func (c *OAuthConfig) Endpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  c.authurl,
		TokenURL: c.tokenurl,
	}
}

func InitializeQueries() *Queries {
	result := &Queries{}
	data := bytes.NewBufferString(queries)
	result.bank = sparql.LoadBank(data)
	return result
}

func NewSparql(repo *sparql.Repo) oauthenticator.Provider[*OAuthConfig] {
	return &sparqlProvider{
		repo:    repo,
		queries: InitializeQueries(),
	}
}

func (p *sparqlProvider) Options(c *OAuthConfig) []oauth2.AuthCodeOption {
	params, err := p.queries.GetParams(p.repo, c.client)
	if err != nil {
		log.Println(err)
	}
	return params
}

func (p *sparqlProvider) Config(term rdf.Term) (*OAuthConfig, error) {
	return p.queries.GetConfig(p.repo, term)
}

func (p *sparqlProvider) Configs() ([]*OAuthConfig, error) {
	return p.queries.ReadConfigs(p.repo)
}

func (p *sparqlProvider) ConfigsOfType(ctype string) ([]*OAuthConfig, error) {
	return p.queries.GetClientsOfType(p.repo, ctype)
}

func (p *sparqlProvider) Token(c *OAuthConfig) oauthenticator.TokenPersistence {
	return &tokenInRepo{
		provider: p,
		client:   c.client,
	}
}

func (tp *tokenInRepo) Token() (*oauth2.Token, error) {
	return tp.provider.queries.ReadToken(tp.provider.repo, tp.client)
}

func (tp *tokenInRepo) SetToken(t *oauth2.Token) {
	err := tp.provider.queries.WriteToken(tp.provider.repo, tp.client, t)
	if err != nil {
		log.Println(err)
	}
}

func (q *Queries) GetParams(repo *sparql.Repo, client rdf.Term) ([]oauth2.AuthCodeOption, error) {
	query, err := q.bank.Prepare("options", struct {
		Client string
	}{
		Client: client.Serialize(rdf.Turtle),
	})
	if err != nil {
		return nil, err
	}

	result, err := repo.Query(query)
	if err != nil {
		return nil, err
	}
	solutions := result.Solutions()
	if len(solutions) == 0 {
		return nil, nil
	}

	var params []oauth2.AuthCodeOption
	for i := 0; i < len(solutions); i++ {
		solution := solutions[i]
		option := solution["option"].String()
		value := solution["value"].String()
		params = append(params, oauth2.SetAuthURLParam(option, value))
	}

	return params, nil
}

func (q *Queries) WriteToken(repo *sparql.Repo, client rdf.Term, t *oauth2.Token) error {
	tokendata, err := json.Marshal(t)
	if err != nil {
		return err
	}
	tokenlit, err := rdf.NewLiteral(string(tokendata))
	if err != nil {
		return err
	}

	query, err := q.bank.Prepare("updatetoken", struct {
		Graph  string
		Client string
		Token  string
	}{
		Graph:  "<tokens>",
		Client: client.Serialize(rdf.Turtle),
		Token:  tokenlit.Serialize(rdf.Turtle),
	})
	if err != nil {
		return err
	}

	return repo.Update(query)
}

func (q *Queries) ReadToken(repo *sparql.Repo, client rdf.Term) (*oauth2.Token, error) {
	query, err := q.bank.Prepare("token", struct {
		Graph  string
		Client string
	}{
		Graph:  "<tokens>",
		Client: client.Serialize(rdf.Turtle),
	})
	if err != nil {
		return nil, err
	}

	result, err := repo.Query(query)
	if err != nil {
		return nil, err
	}
	solutions := result.Solutions()
	if len(solutions) == 0 {
		return nil, nil
	}

	solution := solutions[0]
	data := solution["token"].String()
	t := &oauth2.Token{}
	err = json.Unmarshal([]byte(data), t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (q *Queries) GetConfig(repo *sparql.Repo, client rdf.Term) (*OAuthConfig, error) {
	query, err := q.bank.Prepare("client", struct{ Client string }{
		Client: client.Serialize(rdf.Turtle),
	})
	if err != nil {
		return nil, err
	}

	res, err := repo.Query(query)
	if err != nil {
		return nil, err
	}

	solutions := res.Solutions()

	for _, solution := range solutions {
		return &OAuthConfig{
			client:       client,
			clientID:     solution["clientid"].String(),
			clientSecret: solution["clientsecret"].String(),
			redirectURL:  solution["redirecturl"].String(),
			authurl:      solution["authurl"].String(),
			tokenurl:     solution["tokenurl"].String(),
			identifier:   solution["identifier"].String(),
			label:        solution["label"].String(),
		}, nil
	}

	return nil, nil
}

func (q *Queries) ReadConfigs(repo *sparql.Repo) ([]*OAuthConfig, error) {
	query, err := q.bank.Prepare("clients")
	if err != nil {
		return nil, err
	}

	res, err := repo.Query(query)
	if err != nil {
		return nil, err
	}

	solutions := res.Solutions()
	result := make([]*OAuthConfig, len(solutions))

	for i := 0; i < len(solutions); i++ {
		solution := solutions[i]
		result[i] = &OAuthConfig{
			client:       solution["client"],
			clientID:     solution["clientid"].String(),
			clientSecret: solution["clientsecret"].String(),
			redirectURL:  solution["redirecturl"].String(),
			authurl:      solution["authurl"].String(),
			tokenurl:     solution["tokenurl"].String(),
			identifier:   solution["identifier"].String(),
			label:        solution["label"].String(),
		}
	}

	return result, nil
}

func (q *Queries) GetClientsOfType(repo *sparql.Repo, clientType string) ([]*OAuthConfig, error) {
	provider := NewSparql(repo)
	query, err := q.bank.Prepare("clientsOfType", struct{ ClientType string }{ClientType: clientType})
	if err != nil {
		return nil, err
	}

	result, err := repo.Query(query)
	if err != nil {
		return nil, err
	}

	solutions := result.Solutions()

	var res []*OAuthConfig
	for _, solution := range solutions {
		c, err := provider.Config(solution["item"])
		if err != nil {
			log.Println(err)
		} else {
			res = append(res, c)
		}
	}

	return res, nil
}

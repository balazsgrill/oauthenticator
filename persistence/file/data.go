package file

import (
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/balazsgrill/oauthenticator"
	"golang.org/x/oauth2"
)

type Configdata struct {
	Label_       string            `json:"label"`
	ClientID     string            `json:"clientid"`
	ClientSecret string            `json:"clientsecret"`
	AuthURL      string            `json:"authurl"`
	TokenURL     string            `json:"tokenurl"`
	Params       map[string]string `json:"params"`
}

type config struct {
	Configdata
	provider *directoryProvider
	path     string
}

func (c *config) load() error {
	return c.Configdata.Load(c.path)
}

func (c *Configdata) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, c)
}

func (c *config) Identifier() string {
	return c.path
}

func (c *Configdata) Label() string {
	return c.Label_
}

func (c *config) Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		Endpoint:     c.Endpoint(),
		RedirectURL:  c.provider.redirecturl,
		Scopes:       []string{},
	}
}

func (c *config) Token() oauthenticator.TokenPersistence {
	return c.provider.Token(c)
}
func (c *config) Options() []oauth2.AuthCodeOption {
	return c.provider.Options(c)
}

func (c *Configdata) Endpoint() oauth2.Endpoint {
	return oauth2.Endpoint{
		AuthURL:  c.AuthURL,
		TokenURL: c.TokenURL,
	}
}

var _ oauthenticator.Config = &config{}

type directoryProvider struct {
	path        string
	redirecturl string
}

func NewDirectory(path string, redirectURL string) oauthenticator.Provider {
	return &directoryProvider{
		path:        path,
		redirecturl: redirectURL,
	}
}

func (p *directoryProvider) Configs() ([]oauthenticator.Config, error) {
	entries, err := os.ReadDir(p.path)
	if err != nil {
		return nil, err
	}
	var result []oauthenticator.Config
	for _, entry := range entries {
		if strings.HasPrefix(strings.ToLower(entry.Name()), ".json") {
			c, err := p.Config(p.path + "/" + entry.Name())
			if err != nil {
				log.Println(err)
			}
			result = append(result, c)

		}
	}
	return result, nil
}

func (p *directoryProvider) ConfigsOfType(ctype string) ([]oauthenticator.Config, error) {
	return nil, nil
}

func (p *directoryProvider) Config(identifier string) (oauthenticator.Config, error) {
	c := &config{
		provider: p,
		path:     identifier,
	}
	err := c.load()
	if err != nil {
		return nil, err
	}
	return c, nil
}

type tokenfile struct {
	path string
}

func (tp *tokenfile) SetToken(t *oauth2.Token) {
	data, err := json.Marshal(t)
	if err != nil {
		data = make([]byte, 0)
	}
	os.WriteFile(tp.path, data, 0666)
}

func (tp *tokenfile) Token() (*oauth2.Token, error) {
	data, err := os.ReadFile(tp.path)
	if err != nil || len(data) == 0 {
		// possibly not existing file
		return nil, nil
	}
	t := &oauth2.Token{}
	return t, json.Unmarshal(data, t)
}

func (p *directoryProvider) Token(c *config) oauthenticator.TokenPersistence {
	return &tokenfile{
		path: c.path + ".token",
	}
}

func (p *directoryProvider) Options(c *config) []oauth2.AuthCodeOption {
	var result []oauth2.AuthCodeOption
	for key, value := range c.Configdata.Params {
		result = append(result, oauth2.SetAuthURLParam(key, value))
	}
	return result
}

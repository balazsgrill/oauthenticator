package server

import (
	"fmt"
	"net/http"
	"time"
)

const header = `
<html>
<head>
<link rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css">
</head>
<body class="w3-container"><ul class="w3-ul w3-card-4 w3-margin" style="max-width:40em">
`

func (s *Server[C]) Index(w http.ResponseWriter, r *http.Request) {
	cs, err := s.provider.Configs()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	fmt.Fprint(w, header)
	for _, c := range cs {
		var class string
		token, err := s.provider.Token(c).Token()
		if err != nil {
			class = "w3-red"
		} else if token == nil {
			class = "w3-white"
		} else if token.AccessToken == "" {
			class = "w3-red"
		} else if token.Expiry.IsZero() || time.Now().Before(token.Expiry) {
			class = "w3-green"
		} else {
			class = "w3-yellow"
		}

		fmt.Fprintf(w, "<a href=\"/auth?id=%s\">", c.Identifier())
		fmt.Fprintf(w, "<li class=\"w3-border %s\">", class)
		fmt.Fprintf(w, "<p>")
		if s.favicon != nil {
			imgsrc := s.favicon.FaviconSrc(c.Endpoint().TokenURL)
			fmt.Fprintf(w, "<img src=\"%s\" style=\"width:3em;height:3em;\">", imgsrc)
		}
		fmt.Fprintf(w, "%s</p></li>", c.Label())
		fmt.Fprintf(w, "</a>")
	}
	fmt.Fprint(w, "</body></html>")
}

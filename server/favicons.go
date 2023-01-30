package server

import (
	"fmt"
	"net/url"
	"strings"
)

type FaviconService interface {
	FaviconSrc(endpoint string) string
}

type faviconKit struct {
	base string
}

func (fk *faviconKit) FaviconSrc(endpoint string) string {
	u, err := url.Parse(endpoint)
	if err == nil {
		return fmt.Sprintf("%s/%s/%d", fk.base, u.Hostname(), 24)
	}
	return ""
}

func InitFaviconService(serviceurl string) FaviconService {
	if strings.Contains(serviceurl, ".faviconkit.com") {
		return &faviconKit{
			base: serviceurl,
		}
	}
	return nil
}

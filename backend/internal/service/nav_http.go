package service

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func newNAVHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		CheckRedirect: func(request *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			if err := validateTNURL(request.URL); err != nil {
				return fmt.Errorf("redirect outside tn.ru is not allowed: %w", err)
			}
			return nil
		},
	}
}

func validateTNURL(address *url.URL) error {
	host := strings.ToLower(address.Hostname())
	if address.Scheme != "https" || (host != "tn.ru" && !strings.HasSuffix(host, ".tn.ru")) {
		return fmt.Errorf("only HTTPS URLs on tn.ru are allowed")
	}
	return nil
}

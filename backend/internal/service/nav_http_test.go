package service

import (
	"net/http"
	"net/url"
	"testing"
)

func TestNAVHTTPClientRejectsExternalRedirect(t *testing.T) {
	client := newNAVHTTPClient(0)
	request := &http.Request{URL: &url.URL{Scheme: "https", Host: "example.com"}}

	if err := client.CheckRedirect(request, []*http.Request{{}}); err == nil {
		t.Fatal("expected external redirect to be rejected")
	}
}

func TestNAVHTTPClientAllowsTNRedirect(t *testing.T) {
	client := newNAVHTTPClient(0)
	request := &http.Request{URL: &url.URL{Scheme: "https", Host: "nav.tn.ru"}}

	if err := client.CheckRedirect(request, []*http.Request{{}}); err != nil {
		t.Fatalf("expected tn.ru redirect to be accepted: %v", err)
	}
}

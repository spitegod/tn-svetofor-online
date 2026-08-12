package service

import (
	"context"
	"os"
	"testing"
)

func TestNAVLiveCatalogLoadsWithoutBrowserFallback(t *testing.T) {
	if os.Getenv("NAV_LIVE_TEST") == "" {
		t.Skip("NAV_LIVE_TEST is not set")
	}
	parser := NewNavParserService(&navParserRepositoryStub{})
	t.Cleanup(parser.Close)

	document, err := parser.fetchDocument(context.Background(), navBaseURL+"/systems/")
	if err != nil {
		t.Fatalf("load NAV catalog: %v", err)
	}
	if categories := collectCategoryURLs(document); len(categories) == 0 {
		t.Fatal("NAV catalog contains no system categories")
	}
}

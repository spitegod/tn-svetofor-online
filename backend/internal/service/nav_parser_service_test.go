package service

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestCollectCategoryURLsIncludesImage(t *testing.T) {
	document, err := html.Parse(strings.NewReader(`
		<a href="/systems/ploskaya-krysha/" class="b-system-page__item">
			<div class="b-system-page__img">
				<img src="https://nav.tn.ru/cloud/roof.webp" alt="">
			</div>
		</a>`))
	if err != nil {
		t.Fatal(err)
	}

	categories := collectCategoryURLs(document)
	if len(categories) != 1 {
		t.Fatalf("expected one category, got %d", len(categories))
	}
	if categories[0].ImageURL != "https://nav.tn.ru/cloud/roof.webp" {
		t.Fatalf("unexpected image URL %q", categories[0].ImageURL)
	}
}

func TestResolveNavAssetURLRejectsExternalHost(t *testing.T) {
	if resolved := resolveNavAssetURL("https://example.com/image.webp"); resolved != "" {
		t.Fatalf("expected external image URL to be rejected, got %q", resolved)
	}
}

func TestParserSystemPercent(t *testing.T) {
	tests := []struct {
		processed int
		total     int
		want      int
	}{
		{processed: 0, total: 100, want: 30},
		{processed: 50, total: 100, want: 64},
		{processed: 100, total: 100, want: 98},
		{processed: 10, total: 0, want: 30},
	}
	for _, test := range tests {
		if got := parserSystemPercent(test.processed, test.total); got != test.want {
			t.Errorf("parserSystemPercent(%d, %d) = %d, want %d", test.processed, test.total, got, test.want)
		}
	}
}

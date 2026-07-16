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

package service

import (
	"strings"
	"testing"
)

func TestSystemURLFromSearchSelectsMatchingSystemResult(t *testing.T) {
	page := `
		<nav><a href="/systems/ploskaya-krysha/pgs/">Плоская крыша</a></nav>
		<a href="/systems/stilobaty/other-system/" class="b-search-teaser__title">Другая система</a>
		<a href="/systems/stilobaty/tn-stilobat-klassik-avto/?sphrase_id=123" class="b-search-teaser__title">
			<b>ТН-СТИЛОБАТ КЛАССИК АВТО</b>
		</a>`

	got := systemURLFromSearch(strings.NewReader(page), "ТН-СТИЛОБАТ КЛАССИК АВТО")
	want := "https://nav.tn.ru/systems/stilobaty/tn-stilobat-klassik-avto/"
	if got != want {
		t.Fatalf("systemURLFromSearch() = %q, want %q", got, want)
	}
}

func TestSystemURLFromSearchReturnsEmptyWithoutExactMatch(t *testing.T) {
	page := `<a href="/systems/stilobaty/tn-stilobat-klassik/" class="b-search-teaser__title">ТН-СТИЛОБАТ КЛАССИК</a>`

	if got := systemURLFromSearch(strings.NewReader(page), "ТН-СТИЛОБАТ КЛАССИК АВТО"); got != "" {
		t.Fatalf("systemURLFromSearch() = %q, want empty URL", got)
	}
}

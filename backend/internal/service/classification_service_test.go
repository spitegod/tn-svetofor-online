package service

import (
	"strings"
	"testing"
)

func TestSystemURLFromSearchSelectsMatchingSystemResult(t *testing.T) {
	page := `
		<nav><a href="/systems/ploskaya-krysha/pgs/">Плоская крыша</a></nav>
		<a href="/systems/stilobaty/other-system/" class="b-search-teaser__title">Другая система</a>
		<div class="b-search-teaser">
			<a href="/systems/stilobaty/tn-stilobat-klassik-avto/?sphrase_id=123" class="b-search-teaser__title">
				<b>ТН-СТИЛОБАТ   КЛАССИК АВТО</b>
			</a>
			<div class="b-search-teaser__constr_segment">ПГС</div>
		</div>`

	got, constructionType := systemDataFromSearch(strings.NewReader(page), "  тн-стилобат классик авто  ")
	want := "https://nav.tn.ru/systems/stilobaty/tn-stilobat-klassik-avto/"
	if got != want {
		t.Fatalf("systemDataFromSearch() URL = %q, want %q", got, want)
	}
	if constructionType != "Промышленное и гражданское строительство" {
		t.Fatalf("systemDataFromSearch() construction type = %q", constructionType)
	}
}

func TestSystemURLFromSearchReturnsEmptyWithoutExactMatch(t *testing.T) {
	page := `<a href="/systems/stilobaty/tn-stilobat-klassik/" class="b-search-teaser__title">ТН-СТИЛОБАТ КЛАССИК</a>`

	got, constructionType := systemDataFromSearch(strings.NewReader(page), "ТН-СТИЛОБАТ КЛАССИК АВТО")
	if got != "" {
		t.Fatalf("systemDataFromSearch() URL = %q, want empty URL", got)
	}
	if constructionType != unassignedConstructionType {
		t.Fatalf("systemDataFromSearch() construction type = %q, want %q", constructionType, unassignedConstructionType)
	}
}

func TestNormalizeConstructionType(t *testing.T) {
	tests := map[string]string{
		"ИЖС": "Индивидуальное жилищное строительство",
		"транспортное строительство": "Транспортное и дорожное строительство",
		"Специальные сооружения":     "Специальные сооружения",
		"неизвестный сегмент":        unassignedConstructionType,
	}
	for value, want := range tests {
		if got := normalizeConstructionType(value); got != want {
			t.Errorf("normalizeConstructionType(%q) = %q, want %q", value, got, want)
		}
	}
}

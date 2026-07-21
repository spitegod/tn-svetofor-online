package repository

import "testing"

func TestContainsLikePatternEscapesWildcards(t *testing.T) {
	if got, want := containsLikePattern(`A%_\B`), `%a\%\_\\b%`; got != want {
		t.Fatalf("containsLikePattern() = %q, want %q", got, want)
	}
}

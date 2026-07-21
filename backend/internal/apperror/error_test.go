package apperror

import (
	"fmt"
	"testing"
)

func TestKindSurvivesWrapping(t *testing.T) {
	err := fmt.Errorf("outer: %w", New(Validation, "invalid input"))

	kind, ok := KindOf(err)
	if !ok || kind != Validation {
		t.Fatalf("KindOf() = %v, %v", kind, ok)
	}
}

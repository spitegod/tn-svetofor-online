package config

import "testing"

func TestLoadRequiresDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	if _, err := Load(); err == nil {
		t.Fatal("expected missing DATABASE_URL to be rejected")
	}
}

func TestLoadReadsConfiguration(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("HTTP_ADDR", ":9090")

	configuration, err := Load()
	if err != nil {
		t.Fatalf("load configuration: %v", err)
	}
	if configuration.DatabaseURL != "postgres://example" || configuration.HTTPAddr != ":9090" {
		t.Fatalf("unexpected configuration: %#v", configuration)
	}
}

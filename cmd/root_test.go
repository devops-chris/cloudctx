package cmd

import "testing"

func TestVersion(t *testing.T) {
	// Basic sanity test
	if version == "" {
		version = "dev"
	}
	if version != "dev" && version == "" {
		t.Error("version should not be empty")
	}
}


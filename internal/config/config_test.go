package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFromEnvironmentParsesNamedRoots(t *testing.T) {
	first := t.TempDir()
	second := t.TempDir()
	t.Setenv("DISK_INSIGHT_ROOTS", first+"::Primary;"+second+"::Archive")
	t.Setenv("DISK_INSIGHT_DATABASE", filepath.Join(t.TempDir(), "test.db"))
	config, err := FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}
	if len(config.Roots) != 2 || config.Roots[0].Name != "Primary" || config.Roots[1].Name != "Archive" {
		t.Fatalf("unexpected roots: %#v", config.Roots)
	}
	if config.Address != ":8080" {
		t.Fatalf("address = %q", config.Address)
	}
}

func TestFromEnvironmentRejectsDuplicates(t *testing.T) {
	root := t.TempDir()
	t.Setenv("DISK_INSIGHT_ROOTS", root+"::One;"+root+"::Two")
	if _, err := FromEnvironment(); err == nil {
		t.Fatal("expected duplicate root error")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

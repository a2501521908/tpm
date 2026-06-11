package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPathTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("home dir: %v", err)
	}

	got, err := ExpandPath("~/tpm-data")
	if err != nil {
		t.Fatalf("expand: %v", err)
	}
	want := filepath.Join(home, "tpm-data")
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestExpandPathEmpty(t *testing.T) {
	if _, err := ExpandPath("  "); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("APPDATA", tmp) // cover windows branch too

	in := &Config{Local: Local{DataDir: filepath.Join(tmp, "store")}}
	if err := in.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	out, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if out.Local.DataDir != in.Local.DataDir {
		t.Fatalf("got %q want %q", out.Local.DataDir, in.Local.DataDir)
	}
}

func TestLoadNotInitialized(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmp)
	t.Setenv("APPDATA", tmp)

	if _, err := Load(); err != ErrNotInitialized {
		t.Fatalf("expected ErrNotInitialized, got %v", err)
	}
}

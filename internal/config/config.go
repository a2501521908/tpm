// Package config handles the per-machine local configuration (env.toml). It only
// stores machine-specific values such as the data directory, which typically
// points into a cloud sync folder (iCloud, OneDrive, Google Drive). The
// encrypted entries themselves live under DataDir, never here.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/BurntSushi/toml"
)

// ErrNotInitialized is returned when env.toml does not exist yet.
var ErrNotInitialized = errors.New("config: not initialized, run `tpm init` first")

// Config is the local machine configuration persisted to env.toml.
type Config struct {
	Local Local `toml:"local"`
}

// Local holds machine-specific settings.
type Local struct {
	// DataDir is the (already expanded) absolute path to the encrypted store,
	// usually inside a cloud sync folder.
	DataDir string `toml:"data_dir"`
}

// Dir returns the directory that holds env.toml for the current OS:
//   - Windows: %APPDATA%\tpm
//   - macOS/Linux: $XDG_CONFIG_HOME/tpm or ~/.config/tpm
func Dir() (string, error) {
	if runtime.GOOS == "windows" {
		if appData := os.Getenv("APPDATA"); appData != "" {
			return filepath.Join(appData, "tpm"), nil
		}
	}
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "tpm"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("config: resolve home dir: %w", err)
	}
	return filepath.Join(home, ".config", "tpm"), nil
}

// Path returns the full path to env.toml.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "env.toml"), nil
}

// Load reads env.toml. It returns ErrNotInitialized when the file is missing.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotInitialized
		}
		return nil, fmt.Errorf("config: read %s: %w", path, err)
	}
	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %s: %w", path, err)
	}
	if cfg.Local.DataDir == "" {
		return nil, fmt.Errorf("config: data_dir is empty in %s", path)
	}
	return &cfg, nil
}

// Save writes env.toml, creating the config directory if needed.
func (c *Config) Save() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("config: create dir %s: %w", dir, err)
	}
	path := filepath.Join(dir, "env.toml")
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("config: open %s: %w", path, err)
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(c); err != nil {
		return fmt.Errorf("config: encode %s: %w", path, err)
	}
	return nil
}

// ExpandPath expands a leading "~" to the user's home directory and returns an
// absolute, cross-platform path.
func ExpandPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", errors.New("config: empty path")
	}
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("config: resolve home dir: %w", err)
		}
		if p == "~" {
			p = home
		} else {
			p = filepath.Join(home, p[2:])
		}
	}
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", fmt.Errorf("config: absolute path: %w", err)
	}
	return abs, nil
}

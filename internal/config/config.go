package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const DefaultBaseURL = "https://ttrack.top"

type Config struct {
	Token   string `toml:"token"`
	BaseURL string `toml:"base_url"`
}

// Path returns the config file location, honouring XDG_CONFIG_HOME.
func Path() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	return filepath.Join(dir, "tt", "config.toml"), nil
}

// Load reads the config file (if present) and applies the TT_TOKEN override.
// The returned warning is non-empty when the file is readable by other users.
func Load() (cfg Config, warning string, err error) {
	path, err := Path()
	if err != nil {
		return cfg, "", err
	}

	info, statErr := os.Stat(path)
	switch {
	case statErr == nil:
		if _, err := toml.DecodeFile(path, &cfg); err != nil {
			return cfg, "", fmt.Errorf("reading %s: %w", path, err)
		}
		if info.Mode().Perm()&0o077 != 0 {
			warning = fmt.Sprintf("%s is readable by other users; run: chmod 600 %s", path, path)
		}
	case !errors.Is(statErr, fs.ErrNotExist):
		return cfg, "", statErr
	}

	if t := os.Getenv("TT_TOKEN"); t != "" {
		cfg.Token = t
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = DefaultBaseURL
	}
	return cfg, warning, nil
}

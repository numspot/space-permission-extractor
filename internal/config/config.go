package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	ClientID     string
	ClientSecret string
	SpaceID      string
	BaseURL      string
}

const DefaultBaseURL = "https://api.eu-west-2.numspot.com"

func (c Config) TokenURL() string {
	return strings.TrimRight(c.BaseURL, "/") + "/iam/token"
}

func (c *Config) LoadFromEnvFile() error {
	envPath := ".env"

	wd, err := os.Getwd()
	if err == nil {
		wdEnvPath := filepath.Join(wd, ".env")
		if _, err := os.Stat(wdEnvPath); err == nil {
			envPath = wdEnvPath
		}
	}

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return nil
	}

	file, err := os.Open(envPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			value = strings.Trim(value, "'")
		}

		switch key {
		case "NUMSPOT_CLIENT_ID", "CLIENT_ID":
			if c.ClientID == "" {
				c.ClientID = value
			}
		case "NUMSPOT_CLIENT_SECRET", "CLIENT_SECRET":
			if c.ClientSecret == "" {
				c.ClientSecret = value
			}
		case "NUMSPOT_SPACE_ID", "SPACE_ID":
			if c.SpaceID == "" {
				c.SpaceID = value
			}
		case "NUMSPOT_BASE_URL", "BASE_URL":
			if c.BaseURL == "" {
				c.BaseURL = value
			}
		}
	}

	return scanner.Err()
}

func (c *Config) SaveToEnvFile() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	envPath := filepath.Join(wd, ".env")

	var lines []string
	lines = append(lines, "# Permission Extractor TUI Configuration")
	lines = append(lines, "# Generated automatically - you can edit this file")
	lines = append(lines, "")

	if c.ClientID != "" {
		lines = append(lines, fmt.Sprintf("NUMSPOT_CLIENT_ID=%s", c.ClientID))
	}
	if c.ClientSecret != "" {
		lines = append(lines, fmt.Sprintf("NUMSPOT_CLIENT_SECRET=%s", c.ClientSecret))
	}
	if c.SpaceID != "" {
		lines = append(lines, fmt.Sprintf("NUMSPOT_SPACE_ID=%s", c.SpaceID))
	}
	if c.BaseURL != "" && c.BaseURL != DefaultBaseURL {
		lines = append(lines, fmt.Sprintf("NUMSPOT_BASE_URL=%s", c.BaseURL))
	} else if c.BaseURL == "" {
		lines = append(lines, fmt.Sprintf("NUMSPOT_BASE_URL=%s", DefaultBaseURL))
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(envPath, []byte(content), 0600)
}

func (c *Config) ApplyDefaults() {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
}

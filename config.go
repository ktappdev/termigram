package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const localConfigFile = "config.json"

func localConfigPath() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to resolve executable path: %w", err)
	}
	return filepath.Join(filepath.Dir(exePath), localConfigFile), nil
}

type Config struct {
	TelegramAppID    int    `json:"telegram_app_id"`
	TelegramAppHash  string `json:"telegram_app_hash"`
	TelegramBotToken string `json:"telegram_bot_token,omitempty"`
	SessionPath      string `json:"session_path,omitempty"`
}

func defaultSessionPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}
	return filepath.Join(homeDir, ".modern-telegram-cli", "session.json"), nil
}

func defaultConfig() (Config, error) {
	sessionPath, err := defaultSessionPath()
	if err != nil {
		return Config{}, err
	}
	return Config{SessionPath: sessionPath}, nil
}

func (c Config) HasUserMode() bool {
	return c.TelegramAppID != 0 && c.TelegramAppHash != ""
}

func (c Config) HasBotMode() bool {
	return c.TelegramBotToken != ""
}

func (c Config) ResolveMode(requested string, interactive bool) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(requested))
	if mode != "" && mode != "user" && mode != "bot" {
		return "", fmt.Errorf("invalid mode %q (expected user|bot)", requested)
	}

	if interactive {
		if mode == "bot" {
			return "", fmt.Errorf("interactive mode only supports --mode user")
		}
		if c.HasUserMode() {
			return "user", nil
		}
		if c.HasBotMode() {
			return "", fmt.Errorf("interactive mode requires user credentials; only bot mode is configured")
		}
		return "", fmt.Errorf("no authentication mode configured")
	}

	if mode == "user" {
		if !c.HasUserMode() {
			return "", fmt.Errorf("user mode requested but user credentials are not configured")
		}
		return "user", nil
	}
	if mode == "bot" {
		if !c.HasBotMode() {
			return "", fmt.Errorf("bot mode requested but telegram_bot_token is not configured")
		}
		return "bot", nil
	}

	if c.HasUserMode() && c.HasBotMode() {
		return "", fmt.Errorf("both user and bot credentials are configured; specify --mode user|bot")
	}
	if c.HasUserMode() {
		return "user", nil
	}
	if c.HasBotMode() {
		return "bot", nil
	}
	return "", fmt.Errorf("no authentication mode configured")
}

func loadConfig() (Config, error) {
	cfg, err := defaultConfig()
	if err != nil {
		return Config{}, err
	}

	configPath, err := localConfigPath()
	if err != nil {
		return Config{}, err
	}

	if data, err := os.ReadFile(configPath); err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return Config{}, fmt.Errorf("invalid %s: %w", configPath, err)
		}
	} else if !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("failed to read %s: %w", configPath, err)
	}

	if appIDStr := strings.TrimSpace(os.Getenv("TELEGRAM_APP_ID")); appIDStr != "" {
		appID, err := strconv.Atoi(appIDStr)
		if err != nil {
			return Config{}, fmt.Errorf("TELEGRAM_APP_ID must be a number, got: %s", appIDStr)
		}
		cfg.TelegramAppID = appID
	}

	if appHash := strings.TrimSpace(os.Getenv("TELEGRAM_APP_HASH")); appHash != "" {
		cfg.TelegramAppHash = appHash
	}

	if botToken := strings.TrimSpace(os.Getenv("TELEGRAM_BOT_TOKEN")); botToken != "" {
		cfg.TelegramBotToken = botToken
	}

	if sessionPath := strings.TrimSpace(os.Getenv("TELEGRAM_SESSION_PATH")); sessionPath != "" {
		cfg.SessionPath = sessionPath
	}

	cfg.TelegramAppHash = strings.TrimSpace(cfg.TelegramAppHash)
	cfg.TelegramBotToken = strings.TrimSpace(cfg.TelegramBotToken)
	cfg.SessionPath = strings.TrimSpace(cfg.SessionPath)

	if (cfg.TelegramAppID == 0) != (cfg.TelegramAppHash == "") {
		return Config{}, fmt.Errorf("incomplete user credentials: set both telegram_app_id and telegram_app_hash")
	}

	if !cfg.HasUserMode() && !cfg.HasBotMode() {
		return Config{}, fmt.Errorf("missing Telegram credentials: set user credentials (telegram_app_id + telegram_app_hash) and/or telegram_bot_token in %s or via TELEGRAM_APP_ID/TELEGRAM_APP_HASH/TELEGRAM_BOT_TOKEN", configPath)
	}

	if cfg.HasUserMode() && cfg.SessionPath == "" {
		cfg.SessionPath, err = defaultSessionPath()
		if err != nil {
			return Config{}, err
		}
	}

	return cfg, nil
}

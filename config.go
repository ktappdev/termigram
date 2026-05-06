package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	TelegramAppID                  int           `json:"telegram_app_id"`
	TelegramAppHash                string        `json:"telegram_app_hash"`
	SessionPath                    string        `json:"session_path,omitempty"`
	TranscriptLimit                int           `json:"transcript_limit,omitempty"`
	TranscriptHistoryFetchLimit    int           `json:"transcript_history_fetch_limit,omitempty"`
	MaxRemoteImageBytes            int64         `json:"max_remote_image_bytes,omitempty"`
	InteractiveResumeIdleThreshold time.Duration `json:"interactive_resume_idle_threshold,omitempty"`
	InlineImageMaxBytes            int           `json:"inline_image_max_bytes,omitempty"`
}

func (c *Config) UnmarshalJSON(data []byte) error {
	type rawConfig struct {
		TelegramAppID                  *int            `json:"telegram_app_id"`
		TelegramAppHash                *string         `json:"telegram_app_hash"`
		SessionPath                    *string         `json:"session_path,omitempty"`
		TranscriptLimit                *int            `json:"transcript_limit,omitempty"`
		TranscriptHistoryFetchLimit    *int            `json:"transcript_history_fetch_limit,omitempty"`
		MaxRemoteImageBytes            *int64          `json:"max_remote_image_bytes,omitempty"`
		InteractiveResumeIdleThreshold json.RawMessage `json:"interactive_resume_idle_threshold,omitempty"`
		InlineImageMaxBytes            *int            `json:"inline_image_max_bytes,omitempty"`
	}

	var raw rawConfig
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if raw.TelegramAppID != nil {
		c.TelegramAppID = *raw.TelegramAppID
	}
	if raw.TelegramAppHash != nil {
		c.TelegramAppHash = *raw.TelegramAppHash
	}
	if raw.SessionPath != nil {
		c.SessionPath = *raw.SessionPath
	}
	if raw.TranscriptLimit != nil {
		c.TranscriptLimit = *raw.TranscriptLimit
	}
	if raw.TranscriptHistoryFetchLimit != nil {
		c.TranscriptHistoryFetchLimit = *raw.TranscriptHistoryFetchLimit
	}
	if raw.MaxRemoteImageBytes != nil {
		c.MaxRemoteImageBytes = *raw.MaxRemoteImageBytes
	}
	if raw.InlineImageMaxBytes != nil {
		c.InlineImageMaxBytes = *raw.InlineImageMaxBytes
	}

	if len(raw.InteractiveResumeIdleThreshold) == 0 || string(raw.InteractiveResumeIdleThreshold) == "null" {
		return nil
	}

	var durationText string
	if err := json.Unmarshal(raw.InteractiveResumeIdleThreshold, &durationText); err == nil {
		duration, err := time.ParseDuration(strings.TrimSpace(durationText))
		if err != nil {
			return fmt.Errorf("interactive_resume_idle_threshold must be a valid duration, got %q", durationText)
		}
		c.InteractiveResumeIdleThreshold = duration
		return nil
	}

	var seconds int64
	if err := json.Unmarshal(raw.InteractiveResumeIdleThreshold, &seconds); err == nil {
		c.InteractiveResumeIdleThreshold = time.Duration(seconds) * time.Second
		return nil
	}

	return fmt.Errorf("interactive_resume_idle_threshold must be a duration string or integer seconds")
}

func defaultSessionPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}
	return filepath.Join(homeDir, ".termigram", "session.json"), nil
}

func bakedConfig() (Config, error) {
	cfg := Config{TelegramAppHash: strings.TrimSpace(telegramAppHashBaked)}

	appIDStr := strings.TrimSpace(telegramAppIDBaked)
	if appIDStr != "" {
		appID, err := strconv.Atoi(appIDStr)
		if err != nil {
			return Config{}, fmt.Errorf("invalid baked Telegram app id %q: %w", appIDStr, err)
		}
		cfg.TelegramAppID = appID
	}

	if cfg.hasPartialCredentials() {
		return Config{}, fmt.Errorf("incomplete baked Telegram credentials: set both telegramAppIDBaked and telegramAppHashBaked via ldflags")
	}

	return cfg, nil
}

func defaultConfig() (Config, error) {
	sessionPath, err := defaultSessionPath()
	if err != nil {
		return Config{}, fmt.Errorf("failed to get default session path: %w", err)
	}

	cfg, err := bakedConfig()
	if err != nil {
		return Config{}, fmt.Errorf("failed to load baked config: %w", err)
	}
	cfg.SessionPath = sessionPath
	return cfg, nil
}

func (c Config) hasPartialCredentials() bool {
	return (c.TelegramAppID == 0) != (c.TelegramAppHash == "")
}

func (c Config) HasUserMode() bool {
	return c.TelegramAppID != 0 && c.TelegramAppHash != ""
}

func (c Config) ResolveMode(requested string, interactive bool) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(requested))
	if mode != "" && mode != "user" {
		return "", fmt.Errorf("invalid mode %q (only 'user' mode is supported)", requested)
	}

	if !c.HasUserMode() {
		return "", fmt.Errorf("user credentials not configured: provide telegram_app_id and telegram_app_hash via TELEGRAM_APP_ID/TELEGRAM_APP_HASH, %s, or baked-in build credentials", localConfigFile)
	}
	return "user", nil
}

func loadConfig() (Config, error) {
	cfg, err := defaultConfig()
	if err != nil {
		return Config{}, fmt.Errorf("failed to load default config: %w", err)
	}

	configPath, err := localConfigPath()
	if err != nil {
		return Config{}, fmt.Errorf("failed to get local config path: %w", err)
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

	if sessionPath := strings.TrimSpace(os.Getenv("TELEGRAM_SESSION_PATH")); sessionPath != "" {
		cfg.SessionPath = sessionPath
	}

	if transcriptLimitStr := strings.TrimSpace(os.Getenv("TERMIGRAM_TRANSCRIPT_LIMIT")); transcriptLimitStr != "" {
		transcriptLimit, err := strconv.Atoi(transcriptLimitStr)
		if err != nil {
			return Config{}, fmt.Errorf("TERMIGRAM_TRANSCRIPT_LIMIT must be a number, got: %s", transcriptLimitStr)
		}
		cfg.TranscriptLimit = transcriptLimit
	}

	if transcriptHistoryFetchLimitStr := strings.TrimSpace(os.Getenv("TERMIGRAM_TRANSCRIPT_HISTORY_FETCH_LIMIT")); transcriptHistoryFetchLimitStr != "" {
		transcriptHistoryFetchLimit, err := strconv.Atoi(transcriptHistoryFetchLimitStr)
		if err != nil {
			return Config{}, fmt.Errorf("TERMIGRAM_TRANSCRIPT_HISTORY_FETCH_LIMIT must be a number, got: %s", transcriptHistoryFetchLimitStr)
		}
		cfg.TranscriptHistoryFetchLimit = transcriptHistoryFetchLimit
	}

	if maxRemoteImageBytesStr := strings.TrimSpace(os.Getenv("TERMIGRAM_MAX_REMOTE_IMAGE_BYTES")); maxRemoteImageBytesStr != "" {
		maxRemoteImageBytes, err := strconv.ParseInt(maxRemoteImageBytesStr, 10, 64)
		if err != nil {
			return Config{}, fmt.Errorf("TERMIGRAM_MAX_REMOTE_IMAGE_BYTES must be a number, got: %s", maxRemoteImageBytesStr)
		}
		cfg.MaxRemoteImageBytes = maxRemoteImageBytes
	}

	if interactiveResumeIdleThresholdStr := strings.TrimSpace(os.Getenv("TERMIGRAM_INTERACTIVE_RESUME_IDLE_THRESHOLD")); interactiveResumeIdleThresholdStr != "" {
		seconds, err := strconv.Atoi(interactiveResumeIdleThresholdStr)
		if err != nil {
			return Config{}, fmt.Errorf("TERMIGRAM_INTERACTIVE_RESUME_IDLE_THRESHOLD must be integer seconds, got: %s", interactiveResumeIdleThresholdStr)
		}
		cfg.InteractiveResumeIdleThreshold = time.Duration(seconds) * time.Second
	}

	if inlineImageMaxBytesStr := strings.TrimSpace(os.Getenv("TERMIGRAM_INLINE_IMAGE_MAX_BYTES")); inlineImageMaxBytesStr != "" {
		inlineImageMaxBytes, err := strconv.Atoi(inlineImageMaxBytesStr)
		if err != nil {
			return Config{}, fmt.Errorf("TERMIGRAM_INLINE_IMAGE_MAX_BYTES must be a number, got: %s", inlineImageMaxBytesStr)
		}
		cfg.InlineImageMaxBytes = inlineImageMaxBytes
	}

	cfg.TelegramAppHash = strings.TrimSpace(cfg.TelegramAppHash)
	cfg.SessionPath = strings.TrimSpace(cfg.SessionPath)

	if cfg.hasPartialCredentials() {
		return Config{}, fmt.Errorf("incomplete user credentials: set both telegram_app_id and telegram_app_hash (via env, %s, or baked-in build credentials)", configPath)
	}

	if !cfg.HasUserMode() {
		return Config{}, fmt.Errorf("missing Telegram credentials: provide telegram_app_id and telegram_app_hash via TELEGRAM_APP_ID/TELEGRAM_APP_HASH, %s, or baked-in build credentials", configPath)
	}

	if cfg.SessionPath == "" {
		cfg.SessionPath, err = defaultSessionPath()
		if err != nil {
			return Config{}, fmt.Errorf("failed to get default session path: %w", err)
		}
	}

	return cfg, nil
}

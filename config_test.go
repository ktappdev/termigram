package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- helpers ---

// saveBakedVars saves and restores package-level baked credential variables.
func saveBakedVars(t *testing.T) {
	oldID, oldHash := telegramAppIDBaked, telegramAppHashBaked
	t.Cleanup(func() {
		telegramAppIDBaked = oldID
		telegramAppHashBaked = oldHash
	})
}

// writeLocalConfig writes content to the config.json path adjacent to the
// test binary and schedules cleanup.
func writeLocalConfig(t *testing.T, content string) {
	t.Helper()
	cfgPath, err := localConfigPath()
	if err != nil {
		t.Fatalf("localConfigPath: %v", err)
	}
	if err := os.WriteFile(cfgPath, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	t.Cleanup(func() { os.Remove(cfgPath) })
}

// removeLocalConfig removes any config.json adjacent to the test binary and
// schedules cleanup for tests that need no file present.
func removeLocalConfig(t *testing.T) {
	t.Helper()
	cfgPath, err := localConfigPath()
	if err != nil {
		return
	}
	os.Remove(cfgPath) // ignore missing
	t.Cleanup(func() { os.Remove(cfgPath) })
}

// clearCredentialEnv empties the environment variables that loadConfig reads.
func clearCredentialEnv(t *testing.T) {
	t.Setenv("TELEGRAM_APP_ID", "")
	t.Setenv("TELEGRAM_APP_HASH", "")
	t.Setenv("TELEGRAM_SESSION_PATH", "")
	t.Setenv("TERMIGRAM_TRANSCRIPT_LIMIT", "")
	t.Setenv("TERMIGRAM_TRANSCRIPT_HISTORY_FETCH_LIMIT", "")
	t.Setenv("TERMIGRAM_MAX_REMOTE_IMAGE_BYTES", "")
	t.Setenv("TERMIGRAM_INTERACTIVE_RESUME_IDLE_THRESHOLD", "")
	t.Setenv("TERMIGRAM_INLINE_IMAGE_MAX_BYTES", "")
}

// --- TestDefaultConfig ---

func TestDefaultConfig(t *testing.T) {
	saveBakedVars(t)
	telegramAppIDBaked = ""
	telegramAppHashBaked = ""

	cfg, err := defaultConfig()
	if err != nil {
		t.Fatalf("defaultConfig error: %v", err)
	}

	if cfg.TelegramAppID != 0 {
		t.Errorf("expected TelegramAppID=0, got %d", cfg.TelegramAppID)
	}
	if cfg.TelegramAppHash != "" {
		t.Errorf("expected TelegramAppHash=\"\", got %q", cfg.TelegramAppHash)
	}

	// SessionPath should be ~/.termigram/session.json
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir error: %v", err)
	}
	wantSession := filepath.Join(home, ".termigram", "session.json")
	if cfg.SessionPath != wantSession {
		t.Errorf("expected SessionPath=%q, got %q", wantSession, cfg.SessionPath)
	}
}

// --- TestBakedConfig ---

func TestBakedConfig_BothSet(t *testing.T) {
	saveBakedVars(t)
	telegramAppIDBaked = "12345"
	telegramAppHashBaked = "abc123"

	cfg, err := bakedConfig()
	if err != nil {
		t.Fatalf("bakedConfig error: %v", err)
	}
	if cfg.TelegramAppID != 12345 {
		t.Errorf("expected TelegramAppID=12345, got %d", cfg.TelegramAppID)
	}
	if cfg.TelegramAppHash != "abc123" {
		t.Errorf("expected TelegramAppHash=abc123, got %q", cfg.TelegramAppHash)
	}
}

func TestBakedConfig_PartialOnlyID(t *testing.T) {
	saveBakedVars(t)
	telegramAppIDBaked = "12345"
	telegramAppHashBaked = ""

	_, err := bakedConfig()
	if err == nil {
		t.Fatal("expected error for partial baked credentials (only ID)")
	}
	if !strings.Contains(err.Error(), "incomplete baked Telegram credentials") {
		t.Errorf("expected incomplete credentials error, got: %v", err)
	}
}

func TestBakedConfig_PartialOnlyHash(t *testing.T) {
	saveBakedVars(t)
	telegramAppIDBaked = ""
	telegramAppHashBaked = "abc123"

	_, err := bakedConfig()
	if err == nil {
		t.Fatal("expected error for partial baked credentials (only hash)")
	}
	if !strings.Contains(err.Error(), "incomplete baked Telegram credentials") {
		t.Errorf("expected incomplete credentials error, got: %v", err)
	}
}

func TestBakedConfig_NeitherSet(t *testing.T) {
	saveBakedVars(t)
	telegramAppIDBaked = ""
	telegramAppHashBaked = ""

	cfg, err := bakedConfig()
	if err != nil {
		t.Fatalf("bakedConfig error: %v", err)
	}
	if cfg.TelegramAppID != 0 {
		t.Errorf("expected TelegramAppID=0, got %d", cfg.TelegramAppID)
	}
	if cfg.TelegramAppHash != "" {
		t.Errorf("expected TelegramAppHash=\"\", got %q", cfg.TelegramAppHash)
	}
}

// --- TestLoadConfig ---

func TestLoadConfig_ValidJSON(t *testing.T) {
	saveBakedVars(t)
	telegramAppIDBaked = ""
	telegramAppHashBaked = ""
	clearCredentialEnv(t)

	writeLocalConfig(t, `{"telegram_app_id": 12345, "telegram_app_hash": "filehash"}`)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}
	if cfg.TelegramAppID != 12345 {
		t.Errorf("expected TelegramAppID=12345, got %d", cfg.TelegramAppID)
	}
	if cfg.TelegramAppHash != "filehash" {
		t.Errorf("expected TelegramAppHash=filehash, got %q", cfg.TelegramAppHash)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	saveBakedVars(t)
	telegramAppIDBaked = ""
	telegramAppHashBaked = ""
	removeLocalConfig(t)

	t.Setenv("TELEGRAM_APP_ID", "11111")
	t.Setenv("TELEGRAM_APP_HASH", "envhash")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("loadConfig error: %v", err)
	}
	if cfg.TelegramAppID != 11111 {
		t.Errorf("expected TelegramAppID=11111, got %d", cfg.TelegramAppID)
	}
	if cfg.TelegramAppHash != "envhash" {
		t.Errorf("expected TelegramAppHash=envhash, got %q", cfg.TelegramAppHash)
	}
}

func TestLoadConfig_InvalidJSON(t *testing.T) {
	saveBakedVars(t)
	telegramAppIDBaked = ""
	telegramAppHashBaked = ""

	writeLocalConfig(t, `{invalid}`)

	_, err := loadConfig()
	if err == nil {
		t.Fatal("expected error for invalid config JSON, got nil")
	}
}

// --- TestHasPartialCredentials ---

func TestHasPartialCredentials_Both(t *testing.T) {
	c := Config{TelegramAppID: 123, TelegramAppHash: "hash"}
	if c.hasPartialCredentials() {
		t.Error("expected hasPartialCredentials=false when both are set")
	}
}

func TestHasPartialCredentials_Neither(t *testing.T) {
	c := Config{TelegramAppID: 0, TelegramAppHash: ""}
	if c.hasPartialCredentials() {
		t.Error("expected hasPartialCredentials=false when neither is set")
	}
}

func TestHasPartialCredentials_OnlyID(t *testing.T) {
	c := Config{TelegramAppID: 123, TelegramAppHash: ""}
	if !c.hasPartialCredentials() {
		t.Error("expected hasPartialCredentials=true when only ID is set")
	}
}

func TestHasPartialCredentials_OnlyHash(t *testing.T) {
	c := Config{TelegramAppID: 0, TelegramAppHash: "hash"}
	if !c.hasPartialCredentials() {
		t.Error("expected hasPartialCredentials=true when only hash is set")
	}
}

// --- TestValidate via loadConfig ---

func TestValidate_Complete(t *testing.T) {
	saveBakedVars(t)
	telegramAppIDBaked = ""
	telegramAppHashBaked = ""
	removeLocalConfig(t)

	t.Setenv("TELEGRAM_APP_ID", "55555")
	t.Setenv("TELEGRAM_APP_HASH", "validhash")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("expected no error for complete credentials, got: %v", err)
	}
	if cfg.TelegramAppID != 55555 {
		t.Errorf("expected TelegramAppID=55555, got %d", cfg.TelegramAppID)
	}
	if cfg.TelegramAppHash != "validhash" {
		t.Errorf("expected TelegramAppHash=validhash, got %q", cfg.TelegramAppHash)
	}
}

func TestValidate_Partial(t *testing.T) {
	saveBakedVars(t)
	telegramAppIDBaked = ""
	telegramAppHashBaked = ""
	removeLocalConfig(t)

	t.Setenv("TELEGRAM_APP_ID", "55555")
	t.Setenv("TELEGRAM_APP_HASH", "")  // explicitly clear to prevent env leakage

	_, err := loadConfig()
	if err == nil {
		t.Fatal("expected error for partial credentials, got nil")
	}
	if !strings.Contains(err.Error(), "incomplete user credentials") {
		t.Errorf("expected incomplete user credentials error, got: %v", err)
	}
}

// --- TestLoadCredentials_PriorityChain ---

// Priority: baked credentials are used when no config file or env vars override them.
func TestLoadCredentials_PriorityChain(t *testing.T) {
	saveBakedVars(t)
	telegramAppIDBaked = "99999"
	telegramAppHashBaked = "bakedhash"
	removeLocalConfig(t)

	clearCredentialEnv(t)

	cfg, err := loadConfig()
	if err != nil {
		t.Fatalf("expected no error with baked credentials, got: %v", err)
	}
	if cfg.TelegramAppID != 99999 {
		t.Errorf("expected TelegramAppID=99999 (from baked), got %d", cfg.TelegramAppID)
	}
	if cfg.TelegramAppHash != "bakedhash" {
		t.Errorf("expected TelegramAppHash=bakedhash (from baked), got %q", cfg.TelegramAppHash)
	}
}

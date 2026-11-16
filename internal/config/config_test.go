package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigSaveLoadAndEnv(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")

	cfg := Default()
	if cfg.IsConfigured() {
		t.Fatal("default config should not be considered configured")
	}

	cfg.Key = "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAITest"
	cfg.Filesystem = "/share/CACHEDEV2_DATA"
	cfg.CustomEnv = []EnvPair{
		{Key: "LOG_LEVEL", Value: "debug"},
		{Key: "FUTURE_VAR", Value: "42"},
	}
	if err := cfg.Save(cfgPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !loaded.IsConfigured() {
		t.Fatal("loaded config should be configured")
	}

	envMap := loaded.AgentEnvMap("/tmp/agent-data")
	if envMap["LOG_LEVEL"] != "debug" {
		t.Fatalf("expected custom_env override LOG_LEVEL=debug, got %q", envMap["LOG_LEVEL"])
	}
	if envMap["FUTURE_VAR"] != "42" {
		t.Fatalf("expected FUTURE_VAR=42, got %q", envMap["FUTURE_VAR"])
	}
	if envMap["DATA_DIR"] != "/tmp/agent-data" {
		t.Fatalf("expected DATA_DIR=/tmp/agent-data, got %q", envMap["DATA_DIR"])
	}

	envFile := filepath.Join(dir, "beszel-agent.env")
	if err := loaded.WriteEnvFile(envFile, "/tmp/agent-data"); err != nil {
		t.Fatalf("WriteEnvFile failed: %v", err)
	}
	raw, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !strings.Contains(string(raw), "FILESYSTEM=/share/CACHEDEV2_DATA") {
		t.Fatalf("env file missing FILESYSTEM: %s", string(raw))
	}
}

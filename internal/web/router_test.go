package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"beszel-qnap/internal/agentctl"
	"beszel-qnap/internal/auth"
	"beszel-qnap/internal/config"
)

type stubVerifier struct {
	err error
}

func (s stubVerifier) VerifyRequest(_ context.Context, _ *http.Request) (*auth.Session, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &auth.Session{Username: "admin", IsAdmin: true}, nil
}

func TestRouterAuthAndConfig(t *testing.T) {
	dir := t.TempDir()
	mgr := agentctl.NewManager(&agentctl.Paths{Root: dir})

	unauthSrv := NewServer(mgr, stubVerifier{err: auth.ErrUnauthorized})
	rec := httptest.NewRecorder()
	unauthSrv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api.cgi?action=status", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}

	srv := NewServer(mgr, stubVerifier{})

	// Save config
	payload := map[string]any{
		"config": config.Config{
			AutoStart:      true,
			DownloadSource: "mirror",
			Listen:         "45876",
			Key:            "ssh-ed25519 AAAATestKey",
		},
		"restart": false,
	}
	body, _ := json.Marshal(payload)
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api.cgi?action=config", bytes.NewReader(body)))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 saving config, got %d: %s", rec.Code, rec.Body.String())
	}

	// Read status
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api.cgi?action=status", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 status, got %d", rec.Code)
	}
	var resp struct {
		Status agentctl.Status `json:"status"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if !resp.Status.Configured {
		t.Fatal("expected Configured=true after saving key")
	}
}

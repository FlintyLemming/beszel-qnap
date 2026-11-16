package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestExtractSIDPriority(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api.cgi/status?sid=query-sid", nil)
	req.AddCookie(&http.Cookie{Name: "NAS_SID", Value: "cookie-sid"})
	if got := ExtractSID(req); got != "cookie-sid" {
		t.Fatalf("expected cookie-sid, got %q", got)
	}

	req.Header.Set("X-QTS-SID", "header-sid")
	if got := ExtractSID(req); got != "header-sid" {
		t.Fatalf("expected header-sid, got %q", got)
	}

	reqNoCookie := httptest.NewRequest(http.MethodGet, "/api.cgi/status?sid=query-sid", nil)
	if got := ExtractSID(reqNoCookie); got != "query-sid" {
		t.Fatalf("expected query-sid, got %q", got)
	}
}

func TestVerifySID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Query().Get("sid") {
		case "valid-admin":
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><QDocRoot version="1.0"><authPassed><![CDATA[1]]></authPassed><isAdmin><![CDATA[1]]></isAdmin><username><![CDATA[FlintyLemming]]></username></QDocRoot>`))
		case "valid-nonadmin":
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><QDocRoot version="1.0"><authPassed><![CDATA[1]]></authPassed><isAdmin><![CDATA[0]]></isAdmin><username><![CDATA[guest]]></username></QDocRoot>`))
		default:
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?><QDocRoot version="1.0"><authPassed><![CDATA[0]]></authPassed></QDocRoot>`))
		}
	}))
	defer srv.Close()

	v := &Verifier{Endpoints: []string{srv.URL}, Client: srv.Client()}

	sess, err := v.VerifySID(context.Background(), "valid-admin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sess.IsAdmin || sess.Username != "FlintyLemming" {
		t.Fatalf("unexpected session: %+v", sess)
	}

	if _, err := v.VerifySID(context.Background(), "valid-nonadmin"); !errors.Is(err, ErrNotAdmin) {
		t.Fatalf("expected ErrNotAdmin, got %v", err)
	}
	if _, err := v.VerifySID(context.Background(), "bad"); !errors.Is(err, ErrUnauthorized) {
		t.Fatalf("expected ErrUnauthorized, got %v", err)
	}
	if _, err := v.VerifySID(context.Background(), ""); !errors.Is(err, ErrMissingSID) {
		t.Fatalf("expected ErrMissingSID, got %v", err)
	}
}

package auth

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	ErrMissingSID   = errors.New("missing QTS session ID")
	ErrUnauthorized = errors.New("QTS session is invalid or expired")
	ErrNotAdmin     = errors.New("QTS administrator privileges required")
)

type Session struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"is_admin"`
}

type authLoginXML struct {
	AuthPassed string `xml:"authPassed"`
	IsAdmin    string `xml:"isAdmin"`
	Username   string `xml:"username"`
	GroupName  string `xml:"groupname"`
}

type Verifier struct {
	Endpoints []string
	Client    *http.Client
}

// NewQTSVerifier returns a Verifier targeting the local QTS CGI daemon.
func NewQTSVerifier() *Verifier {
	endpoints := []string{"http://127.0.0.1:58080/cgi-bin/authLogin.cgi"}
	if port := readQTSWebPort(); port != "" && port != "58080" {
		endpoints = append(endpoints, fmt.Sprintf("http://127.0.0.1:%s/cgi-bin/authLogin.cgi", port))
	}
	return &Verifier{
		Endpoints: endpoints,
		Client: &http.Client{
			Timeout: 4 * time.Second,
		},
	}
}

func readQTSWebPort() string {
	if _, err := os.Stat("/sbin/getcfg"); err != nil {
		return ""
	}
	out, err := exec.Command("/sbin/getcfg", "System", "Web Access Port", "-f", "/etc/config/uLinux.conf").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// ExtractSID extracts the QTS session ID from cookies, headers, or query params.
func ExtractSID(r *http.Request) string {
	if sid := strings.TrimSpace(r.Header.Get("X-QTS-SID")); sid != "" {
		return sid
	}
	for _, name := range []string{"NAS_SID", "nas_sid", "QTS_SSID", "QTS_SSL_SSID"} {
		if c, err := r.Cookie(name); err == nil {
			if val := strings.TrimSpace(c.Value); val != "" {
				return val
			}
		}
	}
	if sid := strings.TrimSpace(r.URL.Query().Get("sid")); sid != "" {
		return sid
	}
	return ""
}

// VerifyRequest validates the caller's QTS session and requires admin access.
func (v *Verifier) VerifyRequest(ctx context.Context, r *http.Request) (*Session, error) {
	sid := ExtractSID(r)
	if sid == "" {
		return nil, ErrMissingSID
	}
	return v.VerifySID(ctx, sid)
}

// VerifySID checks the QTS session token against authLogin.cgi.
func (v *Verifier) VerifySID(ctx context.Context, sid string) (*Session, error) {
	if sid == "" {
		return nil, ErrMissingSID
	}
	client := v.Client
	if client == nil {
		client = &http.Client{Timeout: 4 * time.Second}
	}

	var lastErr error
	for _, ep := range v.Endpoints {
		reqURL := ep + "?sid=" + url.QueryEscape(sid)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			lastErr = err
			continue
		}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode >= 400 {
			lastErr = fmt.Errorf("authLogin.cgi returned HTTP %d", resp.StatusCode)
			continue
		}

		var doc authLoginXML
		if err := xml.Unmarshal(body, &doc); err != nil {
			lastErr = fmt.Errorf("invalid authLogin XML: %w", err)
			continue
		}
		if strings.TrimSpace(doc.AuthPassed) != "1" {
			return nil, ErrUnauthorized
		}
		isAdmin := strings.TrimSpace(doc.IsAdmin) == "1" || strings.EqualFold(strings.TrimSpace(doc.Username), "admin")
		if !isAdmin {
			return nil, ErrNotAdmin
		}
		return &Session{
			Username: strings.TrimSpace(doc.Username),
			IsAdmin:  true,
		}, nil
	}
	if lastErr != nil {
		return nil, fmt.Errorf("failed to verify QTS session: %w", lastErr)
	}
	return nil, ErrUnauthorized
}

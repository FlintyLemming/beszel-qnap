package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
	"time"
)

const userAgent = "beszel-agent/1.0"

type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Digest      string `json:"digest"`
	Size        int64  `json:"size"`
}

type Release struct {
	TagName     string  `json:"tag_name"`
	Name        string  `json:"name"`
	PublishedAt string  `json:"published_at"`
	HTMLURL     string  `json:"html_url"`
	Body        string  `json:"body"`
	Prerelease  bool    `json:"prerelease"`
	Draft       bool    `json:"draft"`
	Assets      []Asset `json:"assets"`
}

type ReleaseSummary struct {
	TagName     string `json:"tag_name"`
	Version     string `json:"version"`
	PublishedAt string `json:"published_at"`
	HTMLURL     string `json:"html_url"`
	Body        string `json:"body"`
	Prerelease  bool   `json:"prerelease"`
	Asset       *Asset `json:"asset,omitempty"`
}

func NewHTTPClient(proxyURL string, timeout time.Duration) (*http.Client, error) {
	tr := http.DefaultTransport.(*http.Transport).Clone()
	if p := strings.TrimSpace(proxyURL); p != "" {
		u, err := url.Parse(p)
		if err != nil {
			return nil, fmt.Errorf("invalid http_proxy URL: %w", err)
		}
		tr.Proxy = http.ProxyURL(u)
	}
	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}, nil
}

func FetchReleases(ctx context.Context, source, customPrefix, proxyURL string, preferGlibc bool) ([]ReleaseSummary, error) {
	client, err := NewHTTPClient(proxyURL, 20*time.Second)
	if err != nil {
		return nil, err
	}

	var apiURL string
	var singleMode bool
	switch source {
	case "mirror":
		apiURL = "https://gh.beszel.dev/repos/henrygd/beszel/releases/latest?api=true"
		singleMode = true
	default:
		apiURL = "https://api.github.com/repos/henrygd/beszel/releases?per_page=15"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("release API HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var rawReleases []Release
	if singleMode {
		var single Release
		if err := json.Unmarshal(body, &single); err != nil {
			return nil, err
		}
		if single.TagName != "" {
			rawReleases = []Release{single}
		}
	} else {
		if err := json.Unmarshal(body, &rawReleases); err != nil {
			return nil, err
		}
	}

	out := make([]ReleaseSummary, 0, len(rawReleases))
	for _, r := range rawReleases {
		if r.Draft {
			continue
		}
		asset := SelectAgentAsset(r.Assets, runtime.GOOS, runtime.GOARCH, preferGlibc)
		if asset != nil {
			copied := *asset
			copied.DownloadURL = RewriteDownloadURL(copied.DownloadURL, source, customPrefix)
			asset = &copied
		}
		out = append(out, ReleaseSummary{
			TagName:     r.TagName,
			Version:     strings.TrimPrefix(r.TagName, "v"),
			PublishedAt: r.PublishedAt,
			HTMLURL:     r.HTMLURL,
			Body:        r.Body,
			Prerelease:  r.Prerelease,
			Asset:       asset,
		})
	}
	return out, nil
}

func FetchReleaseByTag(ctx context.Context, source, customPrefix, proxyURL, tag string, preferGlibc bool) (*ReleaseSummary, error) {
	client, err := NewHTTPClient(proxyURL, 20*time.Second)
	if err != nil {
		return nil, err
	}

	var apiURL string
	switch {
	case source == "mirror" && (tag == "" || tag == "latest"):
		apiURL = "https://gh.beszel.dev/repos/henrygd/beszel/releases/latest?api=true"
	case tag == "" || tag == "latest":
		apiURL = "https://api.github.com/repos/henrygd/beszel/releases/latest"
	default:
		if !strings.HasPrefix(tag, "v") {
			tag = "v" + tag
		}
		apiURL = fmt.Sprintf("https://api.github.com/repos/henrygd/beszel/releases/tags/%s", url.PathEscape(tag))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("release API HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var rel Release
	if err := json.Unmarshal(body, &rel); err != nil {
		return nil, err
	}
	asset := SelectAgentAsset(rel.Assets, runtime.GOOS, runtime.GOARCH, preferGlibc)
	if asset == nil {
		return nil, fmt.Errorf("no matching beszel-agent asset found for %s/%s in %s", runtime.GOOS, runtime.GOARCH, rel.TagName)
	}
	copied := *asset
	copied.DownloadURL = RewriteDownloadURL(copied.DownloadURL, source, customPrefix)
	return &ReleaseSummary{
		TagName:     rel.TagName,
		Version:     strings.TrimPrefix(rel.TagName, "v"),
		PublishedAt: rel.PublishedAt,
		HTMLURL:     rel.HTMLURL,
		Body:        rel.Body,
		Prerelease:  rel.Prerelease,
		Asset:       &copied,
	}, nil
}

func SelectAgentAsset(assets []Asset, goos, goarch string, preferGlibc bool) *Asset {
	if goos == "" {
		goos = "linux"
	}
	candidates := assetCandidateNames(goos, goarch, preferGlibc)
	for _, want := range candidates {
		for i := range assets {
			if assets[i].Name == want {
				return &assets[i]
			}
		}
	}
	return nil
}

func assetCandidateNames(goos, goarch string, preferGlibc bool) []string {
	switch goarch {
	case "amd64":
		std := fmt.Sprintf("beszel-agent_%s_amd64.tar.gz", goos)
		glibc := fmt.Sprintf("beszel-agent_%s_amd64_glibc.tar.gz", goos)
		if preferGlibc {
			return []string{glibc, std}
		}
		return []string{std, glibc}
	case "arm64":
		return []string{fmt.Sprintf("beszel-agent_%s_arm64.tar.gz", goos)}
	case "arm":
		return []string{
			fmt.Sprintf("beszel-agent_%s_armv7.tar.gz", goos),
			fmt.Sprintf("beszel-agent_%s_arm.tar.gz", goos),
			fmt.Sprintf("beszel-agent_%s_armv5.tar.gz", goos),
		}
	default:
		return []string{fmt.Sprintf("beszel-agent_%s_%s.tar.gz", goos, goarch)}
	}
}

func RewriteDownloadURL(rawURL, source, customPrefix string) string {
	switch source {
	case "mirror":
		return strings.Replace(rawURL, "https://github.com", "https://gh.beszel.dev", 1)
	case "custom":
		prefix := strings.TrimSpace(customPrefix)
		if prefix == "" {
			return rawURL
		}
		if !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}
		return prefix + rawURL
	default:
		return rawURL
	}
}

var ErrNoDownloadURL = errors.New("download URL is empty")

package updater

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"
)

func TestSelectAgentAsset(t *testing.T) {
	assets := []Asset{
		{Name: "beszel-agent_linux_amd64.tar.gz", DownloadURL: "https://github.com/a"},
		{Name: "beszel-agent_linux_amd64_glibc.tar.gz", DownloadURL: "https://github.com/b"},
		{Name: "beszel-agent_linux_arm64.tar.gz", DownloadURL: "https://github.com/c"},
	}

	got := SelectAgentAsset(assets, "linux", "amd64", false)
	if got == nil || got.Name != "beszel-agent_linux_amd64.tar.gz" {
		t.Fatalf("expected standard amd64 asset, got %+v", got)
	}

	gotGlibc := SelectAgentAsset(assets, "linux", "amd64", true)
	if gotGlibc == nil || gotGlibc.Name != "beszel-agent_linux_amd64_glibc.tar.gz" {
		t.Fatalf("expected glibc amd64 asset, got %+v", gotGlibc)
	}

	gotArm := SelectAgentAsset(assets, "linux", "arm64", false)
	if gotArm == nil || gotArm.Name != "beszel-agent_linux_arm64.tar.gz" {
		t.Fatalf("expected arm64 asset, got %+v", gotArm)
	}
}

func TestRewriteDownloadURL(t *testing.T) {
	u := "https://github.com/henrygd/beszel/releases/download/v0.21.2/beszel-agent_linux_amd64.tar.gz"
	if got := RewriteDownloadURL(u, "mirror", ""); got != "https://gh.beszel.dev/henrygd/beszel/releases/download/v0.21.2/beszel-agent_linux_amd64.tar.gz" {
		t.Fatalf("unexpected mirror URL: %s", got)
	}
	if got := RewriteDownloadURL(u, "custom", "https://ghfast.top"); got != "https://ghfast.top/"+u {
		t.Fatalf("unexpected custom URL: %s", got)
	}
}

func TestExtractBinaryFromTarGzAndRaw(t *testing.T) {
	dir := t.TempDir()

	// Build a fake ELF payload inside tar.gz
	payload := append([]byte{0x7f, 'E', 'L', 'F'}, []byte("dummy-elf-content")...)
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	_ = tw.WriteHeader(&tar.Header{Name: "README.md", Mode: 0644, Size: 2})
	_, _ = tw.Write([]byte("hi"))
	_ = tw.WriteHeader(&tar.Header{Name: "beszel-agent", Mode: 0755, Size: int64(len(payload))})
	_, _ = tw.Write(payload)
	_ = tw.Close()
	_ = gw.Close()

	tgzPath := filepath.Join(dir, "agent.tar.gz")
	if err := os.WriteFile(tgzPath, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(dir, "out-bin")
	if err := ExtractBinary(tgzPath, "agent.tar.gz", outPath); err != nil {
		t.Fatalf("ExtractBinary tar.gz failed: %v", err)
	}
	got, err := os.ReadFile(outPath)
	if err != nil || !bytes.Equal(got, payload) {
		t.Fatalf("extracted payload mismatch")
	}

	rawPath := filepath.Join(dir, "raw-agent")
	if err := os.WriteFile(rawPath, payload, 0644); err != nil {
		t.Fatal(err)
	}
	outRaw := filepath.Join(dir, "out-raw")
	if err := ExtractBinary(rawPath, "beszel-agent", outRaw); err != nil {
		t.Fatalf("ExtractBinary raw ELF failed: %v", err)
	}
	gotRaw, err := os.ReadFile(outRaw)
	if err != nil || !bytes.Equal(gotRaw, payload) {
		t.Fatalf("raw ELF payload mismatch")
	}
}

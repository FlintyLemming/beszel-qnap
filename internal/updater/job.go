package updater

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"beszel-qnap/internal/agentctl"
)

type JobRequest struct {
	Tag          string `json:"tag"`        // "latest" or specific tag e.g. "v0.21.2"
	CustomURL    string `json:"custom_url"` // optional direct download URL
	Source       string `json:"source"`     // optional override: "github", "mirror", "custom"
	CustomPrefix string `json:"custom_prefix"`
	PreferGlibc  *bool  `json:"prefer_glibc,omitempty"`
}

type JobStatus struct {
	ID              string `json:"id"`
	State           string `json:"state"` // "idle", "running", "done", "error"
	Stage           string `json:"stage"` // "fetching", "downloading", "verifying", "extracting", "installing", "done", "error"
	Tag             string `json:"tag,omitempty"`
	AssetName       string `json:"asset_name,omitempty"`
	DownloadURL     string `json:"download_url,omitempty"`
	BytesDownloaded int64  `json:"bytes_downloaded"`
	TotalBytes      int64  `json:"total_bytes"`
	Percent         int    `json:"percent"`
	InstalledVer    string `json:"installed_version,omitempty"`
	Message         string `json:"message,omitempty"`
	Error           string `json:"error,omitempty"`
	StartedAt       string `json:"started_at,omitempty"`
	UpdatedAt       string `json:"updated_at,omitempty"`
}

func ReadJobStatus(mgr *agentctl.Manager) *JobStatus {
	data, err := os.ReadFile(mgr.Paths.DownloadStatusPath())
	if err != nil {
		return &JobStatus{State: "idle", Stage: "idle"}
	}
	var st JobStatus
	if err := json.Unmarshal(data, &st); err != nil {
		return &JobStatus{State: "idle", Stage: "idle"}
	}
	return &st
}

func writeJobStatus(mgr *agentctl.Manager, st *JobStatus) {
	st.UpdatedAt = time.Now().Format(time.RFC3339)
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return
	}
	dir := filepath.Dir(mgr.Paths.DownloadStatusPath())
	_ = os.MkdirAll(dir, 0755)
	tmp, err := os.CreateTemp(dir, ".job-*")
	if err != nil {
		return
	}
	_, _ = tmp.Write(append(data, '\n'))
	_ = tmp.Close()
	_ = os.Rename(tmp.Name(), mgr.Paths.DownloadStatusPath())
}

// SpawnDownloadJob launches a detached worker process so CGI requests return immediately.
func SpawnDownloadJob(mgr *agentctl.Manager, req JobRequest) (*JobStatus, error) {
	if err := mgr.Paths.EnsureDirs(); err != nil {
		return nil, err
	}
	cur := ReadJobStatus(mgr)
	if cur.State == "running" {
		if t, err := time.Parse(time.RFC3339, cur.UpdatedAt); err == nil && time.Since(t) < 5*time.Minute {
			return nil, errors.New("an update download is already in progress")
		}
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	reqPath := filepath.Join(mgr.Paths.RunDir(), "download-req.json")
	if err := os.WriteFile(reqPath, reqBytes, 0600); err != nil {
		return nil, err
	}

	jobID := fmt.Sprintf("upd-%d", time.Now().UnixNano())
	initial := &JobStatus{
		ID:        jobID,
		State:     "running",
		Stage:     "fetching",
		Tag:       req.Tag,
		Message:   "Starting update job...",
		StartedAt: time.Now().Format(time.RFC3339),
	}
	writeJobStatus(mgr, initial)

	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(exe, "ctl", "download-job", reqPath, jobID)
	cmd.Dir = mgr.Paths.Root
	cmd.Stdin = nil
	logFile, _ := os.OpenFile(mgr.Paths.ManagerLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if logFile != nil {
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}
	cmd.Env = append(os.Environ(), "BESZEL_QPKG_ROOT="+mgr.Paths.Root)
	agentctl.DetachCmd(cmd)
	if err := cmd.Start(); err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		initial.State = "error"
		initial.Stage = "error"
		initial.Error = err.Error()
		writeJobStatus(mgr, initial)
		return nil, err
	}
	_ = cmd.Process.Release()
	if logFile != nil {
		_ = logFile.Close()
	}
	return initial, nil
}

// RunDownloadJob executes the full download, verification, extraction, and swap flow.
func RunDownloadJob(mgr *agentctl.Manager, reqPath, jobID string) error {
	raw, err := os.ReadFile(reqPath)
	if err != nil {
		return err
	}
	_ = os.Remove(reqPath)

	var req JobRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return err
	}

	st := &JobStatus{
		ID:        jobID,
		State:     "running",
		Stage:     "fetching",
		Tag:       req.Tag,
		StartedAt: time.Now().Format(time.RFC3339),
	}
	fail := func(err error) error {
		st.State = "error"
		st.Stage = "error"
		st.Error = err.Error()
		st.Message = err.Error()
		writeJobStatus(mgr, st)
		mgr.LogManager("Update failed: %v", err)
		return err
	}

	cfg, err := mgr.LoadConfig()
	if err != nil {
		return fail(err)
	}

	source := req.Source
	if source == "" {
		source = cfg.DownloadSource
	}
	customPrefix := req.CustomPrefix
	if customPrefix == "" {
		customPrefix = cfg.CustomMirrorPrefix
	}
	preferGlibc := cfg.PreferGlibc
	if req.PreferGlibc != nil {
		preferGlibc = *req.PreferGlibc
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var downloadURL, assetName, expectedDigest, version string
	var totalBytes int64

	if u := strings.TrimSpace(req.CustomURL); u != "" {
		downloadURL = u
		assetName = filepath.Base(strings.SplitN(u, "?", 2)[0])
		if assetName == "" || assetName == "." || assetName == "/" {
			assetName = "beszel-agent-custom"
		}
		source = "custom_url"
	} else {
		st.Message = "Fetching release metadata..."
		writeJobStatus(mgr, st)
		rel, err := FetchReleaseByTag(ctx, source, customPrefix, cfg.HTTPProxy, req.Tag, preferGlibc)
		if err != nil {
			return fail(err)
		}
		downloadURL = rel.Asset.DownloadURL
		assetName = rel.Asset.Name
		expectedDigest = rel.Asset.Digest
		totalBytes = rel.Asset.Size
		version = rel.Version
		st.Tag = rel.TagName
	}

	st.Stage = "downloading"
	st.AssetName = assetName
	st.DownloadURL = downloadURL
	st.TotalBytes = totalBytes
	st.Message = "Downloading " + assetName + "..."
	writeJobStatus(mgr, st)

	tmpDir, err := os.MkdirTemp(mgr.Paths.RunDir(), "dl-*")
	if err != nil {
		return fail(err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, assetName)
	client, err := NewHTTPClient(cfg.HTTPProxy, 10*time.Minute)
	if err != nil {
		return fail(err)
	}
	if err := downloadWithProgress(ctx, client, downloadURL, archivePath, func(written, total int64) {
		st.BytesDownloaded = written
		if total > 0 {
			st.TotalBytes = total
			st.Percent = int((written * 100) / total)
		}
		writeJobStatus(mgr, st)
	}); err != nil {
		return fail(err)
	}

	if expectedDigest != "" {
		st.Stage = "verifying"
		st.Message = "Verifying SHA-256 checksum..."
		writeJobStatus(mgr, st)
		if err := VerifySHA256(archivePath, expectedDigest); err != nil {
			return fail(err)
		}
	}

	st.Stage = "extracting"
	st.Message = "Extracting beszel-agent..."
	writeJobStatus(mgr, st)

	candidatePath := filepath.Join(mgr.Paths.BinDir(), ".beszel-agent.new")
	_ = os.Remove(candidatePath)
	if err := ExtractBinary(archivePath, assetName, candidatePath); err != nil {
		return fail(err)
	}
	defer os.Remove(candidatePath)

	st.Stage = "installing"
	st.Message = "Installing and verifying binary..."
	writeJobStatus(mgr, st)

	meta := &agentctl.BinaryMeta{
		Version:     version,
		Source:      source,
		AssetName:   assetName,
		DownloadURL: downloadURL,
	}
	if err := InstallPreparedBinary(mgr, candidatePath, meta, true); err != nil {
		return fail(err)
	}

	st.State = "done"
	st.Stage = "done"
	st.Percent = 100
	st.InstalledVer = meta.Version
	st.Message = fmt.Sprintf("Installed beszel-agent %s", meta.Version)
	writeJobStatus(mgr, st)
	return nil
}

func downloadWithProgress(ctx context.Context, client *http.Client, urlStr, destPath string, onProgress func(written, total int64)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download failed with HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	total := resp.ContentLength
	var written int64
	buf := make([]byte, 64*1024)
	lastUpdate := time.Now()
	for {
		n, rerr := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := out.Write(buf[:n]); werr != nil {
				return werr
			}
			written += int64(n)
			if time.Since(lastUpdate) >= 250*time.Millisecond {
				onProgress(written, total)
				lastUpdate = time.Now()
			}
		}
		if errors.Is(rerr, io.EOF) {
			break
		}
		if rerr != nil {
			return rerr
		}
	}
	onProgress(written, total)
	return out.Close()
}

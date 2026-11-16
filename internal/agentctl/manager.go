package agentctl

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"beszel-qnap/internal/config"
)

type BinaryMeta struct {
	Version     string `json:"version"`
	Source      string `json:"source"` // "github", "mirror", "custom", "upload", "rollback"
	AssetName   string `json:"asset_name,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
	UpdatedAt   string `json:"updated_at"`
}

type Status struct {
	Running         bool        `json:"running"`
	PID             int         `json:"pid,omitempty"`
	StartedAt       string      `json:"started_at,omitempty"`
	UptimeSec       int64       `json:"uptime_sec,omitempty"`
	MemoryRSSBytes  int64       `json:"memory_rss_bytes,omitempty"`
	BinaryInstalled bool        `json:"binary_installed"`
	BinaryVersion   string      `json:"binary_version,omitempty"`
	BinarySize      int64       `json:"binary_size,omitempty"`
	BinarySHA256    string      `json:"binary_sha256,omitempty"`
	HasBackup       bool        `json:"has_backup"`
	BackupVersion   string      `json:"backup_version,omitempty"`
	BinaryMeta      *BinaryMeta `json:"binary_meta,omitempty"`
	Configured      bool        `json:"configured"`
	Fingerprint     string      `json:"fingerprint,omitempty"`
	HealthOK        bool        `json:"health_ok"`
	HealthMsg       string      `json:"health_msg,omitempty"`
	Host            HostInfo    `json:"host"`
	QPKGVersion     string      `json:"qpkg_version"`
	InstallPath     string      `json:"install_path"`
}

type Manager struct {
	Paths *Paths
}

func NewManager(paths *Paths) *Manager {
	return &Manager{Paths: paths}
}

func (m *Manager) LogManager(format string, args ...any) {
	_ = m.Paths.EnsureDirs()
	msg := fmt.Sprintf("[%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	f, err := os.OpenFile(m.Paths.ManagerLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(msg)
	_ = f.Close()
}

func (m *Manager) LoadConfig() (*config.Config, error) {
	return config.Load(m.Paths.ConfigPath())
}

func (m *Manager) SaveConfig(cfg *config.Config) error {
	if err := m.Paths.EnsureDirs(); err != nil {
		return err
	}
	if err := cfg.Save(m.Paths.ConfigPath()); err != nil {
		return err
	}
	return cfg.WriteEnvFile(m.Paths.EnvFilePath(), m.Paths.AgentDataDir())
}

func (m *Manager) LoadBinaryMeta() *BinaryMeta {
	data, err := os.ReadFile(m.Paths.BinaryMetaPath())
	if err != nil {
		return nil
	}
	var meta BinaryMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		return nil
	}
	return &meta
}

func (m *Manager) SaveBinaryMeta(meta *BinaryMeta) error {
	data, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.Paths.BinaryMetaPath(), append(data, '\n'), 0644)
}

func (m *Manager) Status(includeFingerprint bool) (*Status, error) {
	_ = m.Paths.EnsureDirs()
	cfg, err := m.LoadConfig()
	if err != nil {
		cfg = config.Default()
	}

	st := &Status{
		Configured:  cfg.IsConfigured(),
		Host:        CollectHostInfoWithPaths(m.Paths),
		QPKGVersion: readCfg(QPKGName, "Version", "/etc/config/qpkg.conf"),
		InstallPath: m.Paths.Root,
	}

	binPath := m.Paths.BinaryPath()
	if fi, err := os.Stat(binPath); err == nil && !fi.IsDir() {
		st.BinaryInstalled = true
		st.BinarySize = fi.Size()
		st.BinaryMeta = m.LoadBinaryMeta()
		if ver, err := m.ProbeBinaryVersion(binPath); err == nil && ver != "" {
			st.BinaryVersion = ver
		} else if st.BinaryMeta != nil && st.BinaryMeta.Version != "" {
			st.BinaryVersion = st.BinaryMeta.Version
		}
		if sum, err := FileSHA256(binPath); err == nil {
			st.BinarySHA256 = sum
		}
	}

	bakPath := m.Paths.BackupBinaryPath()
	if fi, err := os.Stat(bakPath); err == nil && !fi.IsDir() && fi.Size() > 0 {
		st.HasBackup = true
		if ver, err := m.ProbeBinaryVersion(bakPath); err == nil {
			st.BackupVersion = ver
		}
	}

	if pid := m.FindRunningPID(); pid > 0 {
		st.Running = true
		st.PID = pid
		st.MemoryRSSBytes = readProcRSS(pid)
		if started, uptime := readProcTiming(pid); uptime >= 0 {
			st.StartedAt = started
			st.UptimeSec = uptime
		}
		if st.BinaryInstalled {
			ok, msg := m.checkHealth(cfg, st.Host.ExtraPathDirs)
			st.HealthOK = ok
			st.HealthMsg = msg
		}
	}

	if includeFingerprint && st.BinaryInstalled {
		if fp, err := m.GetFingerprint(cfg); err == nil {
			st.Fingerprint = fp
		}
	}

	return st, nil
}

func (m *Manager) ProbeBinaryVersion(binPath string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	for _, flag := range []string{"-v", "--version", "version"} {
		cmd := exec.CommandContext(ctx, binPath, flag)
		out, err := cmd.CombinedOutput()
		if err != nil && len(out) == 0 {
			continue
		}
		s := strings.TrimSpace(string(out))
		// Expected format: "beszel-agent 0.21.2" or "v0.21.2"
		fields := strings.Fields(s)
		if len(fields) >= 2 && strings.Contains(fields[0], "beszel") {
			return strings.TrimPrefix(fields[1], "v"), nil
		}
		if len(fields) == 1 && fields[0] != "" {
			return strings.TrimPrefix(fields[0], "v"), nil
		}
	}
	return "", errors.New("could not determine binary version")
}

func (m *Manager) checkHealth(cfg *config.Config, extraPathDirs []string) (bool, string) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, m.Paths.BinaryPath(), "health")
	cmd.Env = cfg.BuildProcessEnv(os.Environ(), m.Paths.AgentDataDir(), extraPathDirs)
	out, err := cmd.CombinedOutput()
	msg := strings.TrimSpace(string(out))
	if err != nil {
		if msg == "" {
			msg = err.Error()
		}
		return false, msg
	}
	if msg == "" {
		msg = "ok"
	}
	return true, msg
}

func (m *Manager) GetFingerprint(cfg *config.Config) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, m.Paths.BinaryPath(), "fingerprint", "view")
	cmd.Env = cfg.BuildProcessEnv(os.Environ(), m.Paths.AgentDataDir(), nil)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	return strings.TrimSpace(string(out)), nil
}

func (m *Manager) ResetFingerprint() (string, error) {
	cfg, err := m.LoadConfig()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(m.Paths.BinaryPath()); err != nil {
		return "", errors.New("beszel-agent binary is not installed")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, m.Paths.BinaryPath(), "fingerprint", "reset")
	cmd.Env = cfg.BuildProcessEnv(os.Environ(), m.Paths.AgentDataDir(), nil)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%v: %s", err, strings.TrimSpace(string(out)))
	}
	m.LogManager("Fingerprint reset: %s", strings.TrimSpace(string(out)))
	return strings.TrimSpace(string(out)), nil
}

func (m *Manager) Start() error {
	if err := m.Paths.EnsureDirs(); err != nil {
		return err
	}
	binPath := m.Paths.BinaryPath()
	if fi, err := os.Stat(binPath); err != nil || fi.IsDir() {
		return errors.New("beszel-agent binary is not installed; please download or upload it first")
	}
	_ = os.Chmod(binPath, 0755)

	cfg, err := m.LoadConfig()
	if err != nil {
		return err
	}
	if !cfg.IsConfigured() {
		return errors.New("beszel-agent is not configured: set KEY (SSH public key) or HUB_URL + TOKEN first")
	}

	if pid := m.FindRunningPID(); pid > 0 {
		return nil
	}

	if err := cfg.WriteEnvFile(m.Paths.EnvFilePath(), m.Paths.AgentDataDir()); err != nil {
		m.LogManager("Warning: failed writing env snapshot: %v", err)
	}

	rotateLogIfLarge(m.Paths.AgentLogPath(), 5*1024*1024)
	logFile, err := os.OpenFile(m.Paths.AgentLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open agent log file: %w", err)
	}
	defer logFile.Close()

	_, _ = fmt.Fprintf(logFile, "\n--- Starting beszel-agent at %s ---\n", time.Now().Format(time.RFC3339))

	args, err := splitArgs(cfg.ExtraArgs)
	if err != nil {
		return fmt.Errorf("invalid extra_args: %w", err)
	}

	cmd := exec.Command(binPath, args...)
	cmd.Dir = m.Paths.AgentDataDir()
	cmd.Stdin = nil
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Env = cfg.BuildProcessEnv(os.Environ(), m.Paths.AgentDataDir(), detectExtraPathDirs(m.Paths))
	DetachCmd(cmd)

	if err := cmd.Start(); err != nil {
		m.LogManager("Failed to start beszel-agent: %v", err)
		return fmt.Errorf("failed to start beszel-agent: %w", err)
	}
	pid := cmd.Process.Pid
	_ = cmd.Process.Release()
	_ = os.WriteFile(m.Paths.PIDFilePath(), []byte(strconv.Itoa(pid)+"\n"), 0644)

	// Wait briefly to detect immediate startup failures (invalid key, port in use, etc.).
	deadline := time.Now().Add(600 * time.Millisecond)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			_ = os.Remove(m.Paths.PIDFilePath())
			tail, _ := m.ReadLogs("agent", 15)
			m.LogManager("beszel-agent exited immediately (pid=%d)", pid)
			return fmt.Errorf("beszel-agent exited immediately after start:\n%s", strings.Join(tail, "\n"))
		}
		time.Sleep(150 * time.Millisecond)
	}

	m.LogManager("Started beszel-agent (pid=%d)", pid)
	return nil
}

func (m *Manager) Stop() error {
	pid := m.FindRunningPID()
	if pid <= 0 {
		_ = os.Remove(m.Paths.PIDFilePath())
		return nil
	}

	if err := signalTerm(pid); err != nil {
		_ = os.Remove(m.Paths.PIDFilePath())
		return nil
	}

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			_ = os.Remove(m.Paths.PIDFilePath())
			m.LogManager("Stopped beszel-agent (pid=%d)", pid)
			return nil
		}
		time.Sleep(150 * time.Millisecond)
	}

	_ = signalKill(pid)
	time.Sleep(300 * time.Millisecond)
	_ = os.Remove(m.Paths.PIDFilePath())
	m.LogManager("Force-killed beszel-agent (pid=%d)", pid)
	return nil
}

func (m *Manager) Restart() error {
	if err := m.Stop(); err != nil {
		return err
	}
	return m.Start()
}

// StartIfConfigured is called by BeszelAgent.sh on QPKG startup; never blocks QPKG enablement if unconfigured.
func (m *Manager) StartIfConfigured() error {
	cfg, err := m.LoadConfig()
	if err != nil {
		m.LogManager("Config load warning during start: %v", err)
		return nil
	}
	if !cfg.AutoStart {
		m.LogManager("auto_start disabled; skipping agent launch")
		return nil
	}
	if _, err := os.Stat(m.Paths.BinaryPath()); err != nil {
		m.LogManager("beszel-agent binary not installed yet; waiting for Web UI setup")
		return nil
	}
	if !cfg.IsConfigured() {
		m.LogManager("beszel-agent not configured yet; waiting for Web UI setup")
		return nil
	}
	if err := m.Start(); err != nil {
		m.LogManager("StartIfConfigured error: %v", err)
	}
	return nil
}

func (m *Manager) FindRunningPID() int {
	binPath := filepath.Clean(m.Paths.BinaryPath())
	if raw, err := os.ReadFile(m.Paths.PIDFilePath()); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(raw))); err == nil && pid > 0 {
			if m.matchesAgentProcess(pid, binPath) {
				return pid
			}
		}
		_ = os.Remove(m.Paths.PIDFilePath())
	}

	// Fallback scan /proc in case PID file is missing.
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil || pid <= 0 {
			continue
		}
		if m.matchesAgentProcess(pid, binPath) {
			_ = os.WriteFile(m.Paths.PIDFilePath(), []byte(strconv.Itoa(pid)+"\n"), 0644)
			return pid
		}
	}
	return 0
}

func (m *Manager) matchesAgentProcess(pid int, binPath string) bool {
	if !processAlive(pid) {
		return false
	}
	if exe, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid)); err == nil {
		cleanExe := strings.TrimSuffix(filepath.Clean(exe), " (deleted)")
		if cleanExe == binPath {
			return true
		}
	}
	if cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid)); err == nil {
		parts := bytes.Split(cmdline, []byte{0})
		if len(parts) > 0 {
			arg0 := strings.TrimSpace(string(parts[0]))
			if filepath.Clean(arg0) == binPath {
				return true
			}
		}
	}
	return false
}

func readProcRSS(pid int) int64 {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return 0
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "VmRSS:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				kb, _ := strconv.ParseInt(fields[1], 10, 64)
				return kb * 1024
			}
		}
	}
	return 0
}

func readProcTiming(pid int) (string, int64) {
	fi, err := os.Stat(fmt.Sprintf("/proc/%d", pid))
	if err != nil {
		return "", -1
	}
	started := fi.ModTime()
	uptime := int64(time.Since(started).Seconds())
	if uptime < 0 {
		uptime = 0
	}
	return started.Format(time.RFC3339), uptime
}

func (m *Manager) ReadLogs(kind string, maxLines int) ([]string, error) {
	if maxLines <= 0 {
		maxLines = 200
	}
	if maxLines > 2000 {
		maxLines = 2000
	}
	path := m.Paths.AgentLogPath()
	if kind == "manager" {
		path = m.Paths.ManagerLogPath()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []string{}, nil
		}
		return nil, err
	}
	s := strings.ReplaceAll(string(data), "\r\n", "\n")
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return []string{}, nil
	}
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return lines, nil
}

func (m *Manager) ClearLogs(kind string) error {
	path := m.Paths.AgentLogPath()
	if kind == "manager" {
		path = m.Paths.ManagerLogPath()
	}
	return os.WriteFile(path, []byte{}, 0644)
}

func rotateLogIfLarge(path string, maxBytes int64) {
	fi, err := os.Stat(path)
	if err != nil || fi.Size() < maxBytes {
		return
	}
	_ = os.Rename(path, path+".1")
}

func FileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func splitArgs(s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	var args []string
	var cur strings.Builder
	var quote rune
	escaped := false

	for _, r := range s {
		switch {
		case escaped:
			cur.WriteRune(r)
			escaped = false
		case r == '\\' && quote != '\'':
			escaped = true
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '"' || r == '\'':
			quote = r
		case r == ' ' || r == '\t' || r == '\n':
			if cur.Len() > 0 {
				args = append(args, cur.String())
				cur.Reset()
			}
		default:
			cur.WriteRune(r)
		}
	}
	if escaped || quote != 0 {
		return nil, errors.New("unterminated quote or escape in extra_args")
	}
	if cur.Len() > 0 {
		args = append(args, cur.String())
	}
	return args, nil
}

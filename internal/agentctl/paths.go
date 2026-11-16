package agentctl

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const QPKGName = "BeszelAgent"

type Paths struct {
	Root string
}

func ResolvePaths() (*Paths, error) {
	if root := strings.TrimSpace(os.Getenv("BESZEL_QPKG_ROOT")); root != "" {
		return &Paths{Root: filepath.Clean(root)}, nil
	}

	if _, err := os.Stat("/sbin/getcfg"); err == nil {
		if out, err := exec.Command("/sbin/getcfg", QPKGName, "Install_Path", "-f", "/etc/config/qpkg.conf").Output(); err == nil {
			if p := strings.TrimSpace(string(out)); p != "" {
				return &Paths{Root: filepath.Clean(p)}, nil
			}
		}
	}

	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	exe, _ = filepath.EvalSymlinks(exe)
	dir := filepath.Dir(exe)
	// When invoked as <Root>/www/api.cgi or <Root>/bin/beszel-qnap, parent is Root.
	base := filepath.Base(dir)
	if base == "www" || base == "bin" {
		return &Paths{Root: filepath.Dir(dir)}, nil
	}
	return &Paths{Root: dir}, nil
}

func (p *Paths) DataDir() string             { return filepath.Join(p.Root, "data") }
func (p *Paths) BundledBinDir() string       { return filepath.Join(p.Root, "bin") }
func (p *Paths) BundledSmartctlPath() string { return filepath.Join(p.BundledBinDir(), "smartctl") }
func (p *Paths) ConfigPath() string         { return filepath.Join(p.DataDir(), "config.json") }
func (p *Paths) EnvFilePath() string        { return filepath.Join(p.DataDir(), "beszel-agent.env") }
func (p *Paths) BinDir() string             { return filepath.Join(p.DataDir(), "bin") }
func (p *Paths) BinaryPath() string         { return filepath.Join(p.BinDir(), "beszel-agent") }
func (p *Paths) BackupBinaryPath() string   { return filepath.Join(p.BinDir(), "beszel-agent.bak") }
func (p *Paths) BinaryMetaPath() string     { return filepath.Join(p.BinDir(), "meta.json") }
func (p *Paths) AgentDataDir() string       { return filepath.Join(p.DataDir(), "agent-data") }
func (p *Paths) LogsDir() string            { return filepath.Join(p.DataDir(), "logs") }
func (p *Paths) AgentLogPath() string       { return filepath.Join(p.LogsDir(), "beszel-agent.log") }
func (p *Paths) ManagerLogPath() string     { return filepath.Join(p.LogsDir(), "manager.log") }
func (p *Paths) RunDir() string             { return filepath.Join(p.DataDir(), "run") }
func (p *Paths) PIDFilePath() string        { return filepath.Join(p.RunDir(), "beszel-agent.pid") }
func (p *Paths) DownloadStatusPath() string { return filepath.Join(p.RunDir(), "download.json") }
func (p *Paths) LockPath() string           { return filepath.Join(p.RunDir(), "manager.lock") }

func (p *Paths) EnsureDirs() error {
	for _, d := range []string{p.DataDir(), p.BinDir(), p.AgentDataDir(), p.LogsDir(), p.RunDir()} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

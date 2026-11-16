package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type EnvPair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Config stores QPKG runtime settings and beszel-agent environment configuration.
type Config struct {
	// QPKG runtime settings
	AutoStart          bool   `json:"auto_start"`
	DownloadSource     string `json:"download_source"`      // "github", "mirror", "custom"
	CustomMirrorPrefix string `json:"custom_mirror_prefix"` // e.g. https://ghfast.top/
	PreferGlibc        bool   `json:"prefer_glibc"`
	HTTPProxy          string `json:"http_proxy"`
	ExtraArgs          string `json:"extra_args"`

	// Beszel Agent - Connection
	Listen     string `json:"listen"`
	Key        string `json:"key"`
	HubURL     string `json:"hub_url"`
	Token      string `json:"token"`
	DisableSSH bool   `json:"disable_ssh"`
	SystemName string `json:"system_name"`
	AllProxy   string `json:"all_proxy"`
	CACertFile string `json:"ca_cert_file"`
	HubHeaders string `json:"hub_headers"`

	// Beszel Agent - Storage & Filesystems
	Filesystem       string `json:"filesystem"`
	ExtraFilesystems string `json:"extra_filesystems"`
	DiskUsageCache   string `json:"disk_usage_cache"`

	// Beszel Agent - Docker / Container Station
	DockerHost         string `json:"docker_host"`
	DockerTimeout      string `json:"docker_timeout"`
	DockerImageCheck   string `json:"docker_image_check"`
	ExcludeContainers  string `json:"exclude_containers"`
	ContainerNameLabel string `json:"container_name_label"`

	// Beszel Agent - SMART & Sensors & GPU
	SmartDevices   string `json:"smart_devices"`
	ExcludeSmart   string `json:"exclude_smart"`
	SmartInterval  string `json:"smart_interval"`
	Sensors        string `json:"sensors"`
	PrimarySensor  string `json:"primary_sensor"`
	FanSensors     string `json:"fan_sensors"`
	SensorsTimeout string `json:"sensors_timeout"`
	SkipGPU        bool   `json:"skip_gpu"`
	GPUCollector   string `json:"gpu_collector"`

	// Beszel Agent - System & Logging
	NICs        string `json:"nics"`
	MemCalc     string `json:"mem_calc"`
	LogLevel    string `json:"log_level"`
	SkipSystemd bool   `json:"skip_systemd"`
	SkipWifi    bool   `json:"skip_wifi"`

	// Arbitrary extra environment variables for forward compatibility
	CustomEnv []EnvPair `json:"custom_env"`
}

func Default() *Config {
	return &Config{
		AutoStart:      true,
		DownloadSource: "github",
		Listen:         "45876",
		LogLevel:       "info",
		SkipSystemd:    true,
		CustomEnv:      []EnvPair{},
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return nil, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return cfg, nil
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid config JSON: %w", err)
	}
	cfg.normalize()
	return cfg, nil
}

func (c *Config) normalize() {
	if c.DownloadSource == "" {
		c.DownloadSource = "github"
	}
	if c.LogLevel == "" {
		c.LogLevel = "info"
	}
	if c.CustomEnv == nil {
		c.CustomEnv = []EnvPair{}
	}
	filtered := make([]EnvPair, 0, len(c.CustomEnv))
	for _, pair := range c.CustomEnv {
		k := strings.TrimSpace(pair.Key)
		if k == "" {
			continue
		}
		filtered = append(filtered, EnvPair{Key: k, Value: pair.Value})
	}
	c.CustomEnv = filtered
}

func (c *Config) Save(path string) error {
	c.normalize()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeAtomic(path, data, 0600)
}

func (c *Config) IsConfigured() bool {
	if strings.TrimSpace(c.Key) != "" {
		return true
	}
	if strings.TrimSpace(c.HubURL) != "" && strings.TrimSpace(c.Token) != "" {
		return true
	}
	env := c.AgentEnvMap("")
	if strings.TrimSpace(env["KEY"]) != "" || strings.TrimSpace(env["KEY_FILE"]) != "" {
		return true
	}
	if strings.TrimSpace(env["HUB_URL"]) != "" && (strings.TrimSpace(env["TOKEN"]) != "" || strings.TrimSpace(env["TOKEN_FILE"]) != "") {
		return true
	}
	return false
}

// AgentEnvMap returns the key/value environment variables passed to beszel-agent.
func (c *Config) AgentEnvMap(agentDataDir string) map[string]string {
	env := make(map[string]string)
	setIfNotEmpty := func(k, v string) {
		v = strings.TrimSpace(v)
		if v != "" {
			env[k] = v
		}
	}

	if agentDataDir != "" {
		env["DATA_DIR"] = agentDataDir
	}

	setIfNotEmpty("LISTEN", c.Listen)
	setIfNotEmpty("KEY", c.Key)
	setIfNotEmpty("HUB_URL", c.HubURL)
	setIfNotEmpty("TOKEN", c.Token)
	if c.DisableSSH {
		env["DISABLE_SSH"] = "true"
	}
	setIfNotEmpty("SYSTEM_NAME", c.SystemName)
	setIfNotEmpty("BESZEL_AGENT_ALL_PROXY", c.AllProxy)
	setIfNotEmpty("CA_CERT_FILE", c.CACertFile)
	setIfNotEmpty("HUB_HEADERS", c.HubHeaders)

	setIfNotEmpty("FILESYSTEM", c.Filesystem)
	setIfNotEmpty("EXTRA_FILESYSTEMS", c.ExtraFilesystems)
	setIfNotEmpty("DISK_USAGE_CACHE", c.DiskUsageCache)

	setIfNotEmpty("DOCKER_HOST", c.DockerHost)
	setIfNotEmpty("DOCKER_TIMEOUT", c.DockerTimeout)
	setIfNotEmpty("DOCKER_IMAGE_CHECK", c.DockerImageCheck)
	setIfNotEmpty("EXCLUDE_CONTAINERS", c.ExcludeContainers)
	setIfNotEmpty("CONTAINER_NAME_LABEL", c.ContainerNameLabel)

	setIfNotEmpty("SMART_DEVICES", c.SmartDevices)
	setIfNotEmpty("EXCLUDE_SMART", c.ExcludeSmart)
	setIfNotEmpty("SMART_INTERVAL", c.SmartInterval)
	setIfNotEmpty("SENSORS", c.Sensors)
	setIfNotEmpty("PRIMARY_SENSOR", c.PrimarySensor)
	setIfNotEmpty("FAN_SENSORS", c.FanSensors)
	setIfNotEmpty("SENSORS_TIMEOUT", c.SensorsTimeout)
	if c.SkipGPU {
		env["SKIP_GPU"] = "true"
	}
	setIfNotEmpty("GPU_COLLECTOR", c.GPUCollector)

	setIfNotEmpty("NICS", c.NICs)
	setIfNotEmpty("MEM_CALC", c.MemCalc)
	setIfNotEmpty("LOG_LEVEL", c.LogLevel)
	if c.SkipSystemd {
		env["SKIP_SYSTEMD"] = "true"
	}
	if c.SkipWifi {
		env["SKIP_WIFI"] = "true"
	}

	for _, pair := range c.CustomEnv {
		k := strings.TrimSpace(pair.Key)
		if k != "" {
			env[k] = pair.Value
		}
	}

	return env
}

// BuildProcessEnv merges baseEnv with AgentEnvMap and extra PATH directories.
func (c *Config) BuildProcessEnv(baseEnv []string, agentDataDir string, extraPathDirs []string) []string {
	merged := make(map[string]string, len(baseEnv)+32)
	for _, entry := range baseEnv {
		if idx := strings.IndexByte(entry, '='); idx > 0 {
			merged[entry[:idx]] = entry[idx+1:]
		}
	}

	if len(extraPathDirs) > 0 {
		curPath := merged["PATH"]
		parts := make([]string, 0, len(extraPathDirs)+1)
		if curPath != "" {
			parts = append(parts, curPath)
		}
		for _, dir := range extraPathDirs {
			if dir != "" && !containsPathSegment(curPath, dir) {
				parts = append(parts, dir)
			}
		}
		merged["PATH"] = strings.Join(parts, string(os.PathListSeparator))
	}

	for k, v := range c.AgentEnvMap(agentDataDir) {
		merged[k] = v
	}

	keys := make([]string, 0, len(merged))
	for k := range merged {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(keys))
	for _, k := range keys {
		out = append(out, k+"="+merged[k])
	}
	return out
}

func containsPathSegment(pathVal, target string) bool {
	for _, seg := range filepath.SplitList(pathVal) {
		if seg == target {
			return true
		}
	}
	return false
}

// WriteEnvFile writes a readable env snapshot to disk for transparency/debugging.
func (c *Config) WriteEnvFile(path, agentDataDir string) error {
	env := c.AgentEnvMap(agentDataDir)
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	sb.WriteString("# Generated by BeszelAgent QPKG\n")
	for _, k := range keys {
		v := strings.ReplaceAll(env[k], "\n", "\\n")
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(v)
		sb.WriteByte('\n')
	}
	return writeAtomic(path, []byte(sb.String()), 0600)
}

func writeAtomic(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Chmod(perm); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

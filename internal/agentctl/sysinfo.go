package agentctl

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

type VolumeInfo struct {
	Path       string `json:"path"`
	TotalBytes uint64 `json:"total_bytes"`
	FreeBytes  uint64 `json:"free_bytes"`
}

type SensorInfo struct {
	Device string `json:"device"`
	Name   string `json:"name"`
}

type HostInfo struct {
	Hostname      string       `json:"hostname"`
	OS            string       `json:"os"`
	Arch          string       `json:"arch"`
	UnameArch     string       `json:"uname_arch"`
	QTSVersion    string       `json:"qts_version"`
	QTSModel      string       `json:"qts_model"`
	HasGlibc      bool         `json:"has_glibc"`
	DockerSocket  string       `json:"docker_socket"`
	SmartctlPath  string       `json:"smartctl_path"`
	Volumes       []VolumeInfo `json:"volumes"`
	HwmonSensors  []SensorInfo `json:"hwmon_sensors"`
	ExtraPathDirs []string     `json:"extra_path_dirs"`
}

func CollectHostInfo() HostInfo {
	return CollectHostInfoWithPaths(nil)
}

// CollectHostInfoWithPaths probes host metadata while prioritizing QPKG-bundled tools.
func CollectHostInfoWithPaths(paths *Paths) HostInfo {
	host, _ := os.Hostname()
	info := HostInfo{
		Hostname:      host,
		OS:            runtime.GOOS,
		Arch:          runtime.GOARCH,
		UnameArch:     readUnameM(),
		QTSVersion:    readCfg("System", "Version", "/etc/config/uLinux.conf"),
		QTSModel:      readCfg("System", "Model", "/etc/config/uLinux.conf"),
		HasGlibc:      detectGlibc(),
		DockerSocket:  detectDockerSocket(),
		SmartctlPath:  detectSmartctl(paths),
		Volumes:       detectStorageVolumes(),
		HwmonSensors:  detectHwmonSensors(),
		ExtraPathDirs: detectExtraPathDirs(paths),
	}
	return info
}

func readUnameM() string {
	if runtime.GOOS != "linux" {
		return runtime.GOARCH
	}
	out, err := exec.Command("uname", "-m").Output()
	if err != nil {
		return runtime.GOARCH
	}
	return strings.TrimSpace(string(out))
}

func readCfg(section, key, file string) string {
	if _, err := os.Stat("/sbin/getcfg"); err != nil {
		return ""
	}
	out, err := exec.Command("/sbin/getcfg", section, key, "-f", file).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func detectGlibc() bool {
	for _, p := range []string{
		"/lib64/ld-linux-x86-64.so.2",
		"/lib/x86_64-linux-gnu/ld-linux-x86-64.so.2",
		"/lib/ld-linux-x86-64.so.2",
		"/lib/ld-linux-aarch64.so.1",
	} {
		if _, err := os.Stat(p); err == nil {
			return true
		}
	}
	return false
}

func detectDockerSocket() string {
	for _, p := range []string{"/var/run/docker.sock", "/run/docker.sock"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func detectExtraPathDirs(paths *Paths) []string {
	candidates := make([]string, 0, 8)
	if paths != nil {
		candidates = append(candidates, paths.BundledBinDir(), paths.BinDir())
	}
	candidates = append(candidates,
		"/opt/sbin",
		"/opt/bin",
		"/usr/local/bin",
		"/usr/local/sbin",
	)
	if matches, err := filepath.Glob("/share/*/.qpkg/container-station/bin"); err == nil {
		candidates = append(candidates, matches...)
	}
	out := make([]string, 0, len(candidates))
	seen := make(map[string]struct{})
	for _, c := range candidates {
		if _, ok := seen[c]; ok {
			continue
		}
		if st, err := os.Stat(c); err == nil && st.IsDir() {
			seen[c] = struct{}{}
			out = append(out, c)
		}
	}
	return out
}

func detectSmartctl(paths *Paths) string {
	candidates := make([]string, 0, 8)
	if paths != nil {
		candidates = append(candidates,
			paths.BundledSmartctlPath(),
			filepath.Join(paths.BinDir(), "smartctl"),
		)
	}
	candidates = append(candidates,
		"/opt/sbin/smartctl",
		"/opt/bin/smartctl",
		"/usr/sbin/smartctl",
		"/sbin/smartctl",
		"/usr/local/sbin/smartctl",
	)
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	if p, err := exec.LookPath("smartctl"); err == nil {
		return p
	}
	return ""
}

func detectStorageVolumes() []VolumeInfo {
	mounted := readMountedSharePaths()
	if len(mounted) > 0 {
		vols := make([]VolumeInfo, 0, len(mounted))
		for _, m := range mounted {
			total, free, ok := statfsDisk(m)
			if !ok || total == 0 {
				continue
			}
			vols = append(vols, VolumeInfo{
				Path:       m,
				TotalBytes: total,
				FreeBytes:  free,
			})
		}
		sort.Slice(vols, func(i, j int) bool {
			return vols[i].Path < vols[j].Path
		})
		return vols
	}

	patterns := []string{
		"/share/CACHEDEV*_DATA",
		"/share/MD*_DATA",
		"/share/ZFS*_DATA",
		"/share/HDA_DATA",
	}
	seen := make(map[string]struct{})
	vols := make([]VolumeInfo, 0)
	for _, pat := range patterns {
		matches, err := filepath.Glob(pat)
		if err != nil {
			continue
		}
		for _, m := range matches {
			if _, ok := seen[m]; ok {
				continue
			}
			st, err := os.Stat(m)
			if err != nil || !st.IsDir() {
				continue
			}
			seen[m] = struct{}{}
			total, free, ok := statfsDisk(m)
			if !ok || total == 0 {
				continue
			}
			vols = append(vols, VolumeInfo{
				Path:       m,
				TotalBytes: total,
				FreeBytes:  free,
			})
		}
	}
	sort.Slice(vols, func(i, j int) bool {
		return vols[i].Path < vols[j].Path
	})
	return vols
}

func readMountedSharePaths() []string {
	data, err := os.ReadFile("/proc/mounts")
	if err != nil {
		return nil
	}
	seen := make(map[string]struct{})
	out := make([]string, 0)
	for _, line := range strings.Split(string(data), "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		mp := fields[1]
		fstype := fields[2]
		if fstype == "tmpfs" || fstype == "overlay" || fstype == "nsfs" {
			continue
		}
		if strings.HasPrefix(mp, "/share/") && strings.HasSuffix(mp, "_DATA") && strings.Count(mp, "/") == 2 {
			if _, ok := seen[mp]; !ok {
				seen[mp] = struct{}{}
				out = append(out, mp)
			}
		}
	}
	return out
}

func detectHwmonSensors() []SensorInfo {
	dirs, err := filepath.Glob("/sys/class/hwmon/hwmon*")
	if err != nil {
		return nil
	}
	sort.Strings(dirs)
	out := make([]SensorInfo, 0, len(dirs))
	for _, d := range dirs {
		nameBytes, err := os.ReadFile(filepath.Join(d, "name"))
		name := strings.TrimSpace(string(nameBytes))
		if err != nil || name == "" {
			name = filepath.Base(d)
		}
		out = append(out, SensorInfo{
			Device: filepath.Base(d),
			Name:   name,
		})
	}
	return out
}

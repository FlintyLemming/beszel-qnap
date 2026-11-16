package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type archSpec struct {
	name           string
	cpuArchRegex   string
	rejectPlatform string
}

var supportedArchs = map[string]archSpec{
	"x86_64":  {name: "x86_64", cpuArchRegex: `x86_64`},
	"arm_64":  {name: "arm_64", cpuArchRegex: `aarch64`},
	"arm-x41": {name: "arm-x41", cpuArchRegex: `armv7l`, rejectPlatform: `ARM_MS`},
}

func main() {
	qpkgDir := flag.String("qpkg-dir", "qpkg", "path to qpkg template directory")
	outDir := flag.String("out-dir", "dist", "output directory for .qpkg files")
	archFlag := flag.String("arch", "x86_64", "target architecture (x86_64, arm_64, arm-x41)")
	flag.Parse()

	spec, ok := supportedArchs[*archFlag]
	if !ok {
		fatalf("unsupported arch %q", *archFlag)
	}

	cfgBytes, err := os.ReadFile(filepath.Join(*qpkgDir, "qpkg.cfg"))
	if err != nil {
		fatalf("read qpkg.cfg: %v", err)
	}
	cfgBytes = normalizeLF(cfgBytes)
	qpkgName := parseCfgValue(string(cfgBytes), "QPKG_NAME")
	qpkgVer := parseCfgValue(string(cfgBytes), "QPKG_VER")
	qpkgDisp := parseCfgValue(string(cfgBytes), "QPKG_DISPLAY_NAME")
	if qpkgDisp == "" {
		qpkgDisp = qpkgName
	}
	if qpkgName == "" || qpkgVer == "" {
		fatalf("QPKG_NAME and QPKG_VER must be set in qpkg.cfg")
	}

	dataTarGz, err := buildDataTarGz(*qpkgDir, spec.name, qpkgName)
	if err != nil {
		fatalf("build data.tar.gz: %v", err)
	}

	ctrlTar, err := buildControlTar(*qpkgDir, cfgBytes)
	if err != nil {
		fatalf("build control.tar: %v", err)
	}

	header := buildQPKGHeader(qpkgName, qpkgDisp, qpkgVer, spec, len(ctrlTar), len(dataTarGz))
	tail := buildQPKGTail(qpkgName, qpkgVer)

	var out bytes.Buffer
	out.Write(header)
	out.Write(ctrlTar)
	out.Write(dataTarGz)
	out.Write(tail)

	raw := out.Bytes()
	applyQPKGEncrypt(raw)

	if err := os.MkdirAll(*outDir, 0755); err != nil {
		fatalf("mkdir out-dir: %v", err)
	}
	outFile := filepath.Join(*outDir, fmt.Sprintf("%s_%s_%s.qpkg", qpkgName, qpkgVer, spec.name))
	if err := os.WriteFile(outFile, raw, 0755); err != nil {
		fatalf("write qpkg: %v", err)
	}
	fmt.Printf("Built %s (%d bytes)\n", outFile, len(raw))
}

func buildDataTarGz(qpkgDir, archName, qpkgName string) ([]byte, error) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	now := time.Now()
	seen := make(map[string]struct{})

	addTree := func(baseDir string) error {
		if _, err := os.Stat(baseDir); errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return filepath.WalkDir(baseDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(baseDir, path)
			if err != nil {
				return err
			}
			rel = filepath.ToSlash(rel)
			if rel == "." {
				return nil
			}
			tarName := "./" + rel
			if d.IsDir() {
				if _, ok := seen[tarName]; ok {
					return nil
				}
				seen[tarName] = struct{}{}
				return tw.WriteHeader(&tar.Header{
					Name:     tarName + "/",
					Mode:     0755,
					Typeflag: tar.TypeDir,
					ModTime:  now,
				})
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			mode := int64(0644)
			if isExecutableAsset(rel, data) {
				mode = 0755
			}
			if isTextAsset(rel) {
				data = normalizeLF(data)
			}
			seen[tarName] = struct{}{}
			return writeTarFile(tw, tarName, data, mode, now)
		})
	}

	if err := addTree(filepath.Join(qpkgDir, "shared")); err != nil {
		return nil, err
	}
	if err := addTree(filepath.Join(qpkgDir, archName)); err != nil {
		return nil, err
	}

	iconMappings := [][2]string{
		{qpkgName + ".gif", ".qpkg_icon.gif"},
		{qpkgName + "_80.gif", ".qpkg_icon_80.gif"},
		{qpkgName + "_gray.gif", ".qpkg_icon_gray.gif"},
	}
	for _, m := range iconMappings {
		p := filepath.Join(qpkgDir, "icons", m[0])
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		if err := writeTarFile(tw, "./"+m[1], data, 0644, now); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildControlTar(qpkgDir string, cfgBytes []byte) ([]byte, error) {
	qinstallBytes, err := os.ReadFile(filepath.Join("scripts", "qinstall.sh"))
	if err != nil {
		return nil, err
	}
	routinesBytes, err := os.ReadFile(filepath.Join(qpkgDir, "package_routines"))
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var gzBuf bytes.Buffer
	gw := gzip.NewWriter(&gzBuf)
	tw := tar.NewWriter(gw)

	builtInfo := []byte("time = " + now.Format("20060102") + "\n")
	files := []struct {
		name string
		data []byte
		mode int64
	}{
		{"./built_info", builtInfo, 0644},
		{"./qinstall.sh", normalizeLF(qinstallBytes), 0755},
		{"./package_routines", normalizeLF(routinesBytes), 0644},
		{"./qpkg.cfg", cfgBytes, 0644},
	}
	for _, f := range files {
		if err := writeTarFile(tw, f.name, f.data, f.mode, now); err != nil {
			return nil, err
		}
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gw.Close(); err != nil {
		return nil, err
	}

	var outer bytes.Buffer
	outerTW := tar.NewWriter(&outer)
	if err := writeTarFile(outerTW, "control.tar.gz", gzBuf.Bytes(), 0644, now); err != nil {
		return nil, err
	}
	if err := outerTW.Close(); err != nil {
		return nil, err
	}
	return outer.Bytes(), nil
}

func buildQPKGHeader(name, displayName, ver string, spec archSpec, ctrlLen, dataLen int) []byte {
	count1k := (dataLen + 1023) / 1024
	tpl := `#!/bin/sh
find_base(){
HDD_MOUNT=` + "`/sbin/getcfg Public path -f /etc/config/smb.conf`" + `
local log_tool="/sbin/log_tool -t2 -uSystem -p127.0.0.1 -mlocalhost -a"

if [ -e "$HDD_MOUNT" ]; then
	if [ -z "$QINSTALL_PATH" ]; then
		BASE_GROUP="/share/HDA_DATA /share/HDB_DATA /share/HDC_DATA /share/HDD_DATA /share/HDE_DATA /share/HDF_DATA /share/HDG_DATA /share/HDH_DATA /share/HDI_DATA /share/HDJ_DATA /share/HDK_DATA /share/HDL_DATA /share/MD0_DATA /share/MD1_DATA /share/MD2_DATA /share/MD3_DATA"
		publicdir=` + "`/sbin/getcfg Public path -f /etc/config/smb.conf`" + `
		if [ ! -z $publicdir ] && [ -d $publicdir ];then
			publicdirp1=` + "`/bin/echo $publicdir | /bin/cut -d \"/\" -f 2`" + `
			publicdirp2=` + "`/bin/echo $publicdir | /bin/cut -d \"/\" -f 3`" + `
			publicdirp3=` + "`/bin/echo $publicdir | /bin/cut -d \"/\" -f 4`" + `
			if [ ! -z $publicdirp1 ] && [ ! -z $publicdirp2 ] && [ ! -z $publicdirp3 ]; then
				[ -d "/${publicdirp1}/${publicdirp2}/Public" ] && QPKG_BASE="/${publicdirp1}/${publicdirp2}"
			fi
		fi
		if [ -z $QPKG_BASE ]; then
			for datadirtest in $BASE_GROUP; do
				[ -d $datadirtest/Public ] && QPKG_BASE="/${publicdirp1}/${publicdirp2}"
			done
		fi
		if [ -z $QPKG_BASE ] ; then
			echo "The Public share not found."
			return 1
		fi
		QPKG_INSTALL_PATH="${QPKG_BASE}/.qpkg"
		QPKG_DIR="${QPKG_INSTALL_PATH}/` + name + `"
	else
		if [ -e "$QINSTALL_PATH" ]; then
			QPKG_INSTALL_PATH="${QINSTALL_PATH}"
			QPKG_DIR="${QINSTALL_PATH}/` + name + `"
		else
			if [ -x "/usr/local/sbin/notify" ]; then
				/usr/local/sbin/notify send -A A039 -C C001 -M 50 -l error -t 3 "[{0}] {1} install failed due to QTS application install volume not found." "QPKG" "` + displayName + `"
				echo -1 > /tmp/update_process && exit 1
			else
				$log_tool "[QPKG] Failed to install ` + displayName + `. The selected installation volume is missing."
			fi
			echo -1 > /tmp/update_process && exit 1
		fi
	fi
	return 0
else
	echo -1 > /tmp/update_process && exit 1
fi
}
wrong_arch(){
	if [ -x "/usr/local/sbin/notify" ]; then
		/usr/local/sbin/notify send -A A039 -C C001 -M 51 -l error -t 3 "[{0}] {1} {2} install failed due to the platform is incompatible. Please use correct package for installation." "QPKG" "` + displayName + `" "` + ver + `"
		echo -1 > /tmp/update_process && exit 1
	else
		$log_tool "[QPKG] Failed to install ` + displayName + ` ` + ver + `. Installation package is incompatible. Use the correct package."
		echo -1 > /tmp/update_process && exit 1
	fi
}
arch_ok(){
	local cpu_arch=$(/bin/uname -m)
	local reject_platform=$(/sbin/getcfg "" Platform -f /etc/platform.conf)
	if [ $(/usr/bin/expr match "$reject_platform" "` + spec.rejectPlatform + `") == 0 ]; then
		[ $(/usr/bin/expr match "$cpu_arch" "` + spec.cpuArchRegex + `") -ne 0 ] || return 1
	else
		return 1
	fi
}
/bin/echo "Install QNAP package on TS-NAS..."
/bin/grep "/mnt/HDA_ROOT" /proc/mounts >/dev/null 2>&1 || exit 1
arch_ok || wrong_arch
find_base
_EXTRACT_DIR="$QPKG_INSTALL_PATH/.tmp-` + name + `"
/bin/mkdir -p $_EXTRACT_DIR || exit 1
script_len=SCRIPT_LEN
/bin/dd if="${0}" bs=$script_len skip=1 | /bin/tar -xO | /bin/tar -xzv -C $_EXTRACT_DIR || exit 1
offset=$(/usr/bin/expr $script_len + ` + strconv.Itoa(ctrlLen) + `)
/bin/dd if="${0}" bs=$offset skip=1 | /bin/cat | /bin/dd bs=1024 count=` + strconv.Itoa(count1k) + ` of=$_EXTRACT_DIR/data.tar.gz || exit 1
[ -f /usr/local/bin/python ] && /usr/local/bin/python -c "with open('$_EXTRACT_DIR/data.tar.gz', 'rw+') as f: f.seek(` + strconv.Itoa(dataLen) + `); f.truncate()"
offset=$(/usr/bin/expr $offset + ` + strconv.Itoa(dataLen) + `)
( cd $_EXTRACT_DIR && /bin/sh qinstall.sh || echo "Installation Abort." )
/bin/rm -fr $_EXTRACT_DIR && exit 10
exit 1
`
	tpl = strings.ReplaceAll(tpl, "\r\n", "\n")
	baseLen := len(tpl) - len("SCRIPT_LEN")
	digits := len(strconv.Itoa(baseLen))
	scriptLen := baseLen + digits
	if len(strconv.Itoa(scriptLen)) > digits {
		scriptLen++
	}
	final := strings.Replace(tpl, "SCRIPT_LEN", strconv.Itoa(scriptLen), 1)
	if len(final) != scriptLen {
		fatalf("internal error: header len %d != script_len %d", len(final), scriptLen)
	}
	return []byte(final)
}

func buildQPKGTail(name, ver string) []byte {
	var tail [100]byte
	for i := range tail {
		tail[i] = ' '
	}
	if len(name) > 20 {
		name = name[:20]
	}
	if len(ver) > 10 {
		ver = ver[:10]
	}
	copy(tail[60:80], name)
	copy(tail[80:90], ver)
	copy(tail[90:100], "QNAPQPKG  ")
	return tail[:]
}

func applyQPKGEncrypt(raw []byte) {
	// Matches QDK/src/qpkg_encrypt.c: fileSize * 3589 + 1000000000 at offset -60.
	en := uint64(len(raw))*3589 + 1000000000
	s := strconv.FormatUint(en, 10)
	if len(s) > 10 {
		s = s[:10]
	}
	copy(raw[len(raw)-60:len(raw)-50], s)
}

func writeTarFile(tw *tar.Writer, name string, data []byte, mode int64, modTime time.Time) error {
	if err := tw.WriteHeader(&tar.Header{
		Name:     name,
		Mode:     mode,
		Size:     int64(len(data)),
		Typeflag: tar.TypeReg,
		ModTime:  modTime,
	}); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}

func normalizeLF(b []byte) []byte {
	return bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))
}

func isTextAsset(rel string) bool {
	ext := strings.ToLower(filepath.Ext(rel))
	switch ext {
	case ".sh", ".html", ".css", ".js", ".cfg", ".json":
		return true
	}
	return filepath.Base(rel) == "package_routines"
}

func isExecutableAsset(rel string, data []byte) bool {
	ext := strings.ToLower(filepath.Ext(rel))
	if ext == ".sh" || ext == ".cgi" {
		return true
	}
	return len(data) >= 4 && bytes.Equal(data[:4], []byte{0x7f, 'E', 'L', 'F'})
}

func parseCfgValue(cfg, key string) string {
	for _, line := range strings.Split(cfg, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, key+"=") {
			val := strings.TrimPrefix(line, key+"=")
			return strings.Trim(val, `"'`)
		}
	}
	return ""
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

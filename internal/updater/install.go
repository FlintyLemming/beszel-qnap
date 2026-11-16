package updater

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"debug/elf"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"beszel-qnap/internal/agentctl"
)

// ValidateELF verifies that path is a Linux ELF binary matching targetGOARCH.
func ValidateELF(path, targetGOARCH string) error {
	f, err := elf.Open(path)
	if err != nil {
		return fmt.Errorf("file is not a valid Linux ELF executable: %w", err)
	}
	defer f.Close()

	if f.Type != elf.ET_EXEC && f.Type != elf.ET_DYN {
		return fmt.Errorf("unexpected ELF type: %v", f.Type)
	}

	var wantMachine elf.Machine
	switch targetGOARCH {
	case "amd64":
		wantMachine = elf.EM_X86_64
	case "arm64":
		wantMachine = elf.EM_AARCH64
	case "arm":
		wantMachine = elf.EM_ARM
	default:
		return nil
	}
	if f.Machine != wantMachine {
		return fmt.Errorf("binary architecture mismatch: got %v, expected %v (%s)", f.Machine, wantMachine, targetGOARCH)
	}
	return nil
}

// VerifySHA256 checks expectedDigest (either "sha256:<hex>" or "<hex>") against filePath.
func VerifySHA256(filePath, expectedDigest string) error {
	expectedDigest = strings.TrimSpace(expectedDigest)
	if expectedDigest == "" {
		return nil
	}
	expectedDigest = strings.TrimPrefix(expectedDigest, "sha256:")
	actual, err := agentctl.FileSHA256(filePath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(actual, expectedDigest) {
		return fmt.Errorf("SHA-256 mismatch: expected %s, got %s", expectedDigest, actual)
	}
	return nil
}

// ExtractBinary extracts beszel-agent from a .tar.gz, .zip, or raw ELF binary into destPath.
func ExtractBinary(srcPath, origName, destPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	header := make([]byte, 4)
	n, _ := io.ReadFull(f, header)
	_ = f.Close()

	switch {
	case n >= 4 && bytes.Equal(header[:4], []byte{0x7f, 'E', 'L', 'F'}):
		return copyBinaryFile(srcPath, destPath)
	case n >= 2 && header[0] == 0x1f && header[1] == 0x8b:
		return extractFromTarGz(srcPath, destPath)
	case n >= 4 && header[0] == 'P' && header[1] == 'K' && header[2] == 0x03 && header[3] == 0x04:
		return extractFromZip(srcPath, destPath)
	default:
		return fmt.Errorf("unsupported file format %q (expected beszel-agent ELF binary, .tar.gz, or .zip)", origName)
	}
}

func extractFromTarGz(srcPath, destPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("invalid gzip archive: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("failed reading tar archive: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		base := filepath.Base(hdr.Name)
		if base == "beszel-agent" {
			return writeLimitedFile(tr, destPath, 256<<20)
		}
	}
	return errors.New("beszel-agent binary not found inside tar.gz archive")
}

func extractFromZip(srcPath, destPath string) error {
	zr, err := zip.OpenReader(srcPath)
	if err != nil {
		return fmt.Errorf("invalid zip archive: %w", err)
	}
	defer zr.Close()

	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		if filepath.Base(zf.Name) == "beszel-agent" {
			rc, err := zf.Open()
			if err != nil {
				return err
			}
			err = writeLimitedFile(rc, destPath, 256<<20)
			_ = rc.Close()
			return err
		}
	}
	return errors.New("beszel-agent binary not found inside zip archive")
}

func copyBinaryFile(srcPath, destPath string) error {
	in, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer in.Close()
	return writeLimitedFile(in, destPath, 256<<20)
}

func writeLimitedFile(r io.Reader, destPath string, maxBytes int64) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	n, err := io.Copy(out, io.LimitReader(r, maxBytes+1))
	closeErr := out.Close()
	if err != nil {
		_ = os.Remove(destPath)
		return err
	}
	if closeErr != nil {
		_ = os.Remove(destPath)
		return closeErr
	}
	if n > maxBytes {
		_ = os.Remove(destPath)
		return errors.New("extracted binary exceeds maximum allowed size")
	}
	if n == 0 {
		_ = os.Remove(destPath)
		return errors.New("extracted binary is empty")
	}
	return os.Chmod(destPath, 0755)
}

// InstallPreparedBinary validates, swaps in the new binary with .bak backup, and restarts if appropriate.
func InstallPreparedBinary(mgr *agentctl.Manager, candidatePath string, meta *agentctl.BinaryMeta, restartIfRunning bool) error {
	if err := mgr.Paths.EnsureDirs(); err != nil {
		return err
	}
	if err := ValidateELF(candidatePath, runtime.GOARCH); err != nil {
		return err
	}
	if err := os.Chmod(candidatePath, 0755); err != nil {
		return err
	}

	if ver, err := mgr.ProbeBinaryVersion(candidatePath); err == nil && ver != "" {
		meta.Version = ver
	}
	if meta.SHA256 == "" {
		if sum, err := computeSHA256(candidatePath); err == nil {
			meta.SHA256 = sum
		}
	}
	meta.UpdatedAt = time.Now().Format(time.RFC3339)

	wasRunning := mgr.FindRunningPID() > 0
	if wasRunning {
		if err := mgr.Stop(); err != nil {
			return fmt.Errorf("failed stopping running agent before update: %w", err)
		}
	}

	target := mgr.Paths.BinaryPath()
	backup := mgr.Paths.BackupBinaryPath()

	if _, err := os.Stat(target); err == nil {
		_ = os.Remove(backup)
		if err := os.Rename(target, backup); err != nil {
			return fmt.Errorf("failed creating binary backup: %w", err)
		}
	}

	if err := os.Rename(candidatePath, target); err != nil {
		// Revert backup if swap fails.
		if _, berr := os.Stat(backup); berr == nil {
			_ = os.Rename(backup, target)
		}
		return fmt.Errorf("failed installing new binary: %w", err)
	}
	_ = os.Chmod(target, 0755)
	_ = mgr.SaveBinaryMeta(meta)
	mgr.LogManager("Installed beszel-agent %s (source=%s, sha256=%s)", meta.Version, meta.Source, meta.SHA256)

	if wasRunning && restartIfRunning {
		if err := mgr.Start(); err != nil {
			return fmt.Errorf("binary installed (%s), but restart failed: %w", meta.Version, err)
		}
	}
	return nil
}

// RollbackBinary restores beszel-agent.bak to beszel-agent.
func RollbackBinary(mgr *agentctl.Manager) error {
	target := mgr.Paths.BinaryPath()
	backup := mgr.Paths.BackupBinaryPath()

	fi, err := os.Stat(backup)
	if err != nil || fi.IsDir() || fi.Size() == 0 {
		return errors.New("no backup binary available to roll back")
	}

	wasRunning := mgr.FindRunningPID() > 0
	if wasRunning {
		if err := mgr.Stop(); err != nil {
			return fmt.Errorf("failed stopping agent for rollback: %w", err)
		}
	}

	tmp := target + ".swap"
	_ = os.Remove(tmp)
	hasCurrent := false
	if _, err := os.Stat(target); err == nil {
		hasCurrent = true
		if err := os.Rename(target, tmp); err != nil {
			return err
		}
	}
	if err := os.Rename(backup, target); err != nil {
		if hasCurrent {
			_ = os.Rename(tmp, target)
		}
		return err
	}
	if hasCurrent {
		_ = os.Rename(tmp, backup)
	}
	_ = os.Chmod(target, 0755)

	ver, _ := mgr.ProbeBinaryVersion(target)
	sum, _ := computeSHA256(target)
	_ = mgr.SaveBinaryMeta(&agentctl.BinaryMeta{
		Version:   ver,
		Source:    "rollback",
		SHA256:    sum,
		UpdatedAt: time.Now().Format(time.RFC3339),
	})
	mgr.LogManager("Rolled back beszel-agent to %s", ver)

	if wasRunning {
		return mgr.Start()
	}
	return nil
}

func computeSHA256(path string) (string, error) {
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

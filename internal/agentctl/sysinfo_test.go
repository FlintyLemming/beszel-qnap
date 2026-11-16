package agentctl

import (
	"os"
	"testing"
)

func TestCollectHostInfoWithPathsDetectsBundledSmartctl(t *testing.T) {
	root := t.TempDir()
	paths := &Paths{Root: root}
	if err := paths.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}
	if err := os.MkdirAll(paths.BundledBinDir(), 0755); err != nil {
		t.Fatalf("MkdirAll BundledBinDir: %v", err)
	}
	smartctlPath := paths.BundledSmartctlPath()
	if err := os.WriteFile(smartctlPath, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatalf("WriteFile smartctl: %v", err)
	}

	if got := detectSmartctl(paths); got != smartctlPath {
		t.Fatalf("detectSmartctl() = %q, want %q", got, smartctlPath)
	}

	dirs := detectExtraPathDirs(paths)
	foundBundled := false
	for _, d := range dirs {
		if d == paths.BundledBinDir() {
			foundBundled = true
			break
		}
	}
	if !foundBundled {
		t.Fatalf("detectExtraPathDirs() = %v, expected %q to be included", dirs, paths.BundledBinDir())
	}
}

package compress

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompressor_Compress(t *testing.T) {
	srcDir, err := os.MkdirTemp("", "seu-compress-src-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(srcDir)

	if err := os.WriteFile(filepath.Join(srcDir, "install.sh"), []byte("#!/bin/bash\necho test\n"), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(srcDir, "app.txt"), []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	destPath := filepath.Join(t.TempDir(), "package.tar.gz")
	c := NewCompressor(srcDir, destPath)
	if err := c.Compress(); err != nil {
		t.Fatal(err)
	}

	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("expected output file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("expected output file to have content")
	}
}

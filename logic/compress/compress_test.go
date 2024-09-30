package compress

import (
	"os"
	"testing"
)

func TestCompressor_Compress(t *testing.T) {
	tempDir, err := os.MkdirTemp("/tmp", "test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	c := NewCompressor(tempDir, "testdata.tar.gz")
	if err := c.Compress(); err != nil {
		t.Fatal(err)
	}

	f := NewCompressor("test", "testdata.tar.gz")
	if err := f.Decompress(); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("test")
	defer os.RemoveAll("testdata.tar.gz")
}

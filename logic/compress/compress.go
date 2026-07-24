package compress

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Compressor struct {
	TargetPath string
	Path       string
}

func NewCompressor(path, targetpath string) *Compressor {
	return &Compressor{
		Path:       path,
		TargetPath: targetpath,
	}
}

func (c *Compressor) Compress() error {
	if _, err := os.Stat(c.Path); err != nil {
		return fmt.Errorf("source directory not accessible: %w", err)
	}
	installPath := filepath.Join(c.Path, "install.sh")
	if _, err := os.Stat(installPath); err != nil {
		return fmt.Errorf("install.sh not found in source directory %q: %w", c.Path, err)
	}

	os.RemoveAll(c.TargetPath)
	compressedFile, err := os.Create(c.TargetPath)
	if err != nil {
		return err
	}
	defer compressedFile.Close()

	gw := gzip.NewWriter(compressedFile)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	installScript, err := os.Open(installPath)
	if err != nil {
		return fmt.Errorf("open install.sh: %w", err)
	}
	installScriptStat, err := installScript.Stat()
	if err != nil {
		installScript.Close()
		return err
	}
	installScriptInfo, err := tar.FileInfoHeader(installScriptStat, "")
	if err != nil {
		installScript.Close()
		return err
	}
	installScriptInfo.Name = "install.sh"
	if err := tw.WriteHeader(installScriptInfo); err != nil {
		installScript.Close()
		return err
	}
	if _, err := io.Copy(tw, installScript); err != nil {
		installScript.Close()
		return err
	}
	installScript.Close()

	return filepath.Walk(c.Path, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if file == c.Path {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		if fi.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(file)
			if err != nil {
				return err
			}
			header.Linkname = linkTarget
			header.Typeflag = tar.TypeSymlink
			header.Size = 0
		}
		header.Name = strings.TrimPrefix(file, c.Path+"/")
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if fi.IsDir() {
			return nil
		}
		if !fi.Mode().IsRegular() && fi.Mode().Type() != os.ModeSymlink {
			return nil
		}
		if fi.Mode().Type() == os.ModeSymlink {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(tw, f)
		f.Close()
		return copyErr
	})
}

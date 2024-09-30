package compress

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Compressor struct {
	TargetPath string // 压缩时输出的文件，解压时输入的文件
	Path       string // 压缩时输入的文件目录，解压时输出的文件目录
}

func NewCompressor(path, targetpath string) *Compressor {
	return &Compressor{
		Path:       path,
		TargetPath: targetpath,
	}
}

func (c *Compressor) Compress() error {
	os.RemoveAll(c.TargetPath)
	compressedFile, err := os.Create(c.TargetPath)
	if err != nil {
		return err
	}
	gw := gzip.NewWriter(compressedFile)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	installScript, err := os.Open("./install.sh")
	if err != nil {
		return fmt.Errorf("open install.sh failed, install.sh is required, you can leave it blank if you don't want to use it.: %w", err)
	}
	installScriptStat, err := installScript.Stat()
	if err != nil {
		return err
	}
	installScriptInfo, err := tar.FileInfoHeader(installScriptStat, "")
	if err != nil {
		return err
	}
	tw.WriteHeader(installScriptInfo)
	_, err = io.Copy(tw, installScript)
	if err != nil {
		return err
	}
	defer installScript.Close()
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

		// 处理符号链接
		if fi.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(file)
			if err != nil {
				return err
			}
			header.Linkname = linkTarget
			header.Typeflag = tar.TypeSymlink
			header.Size = 0 // 对于符号链接，Size 应该是 0
		}
		header.Name = strings.TrimPrefix(file, "/")
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() && fi.Mode() != os.ModeSymlink {
			return nil
		}
		if fi.IsDir() {
			return nil
		}
		f, err := os.Open(file)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
}

func (c *Compressor) Decompress() error {
	f, err := os.Open(c.TargetPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(c.Path, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeSymlink:
			linkTarget := header.Linkname
			if err := os.Symlink(linkTarget, target); err != nil && !errors.Is(err, os.ErrExist) {
				fmt.Println()
				return err
			}
			continue
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			defer f.Close()
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}
		}
	}
	return nil
}

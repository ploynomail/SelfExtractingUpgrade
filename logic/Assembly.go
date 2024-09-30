package logic

import (
	"encoding/hex"
	"os"
	"text/template"

	"github.com/ploynomail/SelfExtractingUpgrade/logic/compress"
	"github.com/ploynomail/SelfExtractingUpgrade/logic/keys"
	signatureverify "github.com/ploynomail/SelfExtractingUpgrade/logic/signatureVerify"
)

type AutoDeCompressAssembly struct {
	TargetPath string // 压缩时输出的文件，解压时输入的文件
	Path       string // 压缩时输入的文件目录，解压时输出的文件目录
	IsSign     bool   // 是否需要签名
	PrivateKey string // 私钥
	Isencrypt  bool   // 是否需要加密
	Password   string // 加密密码
}

func NewAutoDeCompressAssembly(path, targetpath string) *AutoDeCompressAssembly {
	return &AutoDeCompressAssembly{
		Path:       path,
		TargetPath: targetpath,
		IsSign:     false,
		Isencrypt:  false,
	}
}

func (c *AutoDeCompressAssembly) WithSign(privateKey string) *AutoDeCompressAssembly {
	c.IsSign = true
	c.PrivateKey = privateKey
	return c
}

func (c *AutoDeCompressAssembly) WithEncrypt(password string) *AutoDeCompressAssembly {
	c.Isencrypt = true
	c.Password = password
	return c
}

func (c *AutoDeCompressAssembly) Assembly() error {
	sctm := template.New("AutoDeCompressAssembly")
	sctm, err := sctm.Parse(ScriptTemplate)
	if err != nil {
		return err
	}
	// 生成压缩包
	comp := compress.NewCompressor(c.Path, c.TargetPath)
	if err := comp.Compress(); err != nil {
		return err
	}
	// 生成签名
	var signature []byte
	if c.IsSign {
		keys := keys.NewGenerateEcdsaKeys()
		privateKey, err := keys.LoadPrivateKey(c.PrivateKey)
		if err != nil {
			return err
		}
		f, err := os.Open(c.TargetPath)
		if err != nil {
			return err
		}
		signature, err = signatureverify.SignFile(privateKey, f)
		if err != nil {
			return err
		}
	}
	// 生成加密
	// 生成最终文件
	f, err := os.ReadFile(c.TargetPath)
	if err != nil {
		return err
	}
	var data = struct {
		Isencrypt bool
		Signature string
		IsSign    bool
	}{
		Isencrypt: c.Isencrypt,
		Signature: hex.EncodeToString(signature),
		IsSign:    c.IsSign,
	}
	// fmt.Println(sctm.ExecuteTemplate(os.Stdout, "AutoDeCompressAssembly", data))
	// 生成模板写入文件
	selfRunfile, err := os.Create(c.TargetPath + ".run")
	selfRunfile.Chmod(0755)
	if err != nil {
		return err
	}
	defer selfRunfile.Close()
	if err := sctm.ExecuteTemplate(selfRunfile, "AutoDeCompressAssembly", data); err != nil {
		return err
	}
	// 写入压缩包
	if _, err := selfRunfile.Write(f); err != nil {
		return err
	}
	return nil
}
# SelfExtractingUpgrade 项目整改报告

**项目版本：** V1.0.0（commit 80978d9）  
**审查日期：** 2026-07-23  
**审查范围：** 全部源代码（cmd/、logic/、logic/compress/、logic/keys/、logic/signatureVerify/）

---

## 一、整改总览

| 优先级 | 数量 | 说明 |
|--------|------|------|
| P0 | 5 | 导致程序崩溃或安全机制完全失效 |
| P1 | 7 | 安全漏洞或核心功能缺陷 |
| P2 | 8 | 可靠性与可维护性问题 |
| P3 | 10 | CLI 设计与代码质量 |
| 架构 | 5 | 整体设计改进 |

---

## 二、P0 — 严重缺陷

### 2.1 硬编码 IV 破坏加密安全性

**位置：** `logic/encryption.go:10`

**现状：**
```go
var IV = []byte("N0P2N1K6A5X8P8N3")
```

**问题：** 所有加密操作使用同一固定 IV。在 CBC 模式下，相同密钥 + 相同 IV 使加密变为确定性操作，攻击者可通过对比密文发现重复明文块，进而实施模式分析和选择明文攻击。

**解决方案：**

1. 每次加密时使用 `crypto/rand` 生成 16 字节随机 IV。
2. 将 IV 拼接在密文前面（IV 无需保密，只需唯一）。
3. 解密时从密文头部读取前 16 字节作为 IV。
4. 生成的 bash 脚本中对应修改：先用 `dd` 提取前 16 字节作为 IV，剩余部分为实际密文。

```go
// 加密
iv := make([]byte, aes.BlockSize)
if _, err := rand.Read(iv); err != nil {
    return nil, err
}
ciphertext := append(iv, encrypted...)

// 解密（bash 脚本中）
// iv=$(dd if=payload bs=1 count=16 | xxd -p)
// dd if=payload bs=1 skip=16 | openssl aes-256-cbc -d -K "$key" -iv "$iv"
```

---

### 2.2 参数校验失败后未终止执行

**位置：** `cmd/make.go:27-51`

**现状：**
```go
if sourcePath == "" {
    FmtError("source path is empty")
    // 没有 return，继续执行
}
```

**问题：** 校验失败后程序继续以空值执行后续逻辑，导致不可预期的错误或静默产出损坏文件。

**解决方案：**

将 `make` 命令的 `Run` 改为 `RunE`，校验失败时直接返回 error：

```go
var makeCmd = &cobra.Command{
    Use:   "make",
    Short: "Make self extracting upgrade package",
    RunE: func(cmd *cobra.Command, args []string) error {
        if sourcePath == "" {
            return fmt.Errorf("source path (-s) is required")
        }
        if destPath == "" {
            return fmt.Errorf("destination path (-d) is required")
        }
        if (isSign || isOverallSign) && privateKey == "" {
            return fmt.Errorf("private key (-k) is required when signing is enabled")
        }
        if isEncrypt && password == "" {
            return fmt.Errorf("password (-p) is required when encryption is enabled")
        }
        // ... 正常逻辑
        return nil
    },
}
```

同时删除 `FmtError` 辅助函数，统一使用 error 返回值。Cobra 会自动打印错误并设置退出码为 1。

---

### 2.3 `os.Create` 失败时空指针 panic

**位置：** `logic/Assembly.go:111-113`

**现状：**
```go
selfRunfile, err := os.Create(c.TargetPath + ".run")
selfRunfile.Chmod(0755)  // err 未检查，selfRunfile 可能为 nil
if err != nil {
```

**问题：** 当文件创建失败（权限不足、磁盘满等），`selfRunfile` 为 nil，调用 `Chmod` 触发 nil pointer dereference panic。

**解决方案：**

```go
selfRunfile, err := os.Create(c.TargetPath + ".run")
if err != nil {
    return fmt.Errorf("create .run file: %w", err)
}
if err := selfRunfile.Chmod(0755); err != nil {
    selfRunfile.Close()
    return fmt.Errorf("chmod .run file: %w", err)
}
```

---

### 2.4 模板前导换行破坏 Shebang

**位置：** `logic/scriptTemplate.go:3-4`

**现状：** 模板字符串以 `\n` 开头，导致生成的 `.run` 文件第一行为空行，`#!/bin/bash` 位于第二行。

**问题：** Linux 内核要求 `#!` 必须是文件的前两个字节。前导换行使内核无法识别脚本类型，执行时可能报错或回退到当前 shell 解释。

**解决方案：**

删除模板开头的换行符，确保 `#!/bin/bash` 是模板的第一个字符：

```go
var ScriptTemplate = `#!/bin/bash
# Self Extracting Upgrade Package
...
`
```

---

### 2.5 `set -e` 与 `$?` 检查逻辑矛盾

**位置：** `logic/scriptTemplate.go:8` 及 `77/89/103` 行

**现状：**
```bash
set -e
# ...
openssl dgst -verify $pubkey -signature ${tmp_dir}/payload.sig ${tmp_dir}/payload
if [ $? -ne 0 ]; then   # 永远不会执行到这里
    echo "Signature verification failed"
    exit 1
fi
```

**问题：** `set -e` 下命令失败时脚本立即退出，`if [ $? ...]` 分支永远不会被执行。所有自定义错误提示和清理逻辑都是死代码。

**解决方案：**

移除 `set -e`，改用 `if ! command; then` 模式进行显式错误处理：

```bash
# 不使用 set -e

if ! openssl dgst -sha256 -verify "$pubkey" -signature "${tmp_dir}/payload.sig" "${tmp_dir}/payload"; then
    echo "ERROR: Signature verification failed" >&2
    exit 1
fi

if ! openssl aes-256-cbc -d -K "$key" -iv "$iv" -in "${tmp_dir}/payload" -out "${tmp_dir}/payload.tar.gz"; then
    echo "ERROR: Decryption failed" >&2
    exit 1
fi

if ! tar -xzf "${tmp_dir}/payload.tar.gz" -C "${tmp_dir}"; then
    echo "ERROR: Extraction failed" >&2
    exit 1
fi
```

---

## 三、P1 — 高风险问题

### 3.1 固定临时目录导致符号链接攻击

**位置：** `logic/scriptTemplate.go:10`

**现状：**
```bash
tmp_dir=/tmp/extract_dir
```

**问题：**
- 攻击者可预先创建 `/tmp/extract_dir` 为指向敏感目录的符号链接，脚本解压时会覆盖目标。
- 多个实例并发执行时互相覆盖。

**解决方案：**

```bash
tmp_dir=$(mktemp -d /tmp/seu_XXXXXXXX)
if [ $? -ne 0 ]; then
    echo "ERROR: Failed to create temp directory" >&2
    exit 1
fi
trap 'rm -rf "$tmp_dir"' EXIT INT TERM
```

同时删除脚本末尾的手动 `rm -rf` 语句，由 trap 统一清理。

---

### 3.2 私钥文件权限过于宽松

**位置：** `logic/keys/ecdsaKeys.go:59`

**现状：**
```go
os.WriteFile(filename+".key", pemPriv, 0644)
```

**问题：** 0644 意味着系统上所有用户都可读取签名私钥。

**解决方案：**

```go
if err := os.WriteFile(filename+".key", pemPriv, 0600); err != nil {
    return fmt.Errorf("save private key: %w", err)
}
if err := os.WriteFile(filename+".pub", pemPub, 0644); err != nil {
    return fmt.Errorf("save public key: %w", err)
}
```

公钥保持 0644（需要分发），私钥改为 0600（仅所有者可读写）。

---

### 3.3 密码和私钥通过 CLI 参数传递

**位置：** `cmd/make.go:65-66`

**问题：** CLI 参数在 `ps aux`、`/proc/*/cmdline`、shell history 中均可见。

**解决方案：**

支持多种输入方式，优先级从高到低：

```go
// 1. 环境变量
//    SEU_PASSWORD=xxx ./SelfExtractingUpgrade make ...
// 2. 从文件读取
//    --password-file /path/to/passfile
// 3. 从 stdin 读取（交互式）
//    终端提示输入，不回显

func resolvePassword(cmd *cobra.Command) (string, error) {
    if pw, _ := cmd.Flags().GetString("password"); pw != "" {
        fmt.Fprintln(os.Stderr, "WARNING: --password on command line is insecure, use SEU_PASSWORD env or --password-file")
        return pw, nil
    }
    if file, _ := cmd.Flags().GetString("password-file"); file != "" {
        data, err := os.ReadFile(file)
        return strings.TrimSpace(string(data)), err
    }
    if pw := os.Getenv("SEU_PASSWORD"); pw != "" {
        return pw, nil
    }
    // 交互式读取
    fmt.Fprint(os.Stderr, "Enter encryption password: ")
    pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
    fmt.Fprintln(os.Stderr)
    return string(pwBytes), err
}
```

私钥同理，改为 `--private-key-file` 接受文件路径（当前 `-k` 参数语义改为文件路径而非密钥内容）。

---

### 3.4 `LoadPrivateKey` 文件不存在时静默生成新密钥

**位置：** `logic/keys/ecdsaKeys.go:64-72`

**问题：** 用户路径拼写错误时不报错，静默生成全新密钥对。后续签名无法通过已有公钥验证，且用户毫不知情。

**解决方案：**

```go
func (g *GenerateEcdsaKeys) LoadPrivateKey(filename string) (*ecdsa.PrivateKey, error) {
    keyBytes, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("read private key file %q: %w", filename, err)
    }
    p, _ := pem.Decode(keyBytes)
    if p == nil {
        return nil, fmt.Errorf("invalid PEM data in %q", filename)
    }
    key, err := x509.ParseECPrivateKey(p.Bytes)
    if err != nil {
        return nil, fmt.Errorf("parse EC private key: %w", err)
    }
    return key, nil
}
```

密钥生成仅通过 `generateKeys` 子命令显式执行。

---

### 3.5 Tar 解压路径穿越（Zip Slip）

**位置：** `logic/compress/compress.go:118`

**现状：**
```go
target := filepath.Join(c.Path, header.Name)
```

**问题：** 恶意 tar 可包含 `../../etc/cron.d/evil` 等路径，解压时写入目标目录之外的位置。

**解决方案：**

```go
target := filepath.Join(c.Path, header.Name)
// 确保解压路径不逃逸目标目录
if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(c.Path)+string(os.PathSeparator)) {
    return fmt.Errorf("illegal path in archive (path traversal): %s", header.Name)
}
```

对符号链接同样校验：

```go
case tar.TypeSymlink:
    linkTarget := filepath.Join(filepath.Dir(target), header.Linkname)
    if !strings.HasPrefix(filepath.Clean(linkTarget), filepath.Clean(c.Path)+string(os.PathSeparator)) {
        return fmt.Errorf("symlink escapes target directory: %s -> %s", header.Name, header.Linkname)
    }
    os.Symlink(header.Linkname, target)
```

---

### 3.6 整体签名（Overall Sign）功能不完整

**位置：** `logic/Assembly.go:127-143`

**问题：** 整体签名仅打印到 stdout，未写入任何文件。验证方无法获取该签名，功能无实际意义。

**解决方案（二选一）：**

**方案 A：生成 `.sig` 伴随文件（推荐）**

```go
if c.OverallSign {
    sig, err := signatureVerify.SignFile(c.TargetPath+".run", privKey)
    if err != nil {
        return err
    }
    sigPath := c.TargetPath + ".run.sig"
    if err := os.WriteFile(sigPath, sig, 0644); err != nil {
        return err
    }
    fmt.Printf("Overall signature saved to: %s\n", sigPath)
    fmt.Printf("Verify with: openssl dgst -sha256 -verify pubkey.pem -signature %s %s.run\n", sigPath, c.TargetPath)
}
```

分发时同时提供 `.run` + `.run.sig` + `pubkey.pem` 三个文件。

**方案 B：将签名追加到 `.run` 文件尾部**

在 `__ARCHIVE_BELOW__` 标记之前插入签名段，脚本启动时先自校验。此方案实现复杂度较高，推荐方案 A。

---

### 3.7 密码直接作为 AES 密钥，无密钥派生

**位置：** `logic/Assembly.go:63`

**现状：**
```go
key := []byte(c.Password)
```

**问题：**
- 密码长度不为 16/24/32 字节时 `aes.NewCipher` 报错。
- 无 KDF 保护，暴力破解成本极低。
- 与脚本端 `openssl aes-256-cbc -K` 的 hex key 语义不一致。

**解决方案：**

使用 PBKDF2 从密码派生固定 32 字节密钥，并生成随机 salt：

```go
import "golang.org/x/crypto/pbkdf2"

salt := make([]byte, 16)
rand.Read(salt)
key := pbkdf2.Key([]byte(c.Password), salt, 100000, 32, sha256.New)
```

密文格式：`[16B salt][16B IV][ciphertext...]`

bash 脚本端对应修改：

```bash
salt=$(dd if="$payload_file" bs=1 count=16 | xxd -p)
iv=$(dd if="$payload_file" bs=1 skip=16 count=16 | xxd -p)
key=$(echo -n "$password" | openssl kdf -keylen 32 -kdfopt digest:SHA256 \
      -kdfopt pass:stdin -kdfopt salt:$salt -kdfopt iter:100000 PBKDF2)
dd if="$payload_file" bs=1 skip=32 | openssl aes-256-cbc -d -K "$key" -iv "$iv"
```

若目标机 openssl 版本不支持 `kdf` 子命令，可退化为在脚本中用 `openssl pbkdf2` 或预计算 key 后以参数传入。

---

## 四、P2 — 可靠性与可维护性

### 4.1 文件句柄泄漏

**位置：** `logic/Assembly.go:87,134`；`logic/compress/compress.go:92`

**解决方案：**

Assembly.go 中签名打开的文件：
```go
f, err := os.Open(c.TargetPath)
if err != nil {
    return err
}
defer f.Close()
```

compress.go Walk 回调中，将 `defer f.Close()` 改为即时关闭：
```go
f, err := os.Open(path)
if err != nil {
    return err
}
_, err = io.Copy(tw, f)
f.Close()  // 立即关闭，不用 defer
return err
```

---

### 4.2 `openssl dgst` 未显式指定哈希算法

**位置：** `logic/scriptTemplate.go:76`

**解决方案：**

```bash
openssl dgst -sha256 -verify "$pubkey" -signature "${tmp_dir}/payload.sig" "${tmp_dir}/payload"
```

与 Go 端 `sha256.New()` 保持一致。

---

### 4.3 AES 密钥长度 Go 端与脚本端不一致

**位置：** `logic/scriptTemplate.go:88`

**问题：** Go 的 `aes.NewCipher` 根据 key 长度自动选择 AES-128/192/256，脚本写死 `aes-256-cbc`。

**解决方案：**

在采用 3.7 的 PBKDF2 方案后，密钥固定为 32 字节，脚本端固定使用 `aes-256-cbc`，两端一致。

若暂不引入 KDF，则在 Go 端强制校验密码长度为 32 字节（64 个 hex 字符），不满足时报错退出：

```go
if len(c.Password) != 32 {
    return fmt.Errorf("password must be exactly 32 bytes for AES-256, got %d", len(c.Password))
}
```

---

### 4.4 `install.sh` 使用 source 执行

**位置：** `logic/scriptTemplate.go:109`

**现状：**
```bash
. ./install.sh
```

**问题：** `install.sh` 中的 `exit` 会终止父脚本，跳过清理；变量和函数泄漏到父 shell。

**解决方案：**

```bash
bash ./install.sh
install_rc=$?
if [ $install_rc -ne 0 ]; then
    echo "ERROR: install.sh exited with code $install_rc" >&2
    exit $install_rc
fi
```

---

### 4.5 `install.sh` 从 CWD 读取而非源目录

**位置：** `logic/compress/compress.go:36`

**现状：**
```go
installFile, err := os.Open("./install.sh")
```

**解决方案：**

从用户指定的源目录读取：

```go
installFile, err := os.Open(filepath.Join(c.Path, "install.sh"))
if err != nil {
    return fmt.Errorf("install.sh not found in source directory %q: %w", c.Path, err)
}
```

---

### 4.6 Symlink 模式比较逻辑错误

**位置：** `logic/compress/compress.go:82`

**现状：**
```go
if !fi.Mode().IsRegular() && fi.Mode() != os.ModeSymlink {
```

**问题：** `fi.Mode()` 包含权限位，与 `os.ModeSymlink`（仅类型位）直接 `!=` 比较永远为 true。

**解决方案：**

```go
if !fi.Mode().IsRegular() && fi.Mode().Type() != os.ModeSymlink {
    return nil  // 跳过非常规文件且非符号链接
}
```

---

### 4.7 `Compress()` 先删除目标再校验输入

**位置：** `logic/compress/compress.go:27-38`

**解决方案：**

将校验前置：

```go
func (c *Compressor) Compress() error {
    // 先校验
    if _, err := os.Stat(c.Path); err != nil {
        return fmt.Errorf("source directory not accessible: %w", err)
    }
    installPath := filepath.Join(c.Path, "install.sh")
    if _, err := os.Stat(installPath); err != nil {
        return fmt.Errorf("install.sh not found in source: %w", err)
    }
    // 校验通过后再删除旧产物
    os.RemoveAll(c.TargetPath)
    // ... 继续压缩
}
```

---

### 4.8 `pem.Decode` 返回值未检查

**位置：** `logic/keys/ecdsaKeys.go:73`

**解决方案：**

```go
p, _ := pem.Decode(keyBytes)
if p == nil {
    return nil, fmt.Errorf("no valid PEM block found in key file")
}
```

---

## 五、P3 — CLI 设计与代码质量

### 5.1 短标志 `-p` 含义冲突

**问题：** `make -p` 是密码，`generateKeys -p` 是路径。

**解决方案：**

| 子命令 | 标志 | 含义 |
|--------|------|------|
| `make` | `-s` source, `-d` dest, `-S` sign, `-E` encrypt, `-k` key-file, `-P` password-file, `-O` overall-sign | |
| `generateKeys` | `-o` output-path | |

---

### 5.2 `rootCmd.Use` 包含空格

**位置：** `cmd/root.go:14`

**解决方案：**

```go
var rootCmd = &cobra.Command{
    Use:   "seu",
    Short: "Self Extracting Upgrade - create self-extracting upgrade packages",
}
```

---

### 5.3 `--sign` 与 `--overall-sign` 关系不清

**解决方案：**

统一为 `--sign` 标志，接受枚举值：

```go
makeCmd.Flags().String("sign", "none", "Signing mode: none, payload, overall, both")
```

- `none`：不签名（默认）
- `payload`：对压缩/加密后的 payload 签名（嵌入脚本内验证）
- `overall`：对完整 `.run` 文件签名（生成 `.sig` 伴随文件）
- `both`：两者都签

---

### 5.4 错误处理无退出码

**解决方案：** 全部改用 `RunE` 返回 error（见 2.2）。删除 `FmtError` 函数。

---

### 5.5 死代码清理

删除以下未使用的代码：

| 文件 | 内容 |
|------|------|
| `logic/encryption.go` | `StringToHex()` 函数 |
| `logic/encryption.go` | `decrypt()` 函数（解密在 bash 中完成） |
| `logic/encryption.go` | `pkcs7Unpadding()` 函数 |
| `logic/signatureAndVerify.go` | `SignatureAndVerify` 接口 |
| `logic/compress/compress.go` | `Decompress()` 方法（仅测试使用，且与 bash 端重复） |
| `logic/compress/compress.go:127` | 空 `fmt.Println()` |

若 `Decompress()` 需保留用于测试，则移至 `_test.go` 文件中。

---

### 5.6 命名规范化

| 现有 | 修改为 |
|------|--------|
| `Isencrypt` | `IsEncrypt` |
| `IsSign` | `IsSign`（保持） |
| `OverallSign` | `IsOverallSign` |
| `AutoDeCompressAssembly` | `Assembly` 或 `Packager` |
| `ScriptTemplate` | `scriptTemplate`（非导出） |

---

### 5.7 生成脚本中变量未加引号

**解决方案：** 所有变量展开加双引号：

```bash
rm -rf "$tmp_dir"
tar -xzf "${tmp_dir}/payload.tar.gz" -C "$tmp_dir"
cd "$tmp_dir" || exit 1
```

---

### 5.8 `which` 替换为 `command -v`

```bash
if ! command -v openssl >/dev/null 2>&1; then
    echo "ERROR: openssl is required but not found" >&2
    exit 1
fi
```

---

### 5.9 添加 cleanup trap

见 3.1 解决方案，使用 `trap 'rm -rf "$tmp_dir"' EXIT INT TERM`。

---

### 5.10 `cd` 失败未处理

```bash
cd "$tmp_dir" || { echo "ERROR: cannot cd to $tmp_dir" >&2; exit 1; }
```

---

## 六、架构改进建议

### 6.1 配置与执行分离

**现状：** `AutoDeCompressAssembly` 既是配置结构体又是执行器。

**目标结构：**

```go
// 配置（不可变，创建时校验）
type Config struct {
    SourceDir    string
    DestPath     string
    SignMode     SignMode  // none/payload/overall/both
    Encrypt      bool
    KeyFile      string
    Password     string
}

func NewConfig(opts ...Option) (*Config, error) {
    // 构造时完成所有校验
}

// 执行器
type Packager struct {
    cfg *Config
}

func (p *Packager) Build() error {
    // 按步骤执行
}
```

---

### 6.2 流水线化构建步骤

```go
func (p *Packager) Build() error {
    var payload []byte
    var err error

    if payload, err = p.compress(); err != nil {
        return fmt.Errorf("compress: %w", err)
    }
    if p.cfg.Encrypt {
        if payload, err = p.encrypt(payload); err != nil {
            return fmt.Errorf("encrypt: %w", err)
        }
    }
    if p.cfg.SignMode == SignPayload || p.cfg.SignMode == SignBoth {
        if err = p.signPayload(payload); err != nil {
            return fmt.Errorf("sign payload: %w", err)
        }
    }
    if err = p.renderScript(payload); err != nil {
        return fmt.Errorf("render script: %w", err)
    }
    if p.cfg.SignMode == SignOverall || p.cfg.SignMode == SignBoth {
        if err = p.signOverall(); err != nil {
            return fmt.Errorf("sign overall: %w", err)
        }
    }
    return nil
}
```

每个步骤独立可测试。

---

### 6.3 生成脚本的测试策略

1. **静态检查：** CI 中对生成的脚本运行 `shellcheck`。
2. **集成测试：** 在 Docker 容器中执行完整流程：
   - 构建包（签名 + 加密）
   - 执行 `.run` 脚本
   - 验证 `install.sh` 被正确调用
   - 验证篡改检测（修改 payload 后验签失败）
3. **单元测试：** 对模板渲染逻辑进行表驱动测试，覆盖所有 sign/encrypt 组合。

---

### 6.4 包格式版本标识

在生成的脚本中添加版本注释：

```bash
#!/bin/bash
# SEU-FORMAT-VERSION: 2
# Generated by: SelfExtractingUpgrade v1.1.0
```

未来格式变更时可通过版本号做兼容处理。

---

### 6.5 加密方案整体重设计

**目标密文格式：**

```
[Magic 4B: "SEU\x01"][Version 1B][Salt 16B][IV 16B][Ciphertext ...][HMAC 32B]
```

- 使用 PBKDF2/Argon2 从密码派生 key
- 随机 salt + 随机 IV
- 添加 HMAC-SHA256 做完整性校验（Encrypt-then-MAC）
- 脚本端使用 `openssl` 对应操作解密

---

## 七、整改实施计划

| 阶段 | 内容 | 预估工作量 |
|------|------|-----------|
| 第一阶段 | 修复 P0（5 项）：崩溃和安全失效问题 | 1 天 |
| 第二阶段 | 修复 P1（7 项）：安全漏洞和功能缺陷 | 2-3 天 |
| 第三阶段 | 修复 P2（8 项）：可靠性问题 | 1-2 天 |
| 第四阶段 | P3 + 架构改进：CLI 重设计、代码清理、测试补充 | 3-4 天 |
| 第五阶段 | 集成测试 + 文档更新 + 发布 V2.0.0 | 1-2 天 |

**总计预估：8-12 个工作日**

---

## 八、验收标准

1. 所有 P0/P1 问题修复并通过代码审查。
2. `go vet ./...` 和 `staticcheck ./...` 零告警。
3. 生成的 `.run` 脚本通过 `shellcheck` 零告警。
4. 完整集成测试覆盖：无签名/仅签名/仅加密/签名+加密/整体签名 五种模式。
5. 篡改测试：修改 payload 后验签必须失败。
6. 并发测试：同时执行两个 `.run` 脚本互不干扰。
7. 错误路径测试：缺少 openssl、密码错误、磁盘满等场景有明确错误提示和非零退出码。

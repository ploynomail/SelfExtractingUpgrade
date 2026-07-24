# Self Extracting Upgrade Tool

A self-extracting upgrade tool for Linux that packages a software directory into a single executable shell script (`.run`). The generated script automatically verifies its signature, decrypts its payload, extracts the archive, and executes `install.sh`.

**Security model:**
- **Signing**: ECDSA P-256 with SHA-256 (sender's private key signs, sender's public key is embedded for verification)
- **Encryption**: ECIES hybrid encryption using ECDH P-256 to protect a random AES-256 session key (recipient's public key encrypts, recipient's private key decrypts)

Signing and encryption are **mandatory** for every package.

---

## Table of Contents

* [Features](#features)
* [Installation](#installation)
* [Commands](#commands)
  * [generateKeys](#generatekeys)
  * [make](#make)
  * [version](#version)
* [Quick Start](#quick-start)
* [Runtime Requirements](#runtime-requirements)
* [Security Notes](#security-notes)
* [License](#license)

---

## Features

- Single-file `.run` upgrade package
- Mandatory ECDSA signature verification before extraction
- Mandatory ECIES encryption of payload using recipient's public key
- Random AES-256 session key per package
- Automatic cleanup of temporary files via `trap`
- Optional overall signature of the entire `.run` file (outputs `.run.sig`)

---

## Installation

Build from source with Go 1.22 or later:

```bash
go build -o seu .
```

The binary is named `seu` by default. You can rename it as needed.

---

## Commands

### `generateKeys`

Generate a new ECDSA P-256 public/private key pair.

```bash
seu generateKeys -o <prefix>
```

Flags:

* `-o, --output string` : output path prefix. Generates `<prefix>.key` and `<prefix>.pub`. If omitted, keys are printed to stdout.
* `-h, --help` : help for generateKeys.

Example:

```bash
seu generateKeys -o sender
seu generateKeys -o recipient
ls sender.* recipient.*
# sender.key  sender.pub  recipient.key  recipient.pub
```

The `.key` file is created with permissions `0600`.

---

### `make`

Create a new self-extracting upgrade package. Encryption and signing are mandatory.

```bash
seu make \
  -s <source-dir> \
  -d <dest-prefix> \
  -k <sender-private-key> \
  --sender-pub <sender-public-key> \
  -r <recipient-public-key> \
  [-O]
```

Flags:

* `-s, --source string` : source directory path. Must contain `install.sh`.
* `-d, --dest string` : destination file prefix. Produces `<prefix>` and `<prefix>.run`.
* `-k, --sender-key string` : sender's private key file (for ECDSA signing).
* `--sender-pub string` : sender's public key file (embedded in `.run` for verification).
* `-r, --recipient-pub string` : recipient's public key file (for ECIES session-key encryption).
* `-O, --overall-sign` : additionally sign the entire `.run` file and write `<prefix>.run.sig`.
* `-h, --help` : help for make.

Example:

```bash
seu make -s ./myapp -d ./myapp-upgrade \
  -k sender.key --sender-pub sender.pub \
  -r recipient.pub -O
```

Output:

```text
myapp-upgrade       # encrypted payload (intermediate file)
myapp-upgrade.run   # self-extracting script
myapp-upgrade.run.sig # optional overall signature
```

---

### `version`

Display the version of the tool.

```bash
seu version
```

---

## Quick Start

### 1. Prepare keys

```bash
./seu generateKeys -o sender
./seu generateKeys -o recipient
```

Distribute `sender.pub` with the package. Keep `recipient.key` secret on the target machine.

### 2. Prepare the source directory

```text
myapp/
├── install.sh          # required, executed after extraction
└── ...                 # other files and directories
```

`install.sh` must exist in the source directory. It will be placed at the root of the extracted archive and executed by `bash`.

### 3. Build the package

```bash
./seu make -s myapp -d myapp-upgrade \
  -k sender.key --sender-pub sender.pub \
  -r recipient.pub
```

### 4. Run the package on the target machine

```bash
./myapp-upgrade.run recipient.key
```

The script will:

1. Verify the ECDSA signature using the embedded `sender.pub`.
2. Use `recipient.key` to derive the ECIES shared secret and decrypt the session key.
3. Decrypt the payload with the session key.
4. Extract the archive to a temporary directory.
5. Execute `install.sh`.
6. Clean up the temporary directory.

### 5. Optional overall signature verification

If you built with `-O`, distribute the `.run.sig` file alongside the `.run` file. Verify on the target machine before execution:

```bash
openssl dgst -sha256 -verify sender.pub \
  -signature myapp-upgrade.run.sig myapp-upgrade.run
```

---

## Runtime Requirements

The target machine must have:

* `bash`
* `openssl` (1.1.1 or later, with `pkeyutl -derive` support)
* `xxd`
* `tar`
* `mktemp`
* `awk`

---

## Security Notes

* **Sender key pair**: signs the package. Protect the private key; distribute only the public key.
* **Recipient key pair**: decrypts the package. Protect the private key on the target machine; the public key is used during packaging.
* The AES session key is randomly generated per package and never stored in plain text.
* The `.run` script writes PEM keys and temporary files to a directory created by `mktemp` and removes it on exit via `trap`.
* Do not pass private keys as command-line arguments to the `make` command; always use file paths (`-k`).

---

## License

[MIT License](LICENSE)

---

Feel free to contribute to the project or report issues!

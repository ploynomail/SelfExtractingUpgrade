# Self Extracting Upgrade Tool

A self extracting upgrade tool that simplifies the process of upgrading software packages. It supports encryption and signing for security, automatic decompression in the Linux shell, and execution of an install script to achieve the desired upgrade effect.

## Table of Contents

* [Usage]()
* [Commands]()
  * [Completion]()
  * [Generate Keys]()
  * [Make]()
  * [Version]()
* [Flags]()

## Usage


## Commands

### Completion

Generate the autocompletion script for the specified shell.


#### Available Commands

* **bash** : Generate the autocompletion script for bash.
* **fish** : Generate the autocompletion script for fish.
* **powershell** : Generate the autocompletion script for powershell.
* **zsh** : Generate the autocompletion script for zsh.

#### Flags

* **-h, --help** : help for completion.

### GenerateKeys

Generate a new public/private key pair used to verify the signature of the package.

#### Flags

* **-h, --help** : help for generateKeys.
* **-p, --privateKeyPath string** : The path to the private key, if not provided a new key will be generated and output to the console.

### Make

Create a new self extracting upgrade package.

#### Flags

* **-d, --dest string** : destination file name.
* **-c, --encrypt** : encrypt the package.
* **-h, --help** : help for make.
* **-p, --password string** : password.
* **-k, --private-key string** : private key.
* **-i, --sign** : sign the package.
* **-s, --source string** : source file path.

### Version

Display the version of the self extracting upgrade tool.

## Flags

* **-h, --help** : help for Self.

Use `Self [command] --help` for more information about a command.

## Examples

### Prepare ecdsa key pair
```bash
./SelfExtractingUpgrade generateKeys -p /root/aa
ls /root/aa.*
    /root/aa.key  /root/aa.pub
```
### Prepare the package directory
```bash
# pwd
/root/snap
# tree
.
├── docker
│   ├── 2915
│   ├── 2932
│   ├── common
│   └── current -> 2932
└── install.sh # This file is specific and necessary. It cannot have any other name. All decompressed logic should be in it.

```

### Packaging、Signature、Encryption

#### signature
```bash
# pwd
/root/snap
# SelfExtractingUpgrade make -d snap.tgz -s docker -i -k /root/aa.key
```

#### encryption
```bash
# pwd
/root/snap
# SelfExtractingUpgrade make -d snap.tgz -s docker  -c -p 0123456789ABCDEF0123456789ABCDEF
```

#### signature and encryption
```bash 
# pwd
/root/snap
# SelfExtractingUpgrade make -d snap.tgz -s docker -i -k /root/aa.key -c -p 0123456789ABCDEF0123456789ABCDEF
```

tips: When you use encryption, the command will output the password and IV after completion.
tips: If you use signature, you need to send the public key of the key pair generated above to the decompressor.
tips: If you use encryption, you need to send the password to the decompressor.

### Auto unpack
#### signature
```bash 
./snap.tgz.run  /root/aa.pub
```
#### encryption
```bash
./snap.tgz.run 3031323334353637383941424344454630313233343536373839414243444546 4e3050324e314b364135583850384e33
```
#### signature and encryption
```bash
./snap.tgz.run  /root/aa.pub 3031323334353637383941424344454630313233343536373839414243444546 4e3050324e314b364135583850384e33
```

## License

[MIT License](https://kimi.moonshot.cn/chat/LICENSE)

---

Feel free to contribute to the project or report issues!

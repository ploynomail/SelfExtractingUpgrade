package logic

var ScriptTemplate = `#!/bin/bash
#
# Self Extracting Upgrade Package
# Encryption: ECIES (ECDH P-256 + AES-256-CBC) for session key
# Signature: ECDSA P-256 with SHA-256
#
SENDER_PUB='{{.SenderPub}}'
EPHEMERAL_PUB='{{.EphemeralPub}}'
ENCRYPTED_KEY='{{.EncryptedKey}}'
IV='{{.IV}}'
SIGNATURE='{{.Signature}}'

ARCHIVE=$(awk '/^__ARCHIVE_BELOW__/ {print NR + 1; exit 0; }' "$0")
tmp_dir=$(mktemp -d /tmp/seu_XXXXXXXX)
if [ $? -ne 0 ]; then
    echo "ERROR: Failed to create temp directory" >&2
    exit 1
fi
trap 'rm -rf "$tmp_dir"' EXIT INT TERM

if [ $# -ne 1 ]; then
    echo "ERROR: Usage: $0 <recipient-private-key-file>" >&2
    exit 1
fi
recipient_key="$1"

if ! command -v openssl >/dev/null 2>&1; then
    echo "ERROR: openssl is required but not found" >&2
    exit 1
fi

if ! command -v xxd >/dev/null 2>&1; then
    echo "ERROR: xxd is required but not found" >&2
    exit 1
fi

printf '%s\n' "$SENDER_PUB" > "${tmp_dir}/sender.pub"
printf '%s\n' "$EPHEMERAL_PUB" > "${tmp_dir}/ephemeral.pub"

# Extract cipher payload
tail -n+"$ARCHIVE" "$0" > "${tmp_dir}/cipherPayload"

# Verify ECDSA signature over cipher payload
echo "$SIGNATURE" | xxd -r -p > "${tmp_dir}/signature.bin"
if ! openssl dgst -sha256 -verify "${tmp_dir}/sender.pub" -signature "${tmp_dir}/signature.bin" "${tmp_dir}/cipherPayload" >/dev/null 2>&1; then
    echo "ERROR: Signature verification failed" >&2
    exit 1
fi
echo "Signature verification succeeded."

# Decrypt session key via ECIES (ECDH derive + SHA-256 KDF + AES-256-CBC)
if ! openssl pkeyutl -derive -inkey "$recipient_key" -peerkey "${tmp_dir}/ephemeral.pub" -out "${tmp_dir}/shared_secret.bin" >/dev/null 2>&1; then
    echo "ERROR: Failed to derive shared secret" >&2
    exit 1
fi
aes_key=$(openssl dgst -sha256 -binary "${tmp_dir}/shared_secret.bin" | xxd -p | tr -d '\n')
echo "$ENCRYPTED_KEY" | xxd -r -p > "${tmp_dir}/encrypted_key.bin"
if ! openssl aes-256-cbc -d -K "$aes_key" -iv "$IV" -in "${tmp_dir}/encrypted_key.bin" -out "${tmp_dir}/session_key.bin" -nosalt >/dev/null 2>&1; then
    echo "ERROR: Failed to decrypt session key" >&2
    exit 1
fi
session_key=$(xxd -p "${tmp_dir}/session_key.bin" | tr -d '\n')

# Decrypt payload (IV is prepended to cipherPayload)
payload_iv=$(dd if="${tmp_dir}/cipherPayload" bs=1 count=16 2>/dev/null | xxd -p | tr -d '\n')
dd if="${tmp_dir}/cipherPayload" bs=1 skip=16 of="${tmp_dir}/payload.enc" 2>/dev/null
if ! openssl aes-256-cbc -d -K "$session_key" -iv "$payload_iv" -in "${tmp_dir}/payload.enc" -out "${tmp_dir}/payload.tar.gz" -nosalt >/dev/null 2>&1; then
    echo "ERROR: Payload decryption failed" >&2
    exit 1
fi
echo "Decryption succeeded."

# Extract
if ! tar -zxf "${tmp_dir}/payload.tar.gz" -C "$tmp_dir"; then
    echo "ERROR: Extraction failed" >&2
    exit 1
fi

cd "$tmp_dir" || { echo "ERROR: Cannot enter temp directory" >&2; exit 1; }
bash ./install.sh
install_rc=$?
if [ $install_rc -ne 0 ]; then
    echo "ERROR: install.sh exited with code $install_rc" >&2
    exit $install_rc
fi
exit 0

#This line must be the last line of the file
__ARCHIVE_BELOW__
`
